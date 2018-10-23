package main

import (
	"log"
	"github.com/tarm/serial"
	"strings"
	"time"
)

func comm2_arduino(sensor string) (output string){
	c := &serial.Config{Name: "/dev/ttyAMA0", Baud: 9600, ReadTimeout: time.Second * 5}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

		cmd := sensor+"?\n"
		_, err = s.Write([]byte(cmd))
		if err != nil {
			log.Printf("%s\n",err)
		}
		log.Printf("Asked: %s \n", cmd)
		buf := make([]byte, 7)
		_, err = s.Read(buf)
		if err != nil {
			log.Printf("%s\n",err)
			output = "0"
		//} else if n < 8 {
		//	log.Printf("read less than 8 bytes (%d)\n",n)
		//	output = "0"

		} else {
			log.Printf("Got: %s\n", string(buf))
			reply := strings.Replace(string(buf),sensor+": ","",1 ) // need to check if it's the right reply
			output = strings.TrimSpace(reply) // strip end of line
		}
	s.Close()
	return output
}
