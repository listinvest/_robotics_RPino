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
}

type sensor struct {
	PIN int
}

type alarms struct {
	Speech_sensors	 []string
	Speech_message	 string	`toml:"speech"`
	Critical_temp	  int	   `toml:"critical_temp"`
}

type analysis struct {
	Depth		  int	   `toml:"historic_depth"`
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

