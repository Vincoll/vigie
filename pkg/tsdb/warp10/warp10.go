package warp10

import (
	"github.com/vincoll/vigie/pkg/tsdb"
)

// https://github.com/miton18/go-warp10/tree/master/base
// https://github.com/PierreZ/Warp10Exporter

type warp10 struct {
	conf tsdb.ConfWarp10
}

var WarpInst warp10

func (w *warp10) validateConnection() error {

	return nil
}

func (w *warp10) NewGTS(metricName string, labels map[string]string, datapoints map[string]string) error {

	// teststep{testsuite=TS01, testcase=TC01, teststep=TSTEP01}
	// gov.noaa.storm.wind{serial=2015066S08170,.app=tuto}{name=PAM} 50.0
	// https://www.warp10.io/content/03_Documentation/03_Interacting_with_Warp_10/03_Ingesting_data/02_GTS_input_format

	return nil
}
