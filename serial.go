package main

import (
	//"bytes"
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
			output = "null"
		} else {
			log.Printf("Got: %s\n", string(buf))
			if strings.Index(string(buf),sensor) == 0 { // check if the reply is what we asked
				reply := strings.Replace(string(buf),sensor+": ","",1 )
				reply = strings.TrimSpace(reply) // strip end of line
				good_read++
				output = strings.Trim(reply, "\x00") //strip null chars
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
