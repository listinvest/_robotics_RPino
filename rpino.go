package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var SensorStat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "SensorStat",
	Help: "Arduino sensors stats",
},
	[]string{"sensor"})

var RPIStat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: "RPIStat",
	Help: "RPI stats",
},
	[]string{"rpi"})

var SerialStat = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "SerialStats",
	Help: "Serial stats",
},
	[]string{"type"})

var (
	verbose           bool
	raising           bool
	arduino_prev_linear_stat map[string][]int
	arduino_prev_exp_stat map[string][]int
	arduino_linear_stat      map[string]int
	arduino_exp_stat      map[string]int
	arduino_cache_stat      map[string]int
	serial_stat	  map[string]int
	rpi_stat          map[string]int
	arduino_in        chan (string) // questions to  Arduino
	arduino_out       chan (string) // replies from Arduino
	start_time        time.Time
	conf              *config
	first_run	  bool = true
)

var lock = &sync.Mutex{}

func init() {
	prometheus.MustRegister(SensorStat)
	prometheus.MustRegister(RPIStat)
	prometheus.MustRegister(SerialStat)
	gpio1 = make(chan string)
	gpio2 = make(chan string)
	arduino_in = make(chan string)
	arduino_out = make(chan string)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	rpi_stat = make(map[string]int)
	serial_stat = make(map[string]int)
	serial_stat["good_read"] = 1
	serial_stat["failed_read"] = 1
	serial_stat["failed_atoi"] = 1
	serial_stat["failed_interval"] = 1
}

func read_arduino() {
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	reply := ""
	for _, s := range conf.Arduino_linear_sensors {
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
					add_linear(s,output)
				} else {
					validated = last_linear(s)
					log.Printf("value for %s is %d, which outside the safe boundaries( %f - %d - %f ), using cached value %d\n", s, output, lower,ref_value, upper,validated)
					serial_stat["failed_interval"] = serial_stat["failed_interval"] + 1
					arduino_cache_stat[s] = arduino_cache_stat[s] + 1
					if !use_cached {
						log.Printf("Using real value\n")
						validated = output
						add_linear(s,output)
					}
				}
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
	for _, s := range conf.Arduino_exp_sensors {
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
				if float32(output) >= lower && float32(output) <= upper{
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
	first_run = false
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

func get_rpi_stat() {
	if conf.Verbose {
		log.Println("RPi stats")
	}
	lock.Lock()
	rpi_stat["wifi-signal"] = get_wireless_signal()
	d, h := get_uptime()
	rpi_stat["rpi_uptime_days"] = d
	rpi_stat["rpi_uptime_hours"] = h
	rpi_stat["cput"] = get_Cpu_temp()
	lock.Unlock()
}

func prometheus_update() {
	lock.Lock()
	for k, v := range arduino_linear_stat {
		SensorStat.WithLabelValues(k).Set(float64(v))
	}
	for k, v := range arduino_exp_stat {
		SensorStat.WithLabelValues(k).Set(float64(v))
	}
	dutyc := dutycycle("T")
	SensorStat.WithLabelValues("dutycycle_T").Set(float64(dutyc))

	for k, v := range rpi_stat {
		RPIStat.WithLabelValues(k).Set(float64(v))
	}

	for k, v := range serial_stat {
		SerialStat.WithLabelValues(k).Add(float64(v))
		serial_stat[k] = 0
	}
	lock.Unlock()
}

func main() {
	confPath := flag.String("c", "cfg.cfg", "Configuration file")
	verbose := flag.Bool("v", false, "Enable logging")
	flag.Parse()
	start_time = time.Now()
	conf = loadConfig(*confPath)

	if *verbose {
		conf.Verbose = true
	}

	// initialize maps
	n := len(conf.Arduino_linear_sensors)
	arduino_linear_stat = make(map[string]int, n)
	arduino_prev_linear_stat = make(map[string][]int, n)
	n = len(conf.Arduino_exp_sensors)
	arduino_exp_stat = make(map[string]int, n)
	arduino_cache_stat = make(map[string]int, n)
	arduino_prev_exp_stat = make(map[string][]int, n)
	history_setup()

	log.Printf("Prometheus metrics will be exposed on %s\n", conf.Listen)
	if conf.Verbose {
		log.Printf("Verbose logging is enabled")
		if conf.Alarms.Siren_enabled {
			log.Printf("Siren for low temperature %d is configured on pin %d ", conf.Outputs["alarm"].PIN, conf.Alarms.Critical_temp)
		}
		if conf.Alarms.Email_enabled {
			log.Printf("Email notification is for  %s ",conf.Alarms.Mailbox)
		}
	}
	flush_serial()
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)
	go func() {
		for t := range ticker.C {
			get_rpi_stat()
			read_arduino()
			time.Sleep(time.Second)
			prometheus_update()
			os.Stderr.WriteString(t.String())
		}
	}()
	go send_gpio1(gpio1)
	go send_gpio2(gpio2)
	go human_presence()
	go alarm_mgr()
	go siren_mgr()
	go start_inputs()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}
