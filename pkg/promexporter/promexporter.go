package promexporter

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"github.com/vincoll/vigie/pkg/utils"
	"github.com/vincoll/vigie/pkg/vigie"
)

var exporter *exporterVigie

const namespace = "vigie"

var hostname, _ = os.Hostname()

type exporterVigie struct {
	vigieInstance *vigie.Vigie
	mutex         sync.RWMutex

	up     *prometheus.Desc
	uptime *prometheus.Desc
	// TestSuite
	statusTestSuite *prometheus.GaugeVec
	// testCase
	statusTestCase *prometheus.GaugeVec
	// TestStep Desc because need for MustNewConstMetric
	statusTestStep *prometheus.Desc
	// Conservé pour belle liste de Statut grafana.. Mais non efficient
	statusTestStep2 *prometheus.GaugeVec
}

func InitPromExporter(confProm ConfPrometheus, vigieInstance *vigie.Vigie) {

	if confProm.Enable == true {
		exporter = newExporter(vigieInstance)

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

	if env == "prod" || env == "production" {

		r := prometheus.NewRegistry()
		r.MustRegister(exporter)
		handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})
		http.Handle("/metrics", handler)
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}
	if env == "dev" || env == "development" {

		// Add Go Runtime Metrics
		//Register metrics with prometheus

		//prometheus.MustRegister(exporter)

		//This section will start the HTTP server and expose
		//any metrics on the /metrics endpoint.
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprint(":", port), nil)
	}
}
