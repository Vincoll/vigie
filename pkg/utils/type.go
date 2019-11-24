package utils

type LogConf struct {
	Stdout      bool   `toml:"stdout"`
	LogFile     bool   `toml:"logfile"`
	Level       string `toml:"level"`
	FilePath    string `toml:"filePath"`
	Environment string `toml:"environment"`
}
