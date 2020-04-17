package main

import (
	toml "github.com/BurntSushi/toml"
	"log"
)

type config struct {
	Listen       string  `toml:"listen"`
	Sensors      sensors `toml:"sensors"`
	Verbose      bool    `toml:"verbose"`
	Time_server  string  `toml:"time_server"`
	Logfile      string  `toml:"log_file"`
	Pidfile      string  `toml:"pid_file"`
	Inputs       map[string]rpigpio
	Outputs      map[string]rpigpio
	Lighting     light       `toml:"lighting"`
	Alarms       alarms      `toml:"alarms"`
	Analysis     analysis    `toml:"data_analysis"`
	Speech       speech      `toml:"speech"`
	Serial       serial_conf `toml:"serial"`
	Temp_control tempc       `toml:"temp_control"`
}

type rpigpio struct {
	PIN int
}

type tempc struct {
	Critical_temp int `toml:"critical_temp"`
	Enabled       bool
	Tap_open      int `toml:"tap_open"`
}

type sensors struct {
	Arduino_linear []string `toml:"arduino_linear"`
	Arduino_exp    []string `toml:"arduino_exp"`
	Poll_interval  int      `toml:"poll_interval"`
	Adj_H          map[string]int
	Adj_T          map[string]int
	Bmp            bool `toml:"bmp"`
	Dht            bool `toml:"dht"`
	Dht_pin        int `toml:"dht_pin"`
	Sds11          bool `toml:"sds11"`
}

type speech struct {
	Read    []string
	Message string `toml:"speech"`
}

type serial_conf struct {
	Tty     string `toml:"tty"`
	Baud    int    `toml:"baud"`
	Timeout int    `toml:"timeout"`
}

type alarms struct {
	Email_enabled bool
	Siren_enabled bool
	Presence      bool
	Critical_temp int `toml:"critical_temp"`
	Slack_token   string `toml:"token"`
	Smtp          string `toml:"smtp"`
	Mailbox       string `toml:"mailbox"`
	Auth_user     string `toml:"auth_user"`
	Auth_pwd      string `toml:"auth_pwd"`
}

type analysis struct {
	Cache_age   int     `toml:"cache_age"`
	Depth       int     `toml:"historic_depth"`
	Upper_limit float32 `toml:"upper_limit"`
	Lower_limit float32 `toml:"lower_limit"`
}

type light struct {
	Red   int `toml:"red_threshold"`
	Start int `toml:"start_hour"`
	End   int `toml:"end_hour"`
}

func loadConfig(path string) *config {
	conf := &config{}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
