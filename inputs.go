package main

import (
	"fmt"
	"log"
	"time"
	"github.com/stianeikeland/go-rpio"
	//"github.com/d2r2/go-bsbmp"
	//"github.com/d2r2/go-i2c"
)

//var lock = &sync.Mutex{}

func start_inputs() {
	if len(conf.Inputs)==0 { fmt.Println("No GPIO to monitor") }
	for sensor, detail := range conf.Inputs {
		fmt.Printf("Starting watched for sensor: %s on pin %d\n", sensor, detail.PIN)
		go gpio_watch(sensor,detail.PIN)
	}
}

func gpio_watch(sensor string,Spin int) {
	// Open and map memory to access gpio, check for errors
	pin := rpio.Pin(Spin)
        if err := rpio.Open(); err != nil {
                log.Fatal(err)
        }
	pin.Input()
        defer rpio.Close()
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)

	for _ = range ticker.C {
		res := pin.Read()
	        //log.Printf("%s",res)
		lock.Lock()
		arduino_linear_stat[sensor]=int(res)
		lock.Unlock()
	}
}
