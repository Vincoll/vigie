package worker

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func loadWorkerConfigFile(confpath string) (vc WorkerConf, err error) {

	// Set default path if no custom path is provided
	if confpath == "" {
		confpath = defaultConfFilePath()
		fmt.Println("Load default webapi conf:", confpath)
	}

	if _, err := os.Stat(confpath); os.IsNotExist(err) {
		return WorkerConf{}, fmt.Errorf(" %s ", err)
	}

	// Viper for Unmarshall toml webapi config file
	vpr := viper.New()
	vpr.SetConfigFile(confpath)
	if err := vpr.ReadInConfig(); err != nil {
		return WorkerConf{}, fmt.Errorf("couldn't load config: %s", err)
	}

	if err := vpr.Unmarshal(&vc); err != nil {
		fmt.Printf("Couldn't read config: %s", err)
	}

	applyEnvironment(&vc)

	return vc, nil
}

func applyEnvironment(vc *WorkerConf) string {

	switch vc.Environment {

	case "dev", "develop", "development":
		vc.Environment = "development"
	default:
		vc.Environment = "production"

	}
	// Apply Environment on app parts
	vc.Prometheus.Environment = vc.Environment
	vc.Log.Environment = vc.Environment

	return vc.Environment
}
