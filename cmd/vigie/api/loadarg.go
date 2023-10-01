package vigieapi

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"github.com/vincoll/vigie/internal/api/conf"
)

func loadVigieConfigFile(confpath string) (vc conf.VigieAPIConf) {

	// Set default path if no custom path is provided
	if confpath == "" {
		confpath = conf.DefaultConfFile
		fmt.Println("Load default webapi conf:", confpath)
	}

	if _, err := os.Stat(confpath); os.IsNotExist(err) {
		fmt.Println("File do not exist:", confpath, err)
		os.Exit(1)
	}

	// Viper for Unmarshall toml webapi config file
	vpr := viper.New()
	vpr.SetConfigFile(confpath)
	if err := vpr.ReadInConfig(); err != nil {
		fmt.Println("Couldn't load config:", err)
		os.Exit(1)
	}

	if err := vpr.Unmarshal(&vc); err != nil {
		fmt.Printf("Couldn't read config: %s", err)
	}

	applyEnvironment(&vc)

	return vc
}

// TODO:AddOSEnvironmentVariables
// Add Variables System Environment Variables
func addOSEnvironmentVariables() (mapvars map[string]string) {

	withEnv := false
	variables := make([]string, 0) // Wrong
	if withEnv {
		variables = append(variables, os.Environ()...)
	}

	for _, a := range variables {
		t := strings.SplitN(a, "=", 2)
		if len(t) < 2 {
			continue
		}
		(mapvars)[t[0]] = strings.Join(t[1:], "")
	}
	return mapvars
}

func applyEnvironment(vc *conf.VigieAPIConf) string {

	switch vc.Env {

	case "dev", "develop", "development":
		vc.Env = "development"
	default:
		vc.Env = "production"

	}
	return vc.Env
}
