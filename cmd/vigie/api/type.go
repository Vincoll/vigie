package vigieapi

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vincoll/vigie/foundation/logg"
	"github.com/vincoll/vigie/internal/api/dbsqlx"
	"github.com/vincoll/vigie/internal/api/webapi"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/promexporter"
)

// https://xuri.me/toml-to-go/

const defaultConfFile = "config/webapi.toml"

type VigieConf struct {
	ApiVersion  float32
	Environment string // Production, Dev
	Import      load.ConfImport
	HTTP        webapi.APIServerConfig
	PG          dbsqlx.PGConfig
	Prometheus  promexporter.ConfPrometheus
	Log         logg.LogConf
}

func defaultConfFilePath() string {

	path, err := os.Getwd()
	if err != nil {
		//	log.Printf(err)
	}

	pathToConf := filepath.Clean(fmt.Sprintf("%s/%s", path, defaultConfFile))

	return pathToConf
}
