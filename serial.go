package main

import (
	"log"
	"github.com/tarm/serial"
	"regexp"
	"strings"
	"time"
)

func comm2_arduino(sensor string) (output string){
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	reg, err := regexp.Compile("[^0-9]+")
	cmd := sensor+"?\n"
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Printf("%s\n",err)
	}
	if conf.Verbose { log.Printf("Asked: %s", cmd) }
	buf := make([]byte, 7)
	_, err = s.Read(buf)
	if err != nil {
		log.Printf("%s\n",err)
		failed_read++
		output = "null"
	} else {
		reply := string(buf)
		if conf.Verbose { log.Printf("Got: %s", reply) }
		if strings.Index(reply,sensor) == 0 { // check if the reply is what we asked
			if sensor == "S" {
				return "ok"
			}
			reply = strings.Replace(reply,sensor+": ","",1 )
			output = reg.ReplaceAllString(reply, "")
			good_read++
		} else {
			log.Printf("Unexpected reply\n")
			failed_read++
			output = "null"
		}
	}
	s.Close()
	return output
}

func flush_serial() {
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 16)
	_, err = s.Read(buf)
	s.Close()
}
