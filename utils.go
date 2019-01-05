package main

import (
    "fmt"
     "time"
    "github.com/beevik/ntp"
)

func get_time() {

	response,errr := ntp.Query(conf.Time_server)
	if errr == nil {
		time := time.Now().Add(response.ClockOffset)
		h,m,s := time.Clock()
		fmt.Printf("%d - %d - %d \n", h,m,s)
	} else {
		fmt.Printf("Error : %s  ! \n", errr)
	}
}
