package tsdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/teststruct"
	"github.com/vincoll/vigie/pkg/utils"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb-client-go/v2"
)

type vInfluxDBv2 struct {
	conf ConfInfluxDBv2
}

func (idb *vInfluxDBv2) UpdateTestState(task teststruct.Task) error {
	//panic("implement me")
	return nil
}

func NewInfluxDBv2(conf ConfInfluxDBv2) (*vInfluxDBv2, error) {

	idb := vInfluxDBv2{conf: conf}

	err := idb.validateConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to validate a connection with %q: %s", conf.Addr, err)
	}

	utils.Log.WithFields(logrus.Fields{
		"component": "tsdb",
		"host":      conf.Addr,
		"org":       conf.Organization,
		"bucket":    conf.Bucket,
	}).Infof(fmt.Sprintf("Vigie TSDB to %s connection succeed", idb.Name()))

	return &idb, nil
}

func (idb *vInfluxDBv2) Name() string {
	return "InfluxDBv2"
}

func (idb *vInfluxDBv2) validateConnection() error {

	client := idb.createClient()

	// Ensures background processes finishes
	defer client.Close()

	// Loop to infinity as long as the base has not responded
	success := false
	retryDelay := 500 * time.Millisecond
	for success == false {

		// Get query client
		queryAPI := client.QueryAPI(idb.conf.Organization)
		// Get parser flux query result
		q := fmt.Sprintf(`from(bucket:"%s") |> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`, idb.conf.Bucket)
		_, err := queryAPI.Query(context.Background(), q)
		if err != nil {

			host := strings.Split(idb.conf.Addr, "//")[1]
			_, tcperr := net.Dial("tcp", host)

			if tcperr != nil {

				utils.Log.WithFields(logrus.Fields{
					"component": "tsdb",
					"host":      idb.conf.Addr,
					"org":       idb.conf.Organization,
					"bucket":    idb.conf.Bucket,
				}).Errorf("cannot reach InfluxDB via TCP %s: %s. Next try : %s", host, tcperr, retryDelay)
			} else {

				utils.Log.WithFields(logrus.Fields{
					"component": "tsdb",
					"host":      idb.conf.Addr,
					"org":       idb.conf.Organization,
					"bucket":    idb.conf.Bucket,
				}).Errorf("cannot reach InfluxDB %s: %s. Next try : %s", host, err, retryDelay)

			}
			time.Sleep(retryDelay)
			// Multiplicative wait
			retryDelay = retryDelay * 2

		} else {
			success = true
		}
	}
	return nil
}

func (idb *vInfluxDBv2) createClient() influxdb2.Client {

	client := influxdb2.NewClientWithOptions(idb.conf.Addr, idb.conf.Token,
		influxdb2.DefaultOptions().
			SetUseGZip(true).
			SetTLSConfig(&tls.Config{
				InsecureSkipVerify: true,
			}))

	return client
}

// WritePoint write to InfluxDB ( for now 1 point = 1 request )
// https://github.com/influxdata/influxdb-client-go#writes
// TODO: Implement non-blocking and buffer
func (idb *vInfluxDBv2) WritePoint(task teststruct.Task, vr *teststruct.VigieResult, tags map[string]string) error {

	c := idb.createClient()
	defer c.Close()

	writeAPI := c.WriteAPIBlocking(idb.conf.Organization, idb.conf.Bucket)

	// Point
	task.RLockAll()

	// Push the Step Results to InfluxDB
	utils.Log.WithFields(logrus.Fields{
		"package":   "tsdb_influx",
		"testsuite": task.TestSuite.Name,
		"testcase":  task.TestCase.Name,
		"teststep":  task.TestStep.Name,
	}).Debug("Push task result into ", idb.Name())

	// Create a Time series data aka points for InfluxDB
	// TAGS are used to identify a task in the DB for later queries
	taskTags := tags
	taskTags["testsuite"] = task.TestSuite.Name
	taskTags["testcase"] = task.TestCase.Name
	taskTags["teststep"] = task.TestStep.Name

	var wg sync.WaitGroup
	wg.Add(len(task.TestStep.VigieResults))
	for _, vigieRes := range vr.TestResults {

		vigieRes := vigieRes
		go func(vr teststruct.TestResult) {

			// create data point
			p := influxdb2.NewPoint(
				task.TestStep.ProbeWrap.Probe.GetName(),                     // TODO : Trouver un meuilleur porteur du nommage de "metric"
				utils.MergeTagMaps(taskTags, vigieRes.ProbeReturn.Labels()), // TODO InfluxDB Opti : Ordonner Key A-Z
				vigieRes.ProbeReturn.Values(),
				task.TestStep.LastChange)

			// write synchronously for now
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				panic(err)
			}

			wg.Done()
		}(vigieRes)

	}
	wg.Wait()

	task.RUnlockAll()

	// Ensures background processes finishes
	return nil
}

/*
// writePointToInfluxDB convenience function to insert data into the database
func (idb *vInfluxDBv2) writePointToInfluxDB(np influxdb.Metric) (err error) {

	_, _ = idb.client.Write(context.TODO(), idb.conf.Bucket, idb.conf.Organization)

	return nil
}

// https://v2.docs.influxdata.com/v2.0/write-data/


*/
