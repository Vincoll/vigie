package promexporter

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/vincoll/vigie/pkg/vigie"
)

func newExporter(v *vigie.Vigie) *exporterVigie {

	var e exporterVigie

	e.vigieInstance = v

	e.up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Could the apache server be reached",
		nil,
		nil)

	e.statusTestSuite = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vigie_testsuite_state",
			Help: "Testsuite ResultStatus",
		}, []string{"testsuite"},
	)

	e.statusTestCase = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vigie_testcase_state",
			Help: "Testcase ResultStatus",
		}, []string{"testsuite", "testcase"},
	)

	e.statusTestStep2 = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "vigie_teststep_state2",
			Help: "Teststep Status2",
		}, []string{"testsuite", "testcase", "teststep"},
	)

	// NewDesc is used in order to use
	e.statusTestStep = prometheus.NewDesc(
		// Name
		"vigie_teststep_state",
		// Help text
		"Teststep ResultStatus",
		// Label Var
		[]string{"testsuite", "testcase", "teststep"},
		// Label Const
		nil)

	e.uptime = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "uptime_seconds_total"),
		"Current uptime in seconds (*)",
		nil,
		nil)

	return &e
}

func (e *exporterVigie) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	ch <- e.uptime

	e.statusTestSuite.Describe(ch)
}

func (e *exporterVigie) collect(ch chan<- prometheus.Metric) error {
	/*
		fmt.Println("Prom Pull", time.Now())

		// Vigie Instance Browsable OK
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

		for idTS, _ := range e.vigieInstance.TestSuites {
			ts := &e.vigieInstance.TestSuites[idTS]
			e.statusTestSuite.WithLabelValues(ts.Name).Set(1)
			for idTC, _ := range ts.JsonTestCases {
				tc := &ts.JsonTestCases[idTC]
				e.statusTestCase.WithLabelValues(ts.Name, tc.Name).Set(b2f64(tc.ResultStatus))
				for i, _ := range tc.TestSteps {
					tstp := &tc.TestSteps[i]
					if !false {

							e.statusTestStep2.WithLabelValues(ts.Name, tc.Name, tstp.Name).Set(tstp.ResponseTime3())

							ch <- prometheus.MustNewConstMetric(
								e.statusTestStep,
								prometheus.GaugeValue,
								tstp.ResponseTime3(),
								ts.Name, tc.Name, tstp.Name)

					}

				}
			}
		}

		e.statusTestSuite.Collect(ch)
		e.statusTestCase.Collect(ch)
		//	e.statusTestStep2.Collect(ch)


	*/
	return nil
}

func (e *exporterVigie) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock() // To protect metrics from concurrent collects.
	defer e.mutex.Unlock()
	if err := e.collect(ch); err != nil {

	}
	return
}

func b2f64(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
