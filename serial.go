package main

import (
	"github.com/tarm/serial"
	"log"
	"regexp"
	"strings"
	"time"
)

func comm2_arduino(sensor string) (output string) {
	c := &serial.Config{Name: conf.Serial.Tty, Baud: conf.Serial.Baud, ReadTimeout: time.Millisecond * time.Duration(conf.Serial.Timeout)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	reg, err := regexp.Compile("[^0-9]+")
	cmd := sensor + "?\n"
	starts := time.Now()
	_, err = s.Write([]byte(cmd))
	if err != nil {
		log.Printf("%s\n", err)
	}
	if conf.Verbose {
		log.Printf("Asked: %s", cmd)
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
		serial_stat["failed_read"] = serial_stat["failed_read"] + 1
		output = "null"
	} else {
		reply := string(buf)
		if conf.Verbose {
			log.Printf("Got %d bytes: %s, took %f", nbytes, reply, elapseds.Seconds())
		}
		if strings.Index(reply, sensor) == 0 { // check if the reply is what we asked
			if sensor == "S" {
				return "ok"
			}
			reply = strings.Replace(reply, sensor+": ", "", 1)
			output = reg.ReplaceAllString(reply, "")
			serial_stat["good_read"] = serial_stat["good_read"] + 1
		} else {
			log.Printf("Unexpected reply\n")
			serial_stat["failed_read"] = serial_stat["failed_read"] + 1
			output = "null"
		}
	}
	s.Close()
	return output
}

func flush_serial() {
	c := &serial.Config{Name: conf.Serial.Tty, Baud: conf.Serial.Baud, ReadTimeout: time.Millisecond * time.Duration(conf.Serial.Timeout)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 16)
	_, err = s.Read(buf)
	s.Close()
}
