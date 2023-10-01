package conf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vincoll/vigie/foundation/logg"
	"github.com/vincoll/vigie/internal/api/dbpgx"
	"github.com/vincoll/vigie/internal/api/webapi"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/promexporter"
	"github.com/vincoll/vigie/pkg/tracing"
)

// https://xuri.me/toml-to-go/

const DefaultConfFile = "config/webapi.toml"

type VigieAPIConf struct {
	ApiVersion float32
	Env        string // Production, Dev
	Import     load.ConfImport
	HTTP       webapi.APIServerConfig
	PG         dbpgx.PGConfig
	Prometheus promexporter.ConfPrometheus
	Log        logg.LogConf
	OTel       tracing.OTelConfig
}

func defaultConfFilePath() string {

	path, err := os.Getwd()
	if err != nil {
		//	log.Printf(err)
	}

	pathToConf := filepath.Clean(fmt.Sprintf("%s/%s", path, DefaultConfFile))

	return pathToConf
}
