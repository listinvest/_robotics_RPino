package main

import (
	toml "github.com/BurntSushi/toml"
	"log"
)

type config struct {
	Listen   string  `toml:"listen"`
	Sensors  sensors `toml:"sensors"`
	Verbose  bool   `toml:"verbose"`
	Time_server  string    `toml:"time_server"`
	Inputs   map[string]rpigpio
	Outputs  map[string]rpigpio
	Lighting light	`toml:"lighting"`
	Alarms   alarms      `toml:"alarms"`
	Analysis analysis    `toml:"data_analysis"`
	Speech   speech      `toml:"speech"`
	Serial   serial_conf `toml:"serial"`
}

type hours struct {
	Hour int
}

type rpigpio struct {
	PIN int
}

type value struct {
	Value int
}

type sensors struct {
	Arduino_linear []string `toml:"arduino_linear"`
	Arduino_exp    []string `toml:"arduino_exp"`
	Poll_interval  int      `toml:"poll_interval"`
	Adj_H          map[string]int
	Adj_T          map[string]int
	Bmp	       int	`toml:"bmp"`
	Dht	       int	`toml:"dht"`
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
	Critical_temp int `toml:"critical_temp"`
	Email_enabled bool
	Siren_enabled bool
	Presence      bool
	Slack_token   string `toml:"token"`
	Smtp          string `toml:"smtp"`
	Mailbox       string `toml:"mailbox"`
	Auth_user     string `toml:"auth_user"`
	Auth_pwd      string `toml:"auth_pwd"`
}

type analysis struct {
	Depth       int     `toml:"historic_depth"`
}

type light struct {
	Red int `toml:"red_threshold"`
	Start int `toml:"start_hour"`
	End int `toml:"end_hour"`
}

func loadConfig(path string) *config {
	conf := &config{}
	_, err := toml.DecodeFile(path, conf)
	if err != nil {
		log.Fatal(err)
	}

	return conf
}
