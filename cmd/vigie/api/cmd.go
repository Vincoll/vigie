package vigieapi

import (
	"fmt"
	"os"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
	"github.com/vincoll/vigie/cmd/vigie/version"
	"github.com/vincoll/vigie/foundation/logg"
	"github.com/vincoll/vigie/internal/api/core"
	"github.com/vincoll/vigie/pkg/vigie"
)

var (
	configfile    string
	vigieInstance *vigie.Vigie
	variables     []string
	withEnv       bool
)

func init() {
	Cmd.Flags().StringVar(&configfile, "config", "", "--config ./webapi.toml")
	Cmd.Flags().BoolVarP(&withEnv, "env", "", false, "Inject environment variables. export FOO=BAR -> you can use {{.FOO}} in your tests")
}

// Cmd run
var Cmd = &cobra.Command{
	Use:     "api",
	Example: "api --config ./config/webapi.toml",
	Short:   "Run Vigie API",
	PreRun: func(cmd *cobra.Command, args []string) {
		//
		// Create Vigie Instance
		//
		var err error

		vigieInstance, err = vigie.NewVigie()
		if err != nil {

			logger, err := logg.NewLogger("vigie", "env", "debug")
			if err != nil {
				fmt.Printf("Error while creating logger: %s", err)
				os.Exit(1)
			}
			logger.Errorw("Vigie API",
				"component", "vigie",
				"status", "error",
				"error", err.Error(),
			)
			os.Exit(1)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {

		const serviceName = "vigie-api"

		vigieConf := loadVigieConfigFile(configfile)

		//
		// Create LOGGER
		//
		logger, err := logg.NewLogger(serviceName, vigieConf.Env, vigieConf.Log.Level)
		if err != nil {
			fmt.Printf("Error while creating logger: %s", err)
			os.Exit(1)
		}
		//
		// Vigie API Info
		//
		logger.Infow("Vigie API",
			"version", version.LdVersion,
			"goruntime", runtime.Version(),
			"arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"builddate", version.LdBuildDate,
			"configfile", configfile,
		)

		//
		// Check ConfigFile
		//
		_, err = govalidator.ValidateStruct(vigieConf)
		if err != nil {
			logger.Errorw("Vigie API",
				"component", "configfile",
				"status", "invalid",
				"error", "Vigie Config File is invalid",
			)
			os.Exit(1)
		}

		// Merge Config - Not the ideal place ...
		vigieConf.OTel.Env = vigieConf.Env
		vigieConf.OTel.ServiceName = serviceName
		vigieConf.OTel.Version = version.LdVersion

		//
		// Start Vigie Instance
		//

		err = core.NewVigieAPI(vigieConf, logger)
		if err != nil {

			logger.Fatalw("Vigie API failed to start",
				"err", err,
				"version", version.LdVersion,
				"goruntime", runtime.Version(),
				"arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
				"builddate", version.LdBuildDate,
				"configfile", configfile,
			)

			os.Exit(1)
		}

	},
}
