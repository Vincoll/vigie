package alertmanager

import "time"

type ConfAlerting struct {
	Enable   bool          `toml:"enable"`
	Interval time.Duration `toml:"interval"`
	Reminder time.Duration `toml:"reminder"`
	Email    struct {
		To       string `toml:"to"`
		From     string `toml:"from"`
		Username string `toml:"username"`
		Password string `toml:"password"`
		SMTP     string `toml:"smtp"`
		Port     int    `toml:"port"`
	} `toml:"email"`
	Slack struct {
		Hook string `toml:"hook" valid:"url"`
	} `toml:"slack"`
	Discord struct {
		Hook string `toml:"hook" valid:"url"`
	} `toml:"discord"`
	Webhook struct {
		Hook string `toml:"hook" valid:"url"`
	} `toml:"webhook"`
}
