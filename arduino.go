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
	if conf.Serial.Tty == "none" { return }
	// initialize maps
	n := len(conf.Sensors.Arduino_linear)
	arduino_linear_stat = make(map[string]int, n)
	arduino_prev_linear_stat = make(map[string][]int, n)
	n = len(conf.Sensors.Arduino_exp)
	arduino_exp_stat = make(map[string]int, n)
	arduino_cache_stat = make(map[string]int, n)
	arduino_prev_exp_stat = make(map[string][]int, n)
	history_setup()
	serial_stat = make(map[string]int)
	serial_stat["good_read"] = 1
	serial_stat["failed_read"] = 1
	serial_stat["failed_atoi"] = 1
	serial_stat["failed_interval"] = 1
}

func read_arduino() {
	if conf.Serial.Tty == "none" { return }
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	reply := ""
	for _, s := range conf.Sensors.Arduino_linear {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		use_cached := true
                if arduino_cache_stat[s] > conf.Analysis.Cache_limit {
                        // if we used the cached value X times, we will prohibit to use again, this will allow MMA to catch up
                        log.Printf("used cache too much for %s, not using this time\n",s)
                        use_cached = false
                        arduino_cache_stat[s] = 0
                }
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				serial_stat["failed_atoi"] = serial_stat["failed_atoi"] + 1
				validated = last_linear(s)
				log.Printf("failed read, using cached value\n")
				arduino_cache_stat[s] = arduino_cache_stat[s] + 1
			} else {
				ref_value := reference(s,output)
				lower := float32(ref_value) * conf.Analysis.Lower_limit
				upper := float32(ref_value) * conf.Analysis.Upper_limit
				if float32(output) >= lower && float32(output) <= upper {
					log.Printf("value for %s is %d, within the safe boundaries( %f - %d - %f )\n", s, output, lower, ref_value, upper)
					validated = output
				} else {
					validated = last_linear(s)
					log.Printf("value for %s is %d, which outside the safe boundaries( %f - %d - %f ), using cached value %d\n", s, output, lower,ref_value, upper,validated)
					serial_stat["failed_interval"] = serial_stat["failed_interval"] + 1
					arduino_cache_stat[s] = arduino_cache_stat[s] + 1
					if !use_cached {
						log.Printf("Using real value for %s\n",s)
						validated = output
					}
				}
				add_linear(s,output)
			}
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_linear(s)
			arduino_cache_stat[s] = arduino_cache_stat[s] + 1
		}
		reply = ""
		lock.Lock()
		arduino_linear_stat[s] = validated
		lock.Unlock()
		time.Sleep(time.Second * 2)
	}
	time.Sleep(time.Second * 2)
	for _, s := range conf.Sensors.Arduino_exp {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		use_cached := true
		if arduino_cache_stat[s] > conf.Analysis.Cache_limit {
			// if we used the cached value X times, we will prohibit to use again, this will allow MMA to catch up
			log.Printf("used cache too much, not using this time\n")
			use_cached = false
			arduino_cache_stat[s] = 0
		}
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				serial_stat["failed_atoi"] = serial_stat["failed_atoi"] + 1
				validated = last_exp(s)
				arduino_cache_stat[s] = arduino_cache_stat[s] + 1
				log.Printf("failed read, using cached value\n")
			} else {
				ref_value_mma := mma(s, output)
				lower := float32(ref_value_mma) * (conf.Analysis.Lower_limit)
				upper := float32(ref_value_mma) * (conf.Analysis.Upper_limit)
				if float32(output) >= lower && float32(output) <= upper {
					log.Printf("EXP: value for %s is %d, within the safe boundaries( %f - %f - %f )\n", s, output, lower, ref_value_mma, upper)
					validated = output
				} else {
					log.Printf("EXP: value for %s is %d, which outside the safe boundaries( %f - %f - %f )\n", s, output, lower, ref_value_mma, upper)
					serial_stat["failed_interval"] = serial_stat["failed_interval"] + 1
					validated = last_exp(s) //will use prev value
					arduino_cache_stat[s] = arduino_cache_stat[s] + 1
					if !use_cached {
						log.Printf("Using real value\n")
						validated = output
					}
				}
				// add every value we recieve to the history
				add_exp(s,validated)
			}
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_exp(s)
			arduino_cache_stat[s] = arduino_cache_stat[s] + 1
		}

		reply = ""
		lock.Lock()
		if validated > 0 {
			inverted := int(1/float32(validated)*10000)
			arduino_exp_stat[s] = inverted
		}
		lock.Unlock()
		time.Sleep(time.Second * 2)
	}
	check := comm2_arduino("S")
	lock.Lock()
	arduino_linear_stat["check_error"] = 0
	if strings.Index(check, "ok") == -1 { // check if the reply is what we asked
		log.Printf("Periodic check failed (%q)!\n", check)
		arduino_linear_stat["check_error"] = 1
	}
	lock.Unlock()
	flush_serial()
}

func comm2_arduino(sensor string) (output string) {
	if conf.Serial.Tty == "none" { return }
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
	if conf.Serial.Tty == "none" { return }
	c := &serial.Config{Name: conf.Serial.Tty, Baud: conf.Serial.Baud, ReadTimeout: time.Millisecond * time.Duration(conf.Serial.Timeout)}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 16)
	_, err = s.Read(buf)
	s.Close()
}
