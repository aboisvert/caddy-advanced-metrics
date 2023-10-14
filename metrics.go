package advanced_metrics

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	module  *AdvancedMetricsModule
	counter bool
	latency bool
}

func (AdvancedMetricsHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.advanced_metrics",
		New: func() caddy.Module { return new(AdvancedMetricsHandler) },
	}
}

func getOption(d *caddyfile.Dispenser, name string) bool {
	exist := false
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case name:
				exist = true
			}
		}
	}
	return exist
}

func getPort(d *caddyfile.Dispenser) int {
	var port int = 6611
	for d.Next() {
		for d.NextBlock(0) {
			switch d.Val() {
			case "port":
				if d.NextArg() {
					i, err := strconv.Atoi(d.Val())
					if err == nil {
						port = i
					}
				}
			}
		}
	}
	return port
}

func (am *AdvancedMetricsHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	fmt.Printf("AdvancedMetricsHandler.UnmarshalCaddyfile\n")
	for d.Next() {
		// if !d.Args(&m.Output) {
		// 	return d.ArgErr()
		// }
	}
	return nil
}

func parseAdvancedMetricsHandler(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	am := AdvancedMetricsHandler{module: &MODULE}
	am.counter = getOption(h.Dispenser, "counter")
	am.latency = getOption(h.Dispenser, "latency")
	if DEBUG {
		fmt.Printf("parseAdvancedMetrics counter=%t\n", am.counter)
		fmt.Printf("parseAdvancedMetrics latency=%t\n", am.latency)
	}

	m := am.module
	m.mutex.Lock()
	if !m.started {
		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		m.logger = zap.New(
			zapcore.NewCore(zapcore.NewConsoleEncoder(developmentCfg), zapcore.AddSync(os.Stdout), zap.InfoLevel))
		m.logger.Sugar().Infof("Starting advanced metrics server\n")
		m.prometheusPort = getPort(h.Dispenser)
		m.StartServer()
	} else {
		m.logger.Sugar().Infof("Already started")
	}
	m.mutex.Unlock()

	return am, nil
}

func (am AdvancedMetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m := am.module
	if m == nil {
		fmt.Printf("am.module is nil")
	}
	fmt.Printf("ServeHTTP %s %s counter=%t latency=%t", r.Host, r.URL.Path, am.counter, am.latency)
	// m.logger.Sugar().Infof("ServeHTTP %s %s counter=%t latency=%t", r.Host, r.URL.Path, am.counter, am.latency)
	return am.HandleRequest(w, r, next)
}

// Interface guards - static test that interfaces are compliant
var (
	_ caddyhttp.MiddlewareHandler = (*AdvancedMetricsHandler)(nil)
	_ caddyfile.Unmarshaler       = (*AdvancedMetricsHandler)(nil)
)

var MODULE AdvancedMetricsModule = AdvancedMetricsModule{}
var DEBUG bool = true
