package run

import (
	"fmt"
	"os"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/vincoll/vigie/internal/worker/pulsar_worker"
	"github.com/vincoll/vigie/pkg/core"
	"github.com/vincoll/vigie/pkg/ha"
	"github.com/vincoll/vigie/pkg/load"
	"github.com/vincoll/vigie/pkg/promexporter"
	"github.com/vincoll/vigie/pkg/tsdb"
	"github.com/vincoll/vigie/pkg/utils/dnscache"
	"github.com/vincoll/vigie/pkg/webapi"
	"go.uber.org/zap"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vincoll/vigie/cmd/vigie/version"
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
	Use:     "run",
	Example: "run --config ./config/webapi.toml",
	Short:   "Old cmd",
	PreRun: func(cmd *cobra.Command, args []string) {
		//
		// Create Vigie Instance
		//
		var err error

		vigieInstance, err = vigie.NewVigie()
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"component": "webapi", "status": "error", "error": "Vigie cannot start"}).Fatal(err)
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
		_, err := govalidator.ValidateStruct(vigieConf)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"component": "configfile", "status": "invalid", "error": "Vigie APIServerConfig File is invalid"}).Fatal(err)
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
				utils.Log.WithFields(logrus.Fields{"component": "webapi", "status": "failed", "error": errWebAPI}).Fatal("[ConfWebAPI] has failed to start")
				os.Exit(2)
			}
		}

		//
		// Start Consul
		//
		if vigieConf.HA.Enable {
			cc, err := ha.InitHAConsul(vigieConf.HA, vigieConf.API.Port)
			if err != nil {
				utils.Log.WithFields(logrus.Fields{"component": "ha", "status": "failed", "error": err}).Fatal("CC Consul failed to start")
				os.Exit(2)
			}
			vigieInstance.ConsulClient = cc
		}

		//
		// Load TSDBs Configs
		//

		// Init Manager
		tsdb.TsdbMgr.Tags = vigieInstance.HostInfo.Tags

		/* Load vInfluxDB APIServerConfig                                      DEPRECATED
		if vigieConf.InfluxDB.Enable {
			idb, errIDB := tsdb.NewInfluxDB(vigieConf.InfluxDB)
			if errIDB != nil {
				utils.Log.Fatal("Vigie failed to connect with InfluxDB: ", errIDB)
			}
			tsdb.TsdbMgr.AddTsdb(idb)
		}

		*/

		/* Load vInfluxDB APIServerConfig
		if vigieConf.InfluxDBv2.Enable {
			idb2, errIDB := tsdb.NewInfluxDBv2(vigieConf.InfluxDBv2)
			if errIDB != nil {
				utils.Log.Fatal("Vigie failed to connect with InfluxDBv2: ", errIDB)
			}
			tsdb.TsdbMgr.AddTsdb(idb2)
		}

		/* Load warp10 APIServerConfig
		if vigieConf.Warp10.Enable {
			w10, errW10 := tsdb.NewWarp10(vigieConf.Warp10)
			if errW10 != nil {
				utils.Log.Fatal("Vigie failed to connect with Warp10: ", errW10)
			}
			tsdb.TsdbMgr.AddTsdb(w10)
		}
		*/
		// DIRTY POINTER !!!
		//vigieInstance.TsdbManager = &tsdb.TsdbMgr

		//
		// Start Prometheus Exporter
		//

		// Start ConfWebAPI
		if vigieConf.Pulsar.Enable {
			_, errPulsar := pulsar_worker.NewClient(pulsar_worker.ConfPulsar(vigieConf.Pulsar))
			if errPulsar != nil {
				zap.S().Fatalf("Could not instantiate Pulsar client: %s ", err)
				os.Exit(2)
			}
		}

		if vigieConf.Prometheus.Enable {
			go promexporter.InitPromExporter(vigieConf.Prometheus)
		}

		//
		// Init AlertManager                                      DEPRECATED
		//

		if vigieConf.Alerting.Enable {
			// go alertmanager.InitAlertManager(vigieConf.Alerting, vigieInstance.HostInfo.Name, vigieInstance.HostInfo.URL)
		}

		//
		// Init ImportManager and add it to Vigie Instance
		//

		vigieInstance.ImportManager, err = load.InitImportManager(vigieConf.Import, vigieInstance.ConsulClient)
		if err != nil {
			utils.Log.WithFields(logrus.Fields{"component": "import", "status": "failed", "error": err}).Fatal("[ConfImport] fail to validate the import.")
			os.Exit(1)
		}

		// DNSCACHED
		core.VigieServer.CacheDNS, _ = dnscache.NewCachedResolver()

		//
		// Start Vigie Instance
		//

		err = vigieInstance.Start()
		if err != nil {
			utils.Log.Fatal("Vigie failed to launch: ", err)
			os.Exit(1)
		}

	},
}
