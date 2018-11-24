package main

import (
	"log"
	toml "github.com/BurntSushi/toml"
)

type config struct {
	Listen            string   `toml:"listen"`
	Socket1           int      `toml:"socket1"`
	Socket2           int      `toml:"socket2"`
	Poll_interval	  int	   `toml:"poll_interval"`
	Critical_temp	  int	   `toml:"critical_temp"`
	Upper_limit	  float32  `toml:"upper_limit"`
	Lower_limit	  float32  `toml:"lower_limit"`
	Alarm_pin	  int	   `toml:"alarm_pin"`
	Arduino_sensors   []string `toml:"arduino_sensors"`
	Relevant_sensors  []string `toml:"relevant_sensors"`
	Verbose           bool     `toml:"verbose"`
	Zero_unreadable   bool     `toml:"zero_unreadable"`
	Speech            string   `toml:"speech"`
}

func loadConfig(path string) (*config) {
	conf := &config{}
	metaData, err := toml.DecodeFile(path, conf)
	if err != nil {
		log.Fatal(err)
	}
	for _, key := range metaData.Undecoded() {
		log.Printf("unknown option %q\n", key.String())
	}

	return conf
}

