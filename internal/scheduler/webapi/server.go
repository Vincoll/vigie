package webapi

type APIServerConfig struct {
	TechPort string `toml:"TechPort"`
	Pprof    string `toml:"pprof"`
	Env      string
}
