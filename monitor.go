package monitor

import (
	"errors"
	"fmt"
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
		return errors.New("invalid params")
	}

	counter, _ := counterMap.Load(name)
	if counter == nil {
		labelKeys := make([]string, 0, len(labelMap))
		for labelKey := range labelMap {
			labelKeys = append(labelKeys, labelKey)
		}

		counter = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: name,
			Help: name,
		}, labelKeys)

		counterMap.Store(name, counter)
	}

	labelValues := make([]string, 0, len(labelMap))
	for _, labelValue := range labelMap {
		labelValues = append(labelValues, labelValue)
	}

	counter.(*prometheus.CounterVec).WithLabelValues(labelValues...).Inc()
	return nil
}
