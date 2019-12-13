package promexporter

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/vincoll/vigie/pkg/utils"
)

func InitPromExporter(confProm ConfPrometheus) {

	if confProm.Enable == true {

		utils.Log.WithFields(logrus.Fields{
			"component": "prometheusexporter",
			"port":      confProm.Port,
			"gometrics": confProm.Gometrics,
			"env":       &confProm.Environment,
		}).Infof(fmt.Sprintln("Vigie Prom exporter is exposed"))

		run(confProm.Port, confProm.Environment)

	}
}

func run(port int, env string) {

	if env == "dev" || env == "development" {

		// Add Go Runtime Metrics
		//This section will start the HTTP server and expose
		//any metrics on the /metrics endpoint.
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}
}
