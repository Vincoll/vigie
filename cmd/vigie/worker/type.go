package worker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vincoll/vigie/internal/worker/pulsar_worker"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/vigie"

	"github.com/vincoll/vigie/pkg/promexporter"
	"github.com/vincoll/vigie/pkg/utils"
)

// https://xuri.me/toml-to-go/

const defaultConfFile = "config/workerconfig.toml"

type WorkerConf struct {
	ApiVersion  float32
	Environment string // Production, Dev
	Host        vigie.HostInfo
	Import      load.ConfImport
	Pulsar      pulsar_worker.ConfPulsar
	Prometheus  promexporter.ConfPrometheus

	Log utils.LogConf
}

func defaultConfFilePath() string {

	path, err := os.Getwd()
	if err != nil {
		//	log.Printf(err)
	}

	pathToConf := filepath.Clean(fmt.Sprintf("%s/%s", path, defaultConfFile))

	return pathToConf
}
