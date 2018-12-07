package main

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
	"github.com/stianeikeland/go-rpio"
)


var (
	gpio1        chan (string)
	gpio2        chan (string)
)


func init() {
	gpio1 = make(chan string)
	gpio2 = make(chan string)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
}

func speak() {
	longv := ""
	sermon := "espeak -g 5 \"" + conf.Speech + ".\n"
	for _, v := range conf.Relevant_sensors {
		val := strconv.Itoa(arduino_linear_stat[v])
		if v == "H" { longv = "humidity"}
		if v == "T" { longv = "temperature"}
		sermon = sermon + longv + " is " + val + ".\n"
	}
	sermon = sermon + "\""
	log.Printf("%s\n", sermon)
	cmd := exec.Command("bash", "-c", sermon)
	err := cmd.Start()
	if err != nil {
		log.Printf("error for speaking")
	}
}

func human_presence() {
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)
	for t := range ticker.C {
		os.Stderr.WriteString(t.String())
		lock.Lock()
		presence := arduino_linear_stat["U"]
		lock.Unlock()
		if presence == 1 {
			if conf.Verbose { log.Printf("Human detected\n")}
			speak()
		} else {
			if conf.Verbose { log.Printf("Human NOT detected\n")}
		}
	}
}

func alarm_mgr() {
	// Open and map memory to access gpio, check for errors
	pin := rpio.Pin(conf.Alarm_pin)
        if err := rpio.Open(); err != nil {
                log.Fatal(err)
                os.Exit(1)
        }
	pin.Output()
	pin.Low()
        defer rpio.Close()
	time.Sleep(time.Minute) //wait for PIR initialization
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)

	for _ = range ticker.C {
		lock.Lock()
		actual_temp := arduino_linear_stat["T"]
		lock.Unlock()
		if actual_temp < conf.Critical_temp {
			log.Printf("Alarm triggered!!\n")
			pin.High()
			time.Sleep(time.Second*30)
			pin.Low()
		}
	}
}


func command_socket(socket string) (reply string) {
	if socket == "on" {
		gpio1 <- "on"
		reply = "Turning ON"
	} else if socket == "off" {
		gpio1 <- "off"
		reply = "Turning OFF"
	} else {
		reply = "Specify 'on' or 'off'"
	}
	return reply
}


func send_gpio1(gpio1 <-chan string) {
	pin := rpio.Pin(conf.Socket1)
	pin.Output()
	for {
		status := <-gpio1
		log.Printf("Sending %s to GPIO1", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

func send_gpio2(gpio2 <-chan string) {
	pin := rpio.Pin(conf.Socket2)
	pin.Output()
	for {
		status := <-gpio2
		log.Printf("Sending %s to GPIO2", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

