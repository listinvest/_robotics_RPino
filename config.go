package main

import (
	"log"
	toml "github.com/BurntSushi/toml"
)

type config struct {
	Listen            string   `toml:"listen"`
	Poll_interval	  int	   `toml:"poll_interval"`
	Arduino_linear_sensors   []string `toml:"arduino_linear_sensors"`
	Arduino_exp_sensors   []string `toml:"arduino_exp_sensors"`
	Verbose           bool     `toml:"verbose"`
	Inputs		  map[string]sensor
	Outputs		  map[string]sensor
	Alarms		  alarms `toml:"alarms"`
	Analysis	  analysis `toml:"data_analysis"`
	Speech		  speech `toml:"speech"`
}

type sensor struct {
	PIN int
}

type speech struct {
	Sensors	 []string
	Message	 string	`toml:"speech"`
}

type alarms struct {
	Critical_temp  int `toml:"critical_temp"`
	Email_enabled bool
	Siren_enabled bool
	Smtp string `toml:"smtp"`
	Mailbox string `toml:"mailbox"`
	Auth_user string `toml:"auth_user"`
	Auth_pwd string `toml:"auth_pwd"`
}

type analysis struct {
	Depth		  int	   `toml:"historic_depth"`
	Cache_limit	  int	   `toml:"cache_limit"`
	Percentile	  float32  `toml:"percentile"`
	Upper_limit	  float32  `toml:"upper_limit"`
	Lower_limit	  float32  `toml:"lower_limit"`
}


func loadConfig(path string) (*config) {
	conf := &config{}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}

