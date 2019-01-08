package main

import (
    "log"
    "sync"
     "time"
    "github.com/beevik/ntp"
)

var hour int
var Tlock = &sync.Mutex{}

func get_time() {
	if conf.Verbose { log.Printf("Get time from: %s\n", conf.Time_server) }
	actual_hour := 0
	options := ntp.QueryOptions{ Timeout: 10*time.Second, TTL: 5 }
	response,_ := ntp.QueryWithOptions(conf.Time_server, options)
	remote_time := response.Time
	hour,_,_ = remote_time.Clock()
	Lticker := time.NewTicker(time.Minute)
	defer Lticker.Stop()
	for _ = range Lticker.C {
		options := ntp.QueryOptions{ Timeout: 10*time.Second, TTL: 5 }
		response,errr := ntp.QueryWithOptions(conf.Time_server, options)
		if errr == nil {
			remote_time := response.Time
			actual_hour,_,_ = remote_time.Clock()
		} else {
			log.Printf("Error NTP: %s  !", errr)
			now := time.Now()
			actual_hour,_,_ = now.Clock()
			lock.Lock()
			rpi_stat["ntp_error"] = 1
			lock.Unlock()
		}
		if conf.Verbose { log.Printf("CLOCK h: %d ", actual_hour) }
		Tlock.Lock()
		hour = actual_hour
		Tlock.Unlock()
	}
}

func light_mgr() {
	if ( conf.Lighting["morning_start"].Hour == 0 && conf.Lighting["morning_end"].Hour == 0 && conf.Lighting["evening_start"].Hour ==0 &&  conf.Lighting["evening_end"].Hour ==0 )  { return }
	Tlock.Lock()
	now := hour
	Tlock.Unlock()
	on := 0
	if ( conf.Lighting["morning_start"].Hour < now && now < conf.Lighting["morning_end"].Hour)  || ( conf.Lighting["evening_start"].Hour < now && now < conf.Lighting["evening_end"].Hour)  {
		if conf.Verbose { log.Printf("Lights on (h:%d)", now) }
		on = 1
	} else {
		if conf.Verbose { log.Printf("Lights off (h:%d)", now) }
	}
	lock.Lock()
	red := arduino_exp_stats["R"]
	lock.Unlock()
	if red < conf.Lighting["minimum_red"].Hour {
		log.Printf("Red light component is lower than threshold - %d", conf.Lighting["minimum_red"].Hour)
		on = 1
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

