package main

import (
	"log"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func comm_arduino() {
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 10}
	for {
	select {
	case sensor :=  <-arduino_in:
		log.Printf("Being asked for: %s \n", sensor)
		s, err := serial.OpenPort(c)
		if err != nil {
			log.Fatal(err)
		}

		cmd := sensor+"?\n"
		_, err = s.Write([]byte(cmd))
		if err != nil {
			log.Printf("%s\n",err)
		}
		log.Printf("Asked for: %s \n", sensor)
		buf := make([]byte, 8)
		n, err := s.Read(buf)
		if err != nil {
			log.Printf("%s\n",err)
			arduino_out <- "0"
		} else {
			log.Printf("read %d bytes\n", n)
			log.Printf("Got: %s\n", string(buf))
			reply := strings.Replace(string(buf),sensor+": ","",1 ) // need to check if it's the right reply
			reply = strings.Replace(reply,"\n","",1 ) // strip end of line
			reply = strings.Replace(reply,"\r","",1 ) // strip end of line
			arduino_out <- strings.Replace(string(reply),".00","",1 ) // stip .00
		}
		s.Close()
	case <-time.After(3 * time.Second):
		log.Println("timeout 2")
	}
	}
}
