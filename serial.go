package main

import (
	"log"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func comm_arduino() {
        //arduino_in = make(chan string)
        //arduino_out = make(chan string)

	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	sensor :=  <- arduino_in

	cmd := sensor+"?\n"
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Printf("%s\n",err)
	}
	log.Printf("Asked for: %s \n", sensor)

	buf := make([]byte, 8)
	_, err = s.Read(buf)
	if err != nil {
		log.Printf("%s\n",err)
	}
	log.Printf("Got: %s\n", string(buf))
	arduino_out <- strings.Replace(string(buf),sensor+": ","",1 )
}
