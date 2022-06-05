package worker

import (
	"fmt"
	"os"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/vincoll/vigie/internal/worker/pulsar_worker"
	"github.com/vincoll/vigie/internal/worker/worker"
	"github.com/vincoll/vigie/pkg/promexporter"
	"go.uber.org/zap"

	"github.com/spf13/cobra"
	"github.com/vincoll/vigie/cmd/vigie/version"
	"github.com/vincoll/vigie/internal/worker/utils"
)

var (
	configfile     string
	workerInstance *worker.Worker
	withEnv        bool
)

func init() {
	Cmd.Flags().StringVar(&configfile, "config", "", "--config ./worker.toml")
	Cmd.Flags().BoolVarP(&withEnv, "env", "", false, "Inject environment variables. export FOO=BAR -> you can use {{.FOO}} in your tests")
}

// Cmd worker
var Cmd = &cobra.Command{
	Use:     "worker",
	Example: "worker run --config ./config.toml",
	Short:   "Run the Worker",
	PreRun: func(cmd *cobra.Command, args []string) {

		workerInstance = worker.NewWorker()

	},

	Run: func(cmd *cobra.Command, args []string) {

		//
		// Init Zap GLOBAL LOGGER
		//
		utils.NewLogger("worker", "", "")

		//
		// Load Config
		//
		workerConf, err := loadWorkerConfigFile(configfile)
		if err != nil {
			zap.S().Fatalw("Worker cannot start: "+err.Error(),
				// Structured context as strongly typed Field values.
				zap.String("version", version.LdVersion),
				zap.String("goruntime", runtime.Version()),
				zap.String("arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
				zap.String("builddate", version.LdBuildDate),
				zap.String("configfile", configfile))
			zap.String("component", "worker")
			os.Exit(1)
		}

		zap.S().Infow("Worker is starting",
			// Structured context as strongly typed Field values.
			zap.String("version", version.LdVersion),
			zap.String("goruntime", runtime.Version()),
			zap.String("arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
			zap.String("builddate", version.LdBuildDate),
			zap.String("configfile", configfile))
		zap.String("component", "worker")

		//
		// Init Zap GLOBAL LOGGER with Configuration File
		//
		_, err = utils.NewLogger("worker", workerConf.Log.Environment, workerConf.Log.Level)
		if err != nil {
			return
		}

		//
		// Check ConfigFile
		//
		_, err = govalidator.ValidateStruct(workerConf)
		if err != nil {
			zap.S().Fatalf("Vigie Config File is invalid: %s", err)
		}

		//
		// Start tooling before Vigie instance
		//

		// Start ConfWebAPI
		if workerConf.Pulsar.Enable {
			_, errPulsar := pulsar_worker.NewClient(workerConf.Pulsar)
			if errPulsar != nil {
				zap.S().Fatalf("Could not instantiate Pulsar client: %s ", err)
				os.Exit(2)
			}
		}

		if workerConf.Prometheus.Enable {
			go promexporter.InitPromExporter(workerConf.Prometheus)
		}

		// DNSCACHED
		workerInstance.CacheDNS = utils.NewCachedResolver()

		// Add info about host
		workerConf.Host.AddHostSytemInfo()
		// workerInstance.HostInfo = workerConf.Host

		//
		// Start Vigie Instance
		//

		err = workerInstance.Start()
		if err != nil {
			zap.S().Fatalf("Vigie Worker failed to launch: ", err)
			os.Exit(1)
		}

	},
}
