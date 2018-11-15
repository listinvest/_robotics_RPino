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
	Alarm_pin	  int	   `toml:"alarm_pin"`
	Arduino_sensors   []string  `toml:"arduino_sensors"`
	Relevant_sensors   []string  `toml:"relevant_sensors"`
	Verbose           bool     `toml:"verbose"`
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

