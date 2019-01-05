package main

import (
    "fmt"
    "log"
    "sync"
     "time"
    "github.com/beevik/ntp"
)

var hour int
var Tlock = &sync.Mutex{}

func get_time() {
	if conf.Verbose { log.Printf("Get time from: %s\n", conf.Time_server) }

	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * 10)

	for _ = range ticker.C {
		response,errr := ntp.Query(conf.Time_server)
		if errr == nil {
			time := time.Now().Add(response.ClockOffset)
			Tlock.Lock()
			hour = time.Hour()
			Tlock.Unlock()
		} else {
			fmt.Printf("Error NTP: %s  ! \n", errr)
		}
	}
}

func internal_cron() {
	Tlock.Lock()
	now := hour
	Tlock.Unlock()
	if ( conf.Lighting["morning_start"].Hour < now && now < conf.Lighting["morning_end"].Hour)  || ( conf.Lighting["evening_start"].Hour < now && now < conf.Lighting["evening_end"].Hour)  {
		log.Println("Lights on")
		gpio2 <- "on"
	} else {
		gpio2 <- "off"
	}

}
