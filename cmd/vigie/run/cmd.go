package run

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/vincoll/vigie/pkg/alertmanager"
	"github.com/vincoll/vigie/pkg/promexporter"
	"github.com/vincoll/vigie/pkg/utils/dnscache"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vincoll/vigie/cmd/vigie/version"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/tsdb"
	"github.com/vincoll/vigie/pkg/utils"
	"github.com/vincoll/vigie/pkg/vigie"
	"github.com/vincoll/vigie/pkg/webapi"
)

var (
	configfile    string
	vigieInstance *vigie.Vigie
	variables     []string
	withEnv       bool
)

func init() {
	Cmd.Flags().StringVar(&configfile, "config", "", "--config ./vigie.toml")
	Cmd.Flags().BoolVarP(&withEnv, "env", "", false, "Inject environment variables. export FOO=BAR -> you can use {{.FOO}} in your tests")
}

// Cmd run
var Cmd = &cobra.Command{
	Use:     "run",
	Example: "run --config ./config/vigie.toml",
	Short:   "Start Tests",
	PreRun: func(cmd *cobra.Command, args []string) {
		//
		// Create Vigie Instance
		//
		var err error

		vigieInstance, err = vigie.NewVigie()
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"component": "vigie", "status": "error", "error": "Vigie cannot start"}).Fatal(err)
			os.Exit(1)
		}

	},
	Run: func(cmd *cobra.Command, args []string) {

		vigieConf := loadVigieConfigFile(configfile)

		//
		// Create GLOBAL LOGGER
		//
		utils.InitLogger(vigieConf.Log)

		// Add info about host
		vigieConf.Host.AddHostSytemInfo()
		vigieInstance.HostInfo = vigieConf.Host
		//
		// Check ConfigFile
		//
		_, errValidate := govalidator.ValidateStruct(vigieConf)
		if errValidate != nil {
			utils.Log.WithFields(logrus.Fields{"component": "configfile", "status": "invalid", "error": "Vigie Config File is invalid"}).Fatal(errValidate)
			os.Exit(1)
		}

		//
		// Vigie Info
		//
		utils.Log.WithFields(logrus.Fields{
			"version":    version.LdVersion,
			"goruntime":  runtime.Version(),
			"arch":       fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
			"builddate":  version.LdBuildDate,
			"configfile": configfile,
		}).Infof("Vigie")

		//
		// Start tooling before Vigie instance
		//

		// Start ConfWebAPI
		if vigieConf.API.Enable {
			errWebAPI := webapi.InitWebAPI(vigieConf.API, vigieInstance)
			if errWebAPI != nil {
				utils.Log.WithFields(logrus.Fields{"component": "api", "status": "failed", "error": errWebAPI}).Fatal("[ConfWebAPI] has failed to start")
				os.Exit(1)
			}
		}

		// Load vInfluxDB Config
		if vigieConf.InfluxDB.Enable {
			errInflux := tsdb.LoadInfluxDB(vigieConf.InfluxDB)
			if errInflux != nil {
				utils.Log.Fatal("Vigie failed to connect with InfluxDB:", errInflux)
				os.Exit(1)
			}
		}

		// Start Prometheus Exporter
		if vigieConf.Prometheus.Enable {
			go promexporter.InitPromExporter(vigieConf.Prometheus)
		}

		if vigieConf.Alerting.Enable {
			// Init AlertManager
			go alertmanager.InitAlertManager(vigieConf.Alerting, vigieInstance.HostInfo.Name, vigieInstance.HostInfo.URL)
		}
		//
		// Git clone Tests if enable
		//
		if vigieConf.Git.Clone {
			errGit := load.CloneGitRepo(vigieConf.Git)
			if errGit != nil {
				utils.Log.Fatal("Failed to clone : ", errGit)
			}
		}

		//
		// Loading Conf to Vigie Instance
		//
		var errImportFile error
		vigieInstance.TestsFiles, vigieInstance.VarsFiles, errImportFile = load.ImportFileandVars(vigieConf.Testfiles, vigieConf.Variables)
		if errImportFile != nil {
			utils.Log.Fatalf("Failed to import tests from %q: %s", vigieConf.Testfiles.Included, errImportFile.Error())
			os.Exit(1)
		} else {
			utils.Log.Infof("Files detected from %s : %s", vigieConf.Testfiles.Included, vigieInstance.TestsFiles)
		}

		// DNSCACHED
		utils.CACHEDDNSRESOLVER, _ = dnscache.NewCachedResolver()

		//
		// Start Vigie Instance
		//

		err := vigieInstance.Start()
		if err != nil {
			utils.Log.Fatal("Vigie failed to launch: ", err)
			os.Exit(1)
		}

	},
}
