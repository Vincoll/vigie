package server

import (
	"github.com/vincoll/vigie/pkg/alertmanager"
	"github.com/vincoll/vigie/pkg/tsdb"
	"github.com/vincoll/vigie/pkg/vigie"
)

//  Ideas for 1.0
type Server struct {
	vigie        *vigie.Vigie
	alertmanager *alertmanager.AlertManager
	tsdb         *tsdb.ConfInfluxDB
	//router

}
