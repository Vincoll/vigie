package tsdb

type ConfInfluxDB struct {
	Enable   bool   `toml:"enable"`
	Addr     string `toml:"addr"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
}

type ConfInfluxDB2 struct {
	Enable       bool   `toml:"enable"`
	Addr         string `toml:"addr"`
	Organization string `toml:"organization"`
	Bucket       string `toml:"bucket"`
	Precision    string `toml:"precision"`
	Token        string `toml:"token"`
}

type ConfWarp10 struct {
	Addr  string
	Token string
}
