package promexporter

type ConfPrometheus struct {
	Enable      bool   `toml:"enable"`
	Port        int    `toml:"port" valid:"port"`
	Gometrics   bool   `toml:"gometrics"`
	Environment string `toml:"environment"`
}
