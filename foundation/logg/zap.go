// Package logg provides a convience function to constructing a logg for use.
package logg

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/*



 */

type LogConf struct {
	// Overwrite the log Level
	Level string `env:"LOGLEVEL,default=info"`
}

// NewLogger constructs a Sugared Logger that writes to stdout and
// provides human readable timestamps.
func NewLogger(appName, env, loglevel string) (*zap.SugaredLogger, error) {

	var logCfg zap.Config

	zlvl, err := zapcore.ParseLevel(loglevel)
	if err != nil {
		return nil, fmt.Errorf("log level is not valid for Zap Log library: %s", err)
	}

	// if prod* then use production config

	if strings.HasPrefix(env, "prod") {

		logCfg = zap.NewProductionConfig()

		logCfg.EncoderConfig = zapcore.EncoderConfig{
			MessageKey: "message",

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		}

	} else {
		logCfg = zap.NewDevelopmentConfig()

		logCfg.EncoderConfig = zapcore.EncoderConfig{
			MessageKey: "message",

			LevelKey:    "level",
			EncodeLevel: zapcore.CapitalLevelEncoder,

			TimeKey:    "time",
			EncodeTime: zapcore.ISO8601TimeEncoder,

			CallerKey:    "caller",
			EncodeCaller: zapcore.ShortCallerEncoder,
		}
	}

	logCfg.Level.SetLevel(zlvl)

	logCfg.OutputPaths = []string{"stdout"}
	logCfg.DisableStacktrace = false
	logCfg.InitialFields = map[string]interface{}{
		"app_name": appName,
	}

	log, err := logCfg.Build()
	if err != nil {
		return nil, err
	}

	/*
		INFO:
		It's possible to replace and use a Global Logger with Zap zap.L() || zap.S().
		However, this way is discouraged in favor of the passage by argument.
		https://github.com/uber-go/zap/blob/master/FAQ.md#why-include-package-global-loggers
		Both are enabled for the example.
	*/
	//zap.ReplaceGlobals(log)
	//zap.S().Infow("An info message", "iteration", 1)

	return log.Sugar(), nil
}

// https://stackoverflow.com/questions/57745017/how-to-initialize-a-zap-logger-once-and-reuse-it-in-other-go-files

// https://github.com/sandipb/zap-examples

/*
https://blog.sandipb.net/2018/05/03/using-zap-creating-custom-loggers/
https://blog.sandipb.net/2018/05/02/using-zap-simple-use-cases/
*/

/*
// Source : https://blog.sandipb.net/2018/05/02/using-zap-simple-use-cases/

logg, _ = zap.NewProduction()
logg.Debug("This is a DEBUG message")
logg.Info("This is an INFO message")
logg.Info("This is an INFO message with fields", zap.String("region", "us-west"), zap.Int("id", 2))
logg.Warn("This is a WARN message")
logg.Error("This is an ERROR message")
// logg.Fatal("This is a FATAL message")   // would exit if uncommented
logg.DPanic("This is a DPANIC message")
// logg.Panic("This is a PANIC message")   // would exit if uncommented


logg, _ = zap.NewDevelopment()
logg.Debug("This is a DEBUG message")
logg.Info("This is an INFO message")
logg.Info("This is an INFO message with fields", zap.String("region", "us-west"), zap.Int("id", 2))
logg.Warn("This is a WARN message")
logg.Error("This is an ERROR message")
// logg.Fatal("This is a FATAL message")   // would exit if uncommented
// logg.DPanic("This is a DPANIC message") // would exit if uncommented
//logg.Panic("This is a PANIC message")    // would exit if uncommented


By comparing the outputs you can make the following observations:

    Both Example and Production loggers use the JSON encoder. Development uses the Console encoder
    The logg.DPanic() function causes a panic in Development logg but not in Example or Production
    The Development logg:
        Prints a stack trace from Warn level and up.
        Always prints the package/file/line number (the caller)
        Tacks any extra fields as a json string at the end of the line
        Prints the level names in uppercase
        Prints timestamp in ISO8601 format with milliseconds
    The Production logg:
        Doesn’t log messages at debug level
        Adds stack trace as a json field for Error, DPanic levels, but not for Warn
        Always adds the caller as a json field
        Prints timestamp in epoch format
        Prints level names in lower case


*/
