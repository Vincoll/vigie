package webapi

type ConfWebAPI struct {
	Enable      bool   `toml:"enable"`
	Hostname    string `toml:"hostname" valid:"hostname"`
	Port        int    `toml:"port" valid:"port"`
	Environment string `toml:"environment"`
}
