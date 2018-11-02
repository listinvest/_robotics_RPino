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
		} else {
			log.Printf("Got: %s\n", string(buf))
			if strings.Index(string(buf),sensor) == 0 { // check if the reply is what we asked
				reply := strings.Replace(string(buf),sensor+": ","",1 )
				output = strings.TrimSpace(reply) // strip end of line
			} else {
				log.Printf("Unexpected reply\n")
				failed_read++
				output = "0"
			}
		}
	s.Close()
	return output
}
