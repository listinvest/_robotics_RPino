package main

import (
	"github.com/tarm/serial"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func initialize_arduino() {
	if conf.Serial.Tty == "none" {
		return
	}
	// initialize maps
	n := len(conf.Sensors.Arduino_linear)
	arduino_linear_stat = make(map[string]int, n)
	arduino_prev_linear_stat = make(map[string][]int, n)
	n = len(conf.Sensors.Arduino_exp)
	arduino_exp_stat = make(map[string]int, n)
	arduino_cache_stat = make(map[string]int, n)
	arduino_prev_exp_stat = make(map[string][]int, n)
	history_setup()
	arduino_connected = true
}

func read_arduino() {
	if conf.Serial.Tty == "none" {
		return
	}
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	reply := ""
	for _, s := range conf.Sensors.Arduino_linear {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				validated = last_linear(s)
			} else {
				validated = output
			}
			add_linear(s, output)
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_linear(s)
			arduino_cache_stat[s] = arduino_cache_stat[s] + 1
		}
		reply = ""
		if arduino_cache_stat[s] > conf.Analysis.Cache_age { validated = 0 }
		lock.Lock()
		arduino_linear_stat[s] = validated
		lock.Unlock()
		time.Sleep(time.Second * 2)
	}
	time.Sleep(time.Second * 2)
	for _, s := range conf.Sensors.Arduino_exp {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				validated = last_exp(s)
			} else {
				validated = output
			}
			// add every value we recieve to the history
			add_exp(s, validated)
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_exp(s)
			arduino_cache_stat[s] = arduino_cache_stat[s] + 1
		}

		reply = ""
		if arduino_cache_stat[s] > conf.Analysis.Cache_age { validated = 0 }
		lock.Lock()
		if validated > 0 {
			inverted := int(1 / float32(validated) * 10000)
			arduino_exp_stat[s] = inverted
		}
		lock.Unlock()
		time.Sleep(time.Second * 2)
	}
	check := comm2_arduino("S")
	lock.Lock()
	arduino_linear_stat["check_error"] = 0
	if !strings.Contains(check, "ok") { // check if the reply is what we asked
		log.Printf("Periodic check failed (%q)!\n", check)
		arduino_linear_stat["check_error"] = 1
	}
	lock.Unlock()
}

func comm2_arduino(sensor string) (output string) {
	if conf.Serial.Tty == "none" {
		return
	}
	c := &serial.Config{Name: conf.Serial.Tty, Baud: conf.Serial.Baud, ReadTimeout: time.Millisecond * time.Duration(conf.Serial.Timeout)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Printf("%s\n", err)
		arduino_connected = false
		return
	}
	reg,_ := regexp.Compile("[^0-9]+")
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
		} else {
			log.Printf("Unexpected reply\n")
			output = "null"
		}
	}
	s.Close()
	return output
}
<<<<<<< HEAD

// not useful anymore with USB connection
func flush_serial() {
	if conf.Serial.Tty == "none" {
		return
	}
	c := &serial.Config{Name: conf.Serial.Tty, Baud: conf.Serial.Baud, ReadTimeout: time.Millisecond * time.Duration(conf.Serial.Timeout)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Printf("%s\n", err)
		arduino_connected = false
		return
	}
	buf := make([]byte, 16)
	_, _ = s.Read(buf)
	s.Close()
}
=======
>>>>>>> 2ffda0bc0121cfba9c0c4976fd168e24e9a53132
