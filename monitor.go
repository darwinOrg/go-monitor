package monitor

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	dgcoll "github.com/darwinOrg/go-common/collection"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	counterMap        = new(sync.Map)
	name2LabelKeysMap = new(sync.Map)
)

var (
	ServerReqCounter  *prometheus.CounterVec
	ServerReqDuration *prometheus.HistogramVec
	ServerReqInFlight *prometheus.GaugeVec
)

var (
	ClientReqCounter  *prometheus.CounterVec
	ClientReqDuration *prometheus.HistogramVec
	ClientReqInFlight *prometheus.GaugeVec
)

func Start(appName string, port int) {
	ServerReqCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_http_requests_total",
		Help: "Total number of HTTP requests received by the server",
		ConstLabels: map[string]string{
			"app_name": appName,
		},
	}, []string{"path"})

	ServerReqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "server_http_request_duration_seconds",
		Help: "Duration of HTTP requests received by the server",
	}, []string{"path", "status"})

	ServerReqInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "server_http_requests_in_flight",
		Help: "Number of HTTP requests currently being served",
	}, []string{"path"})

	ClientReqCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_http_requests_total",
		Help: "Total number of HTTP requests sent by the client",
	}, []string{"url"})

	ClientReqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "client_http_request_duration_seconds",
		Help: "Duration of HTTP requests sent by the client",
	}, []string{"url", "status"})

	ClientReqInFlight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "client_http_requests_in_flight",
		Help: "Number of HTTP requests currently being processed by the client",
	}, []string{"url"})

	ListenAndServe(port)
}

// ListenAndServe 启动监控服务，提供 Prometheus 指标接口
// port: 监听端口号
func ListenAndServe(port int) {
	go func() {
		http.Handle("/monitor/prometheus", promhttp.Handler())
		host := fmt.Sprintf(":%d", port)
		log.Printf("start to monitor:%d....\n", port)
		_ = http.ListenAndServe(host, nil)
	}()
}

// HttpClientCounter 记录客户端HTTP请求次数
// url: 请求的URL地址
func HttpClientCounter(url string) {
	ClientReqCounter.WithLabelValues(url).Inc()
}

// HttpClientDuration 记录客户端HTTP请求耗时
// url: 请求的URL地址
// status: 请求状态码或错误信息
// cost: 请求耗时(毫秒)
func HttpClientDuration(url string, status string, cost int64) {
	ClientReqDuration.WithLabelValues(url, status).Observe(float64(cost) / 1000) // 转换为秒
}

// HttpServerCounter 记录服务端HTTP请求次数
// url: 请求路径
func HttpServerCounter(url string) {
	ServerReqCounter.WithLabelValues(url).Inc()
}

// HttpServerDuration 记录服务端HTTP请求耗时
// url: 请求路径
// status: 响应状态码或错误信息
// cost: 请求处理耗时(毫秒)
func HttpServerDuration(url string, status string, cost int64) {
	ServerReqDuration.WithLabelValues(url, status).Observe(float64(cost) / 1000) // 转换为秒
}

// HttpClientInFlightIncrement 增加客户端正在处理的请求数量
func HttpClientInFlightIncrement(url string) {
	ClientReqInFlight.WithLabelValues(url).Inc()
}

// HttpClientInFlightDecrement 减少客户端正在处理的请求数量
func HttpClientInFlightDecrement(url string) {
	ClientReqInFlight.WithLabelValues(url).Dec()
}

// HttpServerInFlightIncrement 增加服务端正在处理的请求数量
func HttpServerInFlightIncrement(path string) {
	ServerReqInFlight.WithLabelValues(path).Inc()
}

// HttpServerInFlightDecrement 减少服务端正在处理的请求数量
func HttpServerInFlightDecrement(path string) {
	ServerReqInFlight.WithLabelValues(path).Dec()
}

// IncCounter 增加自定义计数器指标
// name: 指标名称
// labelMap: 标签键值对
// 返回错误信息，如果参数不合法则返回错误
func IncCounter(name string, labelMap map[string]string) error {
	if name == "" || len(labelMap) == 0 {
		return dgerr.ARGUMENT_NOT_VALID
	}

	counter, _ := counterMap.Load(name)
	if counter == nil {
		labelKeys := make([]string, 0, len(labelMap))
		for labelKey, _ := range labelMap {
			labelKeys = append(labelKeys, labelKey)
		}
		dgcoll.SimpleSortAsc(labelKeys)

		counter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: name,
		}, labelKeys)

		name2LabelKeysMap.Store(name, labelKeys)
		counterMap.Store(name, counter)
	}

	labelKeys, _ := name2LabelKeysMap.Load(name)
	keys := labelKeys.([]string)
	labelValues := make([]string, 0, len(labelMap))
	for _, labelKey := range keys {
		labelValues = append(labelValues, labelMap[labelKey])
	}

	counter.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc()
	return nil
}
