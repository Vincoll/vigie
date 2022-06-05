package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"strings"
)

/*
 */

type LogConf struct {
	// Overwrite the log Level
	Level string `env:"LOGLEVEL,default=info"`
}

// NewLogger constructs a Sugared Logger that writes to stdout and
// provides human readable timestamps.
func NewLogger(appName, env, level string) (*zap.SugaredLogger, error) {

	var logCfg zap.Config

	if strings.ToLower(env) == "prod" {
		logCfg = zap.NewProductionConfig()
	} else {
		logCfg = zap.NewDevelopmentConfig()
	}

	logCfg.EncoderConfig = zapcore.EncoderConfig{
		MessageKey: "message",

		LevelKey:    "level",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		TimeKey:    "time",
		EncodeTime: zapcore.ISO8601TimeEncoder,

		CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	logCfg.OutputPaths = []string{"stdout"}
	logCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
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
	zap.ReplaceGlobals(log)
	//zap.S().Infow("An info message", "iteration", 1)

	return log.Sugar(), nil
}
