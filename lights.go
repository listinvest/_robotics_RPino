package main

import (
	"github.com/beevik/ntp"
	"log"
	"sync"
	"time"
)

var hour int

// Tlock is a shared lock
var Tlock = &sync.Mutex{}

func get_time() {
	if conf.Verbose {
		log.Printf("Get time from: %s\n", conf.Time_server)
	}
	actual_hour := 0
	options := ntp.QueryOptions{Timeout: 10 * time.Second, TTL: 5}
	response, err := ntp.QueryWithOptions(conf.Time_server, options)
	if err != nil {
		log.Println("NTP query error")
		now := time.Now()
		hour = now.Hour()
	} else {
		remote_time := response.Time
		hour, _, _ = remote_time.Clock()
	}
	Lticker := time.NewTicker(time.Minute)
	defer Lticker.Stop()
	for range Lticker.C {
		options := ntp.QueryOptions{Timeout: 10 * time.Second, TTL: 5}
		response, errr := ntp.QueryWithOptions(conf.Time_server, options)
		if errr == nil {
			err = response.Validate()
			if err == nil {
				remote_time := response.Time
				actual_hour, _, _ = remote_time.Clock()
				clock_offset = int(response.ClockOffset * time.Millisecond)
				if conf.Verbose {
					log.Printf("Got reply from ntp, offset %d\n", clock_offset)
				}
			} else {
				log.Printf("Unsuitable reply from NTP server\n")
			}
		} else {
			log.Printf("Error NTP: %s  !", errr)
			now := time.Now()
			actual_hour, _, _ = now.Clock()
			lock.Lock()
			rpi_stat["ntp_error"] = 1
			lock.Unlock()
		}
		if conf.Verbose {
			log.Printf("CLOCK h: %d ", actual_hour)
		}
		Tlock.Lock()
		hour = actual_hour
		Tlock.Unlock()
	}
}

func light_mgr() {
	if conf.Lighting.Red == 0 {
		return
	}
	Tlock.Lock()
	now := hour
	Tlock.Unlock()
	lock.Lock()
	red := arduino_exp_stat["R"]
	lock.Unlock()
	on := 0
	if (conf.Lighting.Start < now && now < conf.Lighting.End) && red < conf.Lighting.Red {
		log.Printf("Red light component is %d, lower than threshold %d", red, conf.Lighting.Red)
		if conf.Verbose {
			log.Printf("Lights on (h:%d)", now)
		}
		on = 1
	} else {
		if conf.Verbose {
			log.Printf("Lights off (h:%d)", now)
		}
	}

	lock.Lock()
	rpi_stat["light"] = on
	lock.Unlock()
	if on == 1 {
		gpio2 <- "on"
	} else {
		gpio2 <- "off"
	}
}
