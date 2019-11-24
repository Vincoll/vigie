package tsdb

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/sirupsen/logrus"
	"github.com/vincoll/vigie/pkg/utils"
)

// https://github.com/kubernetes/test-infra/blob/master/velodrome/transform/influx.go

type ConfInfluxDB struct {
	Enable   bool   `toml:"enable"`
	Addr     string `toml:"addr"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type vInfluxDB struct {
	conf ConfInfluxDB
}

var InfluxInst vInfluxDB

func LoadInfluxDB(c ConfInfluxDB) error {

	if c.Enable == true {

		InfluxInst = vInfluxDB{
			conf: c,
		}
		// Test if vInfluxDB is accessible
		err := InfluxInst.validateConnection()
		if err != nil {
			return fmt.Errorf("failed to validate a connection with %q: %s", c.Addr, err)
		}
	}
	utils.Log.WithFields(logrus.Fields{
		"component": "tsdb",
		"host":      c.Addr,
		"db":        c.Database,
	}).Infof(fmt.Sprintln("Vigie TSDB connection succeed"))
	return nil
}

func (idb *vInfluxDB) createClient() (client.Client, error) {

	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     idb.conf.Addr,
		Username: idb.conf.User,
		Password: idb.conf.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating vInfluxDB Client: %s", err.Error())
	}
	return c, nil
}

// queryDB convenience function to query the database
func (idb *vInfluxDB) QueryDB(cmd string) (res []client.Result, err error) {

	clnt, err := idb.createClient()
	if err != nil {
		return nil, err
	}
	defer clnt.Close()

	q := client.Query{
		Command:  cmd,
		Database: idb.conf.Database,
	}

	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, fmt.Errorf("error during vInfluxDB Query: %s", err.Error())
	}
	return res, nil
}

func (idb *vInfluxDB) validateConnection() error {

	success := false
	var res []client.Result
	// Loop to infinity as long as the base has not responded
	for success == false {

		r, err := idb.QueryDB("SHOW DATABASES")
		if err != nil {

			host := strings.Split(idb.conf.Addr, "//")[1]
			_, tcperr := net.Dial("tcp", host)

			if tcperr != nil {

				utils.Log.WithFields(logrus.Fields{
					"component": "tsdb",
					"host":      idb.conf.Addr,
					"db":        idb.conf.Database,
				}).Errorf("cannot reach InfluxDB via TCP %s: %s. Next try : 5sec", host, tcperr)

			} else {
				utils.Log.WithFields(logrus.Fields{
					"component": "tsdb",
					"host":      idb.conf.Addr,
					"db":        idb.conf.Database,
				}).Errorf("cannot reach InfluxDB %s: %s. Next try : 5sec", host, err)
			}
			time.Sleep(5 * time.Second)

		} else {
			res = r
			success = true
		}
	}

	// Check if the database name provided in vigie.toml exists in this influxdb instance
	for _, row := range res[0].Series[0].Values {
		str, _ := row[0].(string)
		if str == idb.conf.Database {
			return nil
		}
	}

	return fmt.Errorf("vInfluxDB user: %q can not access DB: %q", idb.conf.User, idb.conf.Database)
}

func (idb *vInfluxDB) batchPoint(retPol string) (client.BatchPoints, error) {

	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        idb.conf.Database,
		Precision:       "ms",
		RetentionPolicy: retPol,
	})
	if err != nil {
		return nil, err
	}
	return bp, nil
}

// WritePoint convenience function to insert data into the database
func (idb *vInfluxDB) WritePoint(np *client.Point, retPol string) (err error) {

	clnt, _ := idb.createClient()
	defer clnt.Close()
	bp, _ := idb.batchPoint(retPol)

	bp.AddPoint(np)
	if err := clnt.Write(bp); err != nil {
		return err
	}

	return nil
}

func (idb *vInfluxDB) IsEnable() bool {
	return idb.conf.Enable
}
