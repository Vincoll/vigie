package vigiescheduler

import (
	"fmt"
	"os"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vincoll/vigie/cmd/vigie/version"
	"github.com/vincoll/vigie/foundation/logg"
	"github.com/vincoll/vigie/internal/scheduler/core"
	"github.com/vincoll/vigie/pkg/utils"
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
	Use:     "scheduler",
	Example: "scheduler --config ./config/webapi.toml",
	Short:   "Start Scheduler",
	PreRun: func(cmd *cobra.Command, args []string) {
		//
		// Create Vigie Instance
		//
		var err error

		vigieInstance, err = vigie.NewVigie()
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"component": "scheduler", "status": "error", "error": "Vigie cannot start"}).Fatal(err)
			os.Exit(1)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {

		vigieConf := loadVigieConfigFile(configfile)

		//
		// Create LOGGER
		//
		logger, err := logg.NewLogger("vigie-scheduler", "dev", "debug")
		if err != nil {
			os.Exit(1)
		}

		//
		// Vigie Scheduler Info
		//
		logger.Infow("Vigie Scheduler",
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
			logger.Errorw("Vigie Scheduler",
				"component", "configfile",
				"status", "invalid",
				"error", "Vigie Config File is invalid",
			)
			os.Exit(1)
		}

		// Merge Config - Not the ideal place ...
		vigieConf.OTel.Env = vigieConf.Environment
		vigieConf.OTel.ServiceName = "vigie-scheduler"
		vigieConf.OTel.Version = "0.0.22"

		//
		// Start Vigie Instance
		//
		err = core.NewVigieScheduler(vigieConf, logger)
		if err != nil {

			logger.Fatalw("Vigie Scheduler failed to start",
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
