package main

import (
	"github.com/tarm/serial"
	"log"
	"os"
	"time"
)

func main() {
	cmd := os.Args[1]
	log.Printf("sent instruction: %s",cmd)
	reply := comm2_arduino(cmd)
	if reply != "null" {
		log.Printf("got reply: %s", reply)
	} else {
		log.Printf("No reply!")
	}
}

func comm2_arduino(sensor string) (output string) {
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600, ReadTimeout: time.Millisecond * time.Duration(2000)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Printf("%s\n", err)
		return
	}
	cmd := sensor + "?\n"
	starts := time.Now()
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Printf("%s\n", err)
	}
	buf := []byte("________")
	nbytes, failed := s.Read(buf)
	t := time.Now()
	elapseds := t.Sub(starts)
	if nbytes < 3 {
		_, failed = s.Read(buf)
	}
	if failed != nil {
		log.Printf("error: %s\n", failed)
		output = "null"
	} else {
		reply := string(buf)
		log.Printf("Got %d bytes: %s, took %f", nbytes, reply, elapseds.Seconds())
	}
	s.Close()
	return output
}
