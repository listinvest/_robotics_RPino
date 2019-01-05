package main

import (
    "fmt"
     "time"
    "github.com/beevik/ntp"
)

func get_time() (hour int) {

	response,errr := ntp.Query(conf.Time_server)
	if errr == nil {
		time := time.Now().Add(response.ClockOffset)
		hour = time.Hour()
		if conf.Verbose { fmt.Printf("hour: %d\n", hour) }
	} else {
		fmt.Printf("Error : %s  ! \n", errr)
	}
	return hour
}

func internal_cron() {
//	if morning_start < get_time() < morning_end
//  		gpio2 <- "on"


}
