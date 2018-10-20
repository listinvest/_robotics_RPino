package main

import (
	"log"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func comm_arduino(sensor string) (string){
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 10}
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
		_, err = s.Read(buf)
		if err != nil {
			log.Printf("%s\n",err)
		}
		log.Printf("Got: %s\n", string(buf))
		return strings.Replace(string(buf),sensor+": ","",1 )
}
