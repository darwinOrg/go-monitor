package monitor

import (
	"fmt"
	dgcoll "github.com/darwinOrg/go-common/collection"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"sync"
)

var ServerReqCounter *prometheus.CounterVec
var ServerReqGauge *prometheus.GaugeVec
var ServerReqDuration *prometheus.HistogramVec

var ClientReqCounter *prometheus.CounterVec
var ClientReqGauge *prometheus.GaugeVec
var ClientReqDuration *prometheus.HistogramVec

var counterMap = new(sync.Map)
var name2LabelKeysMap = new(sync.Map)

func Start(appName string, port int) {
	ServerReqCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "server_http_count",
		Help: "server http counter",
		ConstLabels: map[string]string{
			"appName": appName,
		},
	}, []string{"path"})
	ServerReqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_response_cost_time",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "error"})
	ServerReqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_response_time_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "error"})

	ClientReqCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "client_http_count",
		Help: "client http counter",
	}, []string{"path"})
	ClientReqGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_client_response_cost_time",
		Help: "Duration of HTTP client requests.",
	}, []string{"path", "error"})
	ClientReqDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_client_response_time_seconds",
		Help: "Duration of HTTP client requests.",
	}, []string{"path", "error"})

	ListenAndServe(port)
}

func ListenAndServe(port int) {
	go func() {
		http.Handle("/monitor/prometheus", promhttp.Handler())
		host := fmt.Sprintf(":%d", port)
		log.Printf("start to monitor:%d....\n", port)
		_ = http.ListenAndServe(host, nil)
	}()
}

func HttpClientCounter(url string) {
	ClientReqCounter.WithLabelValues(url).Inc()
}

func HttpClientDuration(url string, e string, cost int64) {
	ClientReqDuration.WithLabelValues(url, e).Observe(float64(cost))
	ClientReqGauge.WithLabelValues(url, e).Set(float64(cost))
}

func HttpServerCounter(url string) {
	ServerReqCounter.WithLabelValues(url).Inc()
}

func HttpServerDuration(url string, e string, cost int64) {
	ServerReqDuration.WithLabelValues(url, e).Observe(float64(cost))
	ServerReqGauge.WithLabelValues(url, e).Set(float64(cost))
}

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
