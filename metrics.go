package advanced_metrics

import (
	"net/http"
	"strconv"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(AdvancedMetrics{})
	httpcaddyfile.RegisterHandlerDirective("advanced_metrics", parseAdvancedMetrics)
}

type AdvancedMetrics struct {
	prometheusPort  int
	logger          *zap.Logger
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.SummaryVec
	responseSize    *prometheus.SummaryVec
}

func (AdvancedMetrics) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.advanced_metrics",
		New: func() caddy.Module { return new(AdvancedMetrics) },
	}
}

func (AdvancedMetrics) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if !d.NextArg() {
			return d.ArgErr()
		}
	}
	return nil
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

func parseAdvancedMetrics(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	port := getPort(h.Dispenser)
	am := AdvancedMetrics{prometheusPort: port}
	return am, nil
}

func (am *AdvancedMetrics) Provision(ctx caddy.Context) error {
	am.prometheusPort = 6611
	am.logger = ctx.Logger()
	am.logger.Sugar().Infof("Provisioning advanced metrics")
	am.StartServer()
	return nil
}

func (am AdvancedMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	return am.HandleRequest(w, r, next)
}

// Interface guards - static test that interfaces are compliant
var (
	_ caddy.Provisioner           = (*AdvancedMetrics)(nil)
	_ caddyfile.Unmarshaler       = (*AdvancedMetrics)(nil)
	_ caddyhttp.MiddlewareHandler = (*AdvancedMetrics)(nil)
)
