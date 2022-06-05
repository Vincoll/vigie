package promexporter

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func InitPromExporter(confProm ConfPrometheus) {

	if confProm.Enable == true || confProm.Environment == "dev" || confProm.Environment == "development" {

		zap.S().Infow("(Vigie Prom exporter is exposed",
			zap.String("component", "prometheusexporter"),
			zap.Int("port", confProm.Port),
			zap.String("env", confProm.Environment),
		)

		run(confProm.Port)

	}
}

func run(port int) {

	// Add Go Runtime Metrics
	// This section will start the HTTP server and expose
	// any metrics on the /metrics endpoint.
	http.Handle("/metrics", promhttp.Handler())
	_ = http.ListenAndServe(fmt.Sprint(":", port), nil)
}
