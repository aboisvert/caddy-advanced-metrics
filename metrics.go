package advanced_metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AdvancedMetricsHandler{module: &MODULE})
	httpcaddyfile.RegisterHandlerDirective("advanced_metrics", parseAdvancedMetricsHandler)
}

type AdvancedMetricsModule struct {
	mutex           sync.RWMutex
	started         bool
	prometheusPort  int
	logger          *zap.Logger
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.SummaryVec
	responseSize    *prometheus.SummaryVec
}

type AdvancedMetricsHandler struct {
	PrometheusPort int  `json:"port,omitempty"`
	Counter        bool `json:"counter,omitempty"`
	Latency        bool `json:"latency,omitempty"`
	module         *AdvancedMetricsModule
}

func (AdvancedMetricsHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.advanced_metrics",
		New: func() caddy.Module {
			return new(AdvancedMetricsHandler)
		},
	}
}

func parseAdvancedMetricsHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	am := AdvancedMetricsHandler{}

	d := h.Dispenser
	am.PrometheusPort = 6611
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "port":
				if d.NextArg() {
					i, err := strconv.Atoi(d.Val())
					if err == nil {
						am.PrometheusPort = i
					}
				}
			case "counter":
				if d.NextArg() {
					am.Counter = (d.Val() == "true")
				}
			case "latency":
				if d.NextArg() {
					am.Latency = (d.Val() == "true")
				}
			}

		}
	}

	if DEBUG {
		fmt.Printf("parseAdvancedMetrics port=%d\n", am.PrometheusPort)
		fmt.Printf("parseAdvancedMetrics counter=%t\n", am.Counter)
		fmt.Printf("parseAdvancedMetrics latency=%t\n", am.Latency)
	}
	return &am, nil
}

func (am *AdvancedMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m := am.module
	if m == nil {
		fmt.Printf("am.module is nil\n")
	}
	if DEBUG {
		fmt.Printf("ServeHTTP %s %s counter=%t latency=%t\n", r.Host, r.URL.Path, am.Counter, am.Latency)
	}
	return am.HandleRequest(w, r, next)
}

func (am *AdvancedMetricsHandler) Provision(ctx caddy.Context) error {
	if DEBUG {
		fmt.Printf("Provision counter=%t\n", am.Counter)
		fmt.Printf("Provision latency=%t\n", am.Latency)
	}

	am.module = &MODULE
	m := am.module

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.started {
		m.logger = ctx.Logger()
		m.logger.Sugar().Infof("Starting advanced metrics server\n")
		m.StartServer()
	} else {
		m.logger.Sugar().Infof("Already started")
	}

	return nil
}

// Interface guards - static test that interfaces are compliant
var (
	_ caddyhttp.MiddlewareHandler = (*AdvancedMetricsHandler)(nil)
	_ caddy.Provisioner           = (*AdvancedMetricsHandler)(nil)
)

var MODULE AdvancedMetricsModule = AdvancedMetricsModule{}

const DEBUG bool = false
