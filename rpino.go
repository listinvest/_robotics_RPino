package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"strconv"
	"strings"
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

var (
	verbose      bool
	arduino_prev_stat map[string]int
	arduino_stat map[string]int
	rpi_stat     map[string]int
	arduino_in   chan (string) // questions to  Arduino
	arduino_out  chan (string) // replies from Arduino
	start_time   time.Time
	good_read int = 1
	failed_read int = 1
	conf *config
)

var mutex = &sync.Mutex{}

func init() {
	prometheus.MustRegister(SensorStat)
	prometheus.MustRegister(RPIStat)
	gpio1 = make(chan string)
	gpio2 = make(chan string)
	arduino_in = make(chan string)
	arduino_out = make(chan string)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	arduino_stat = make(map[string]int)
	arduino_prev_stat = make(map[string]int)
	rpi_stat = make(map[string]int)
}

func read_arduino() {
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	for _, s := range conf.Arduino_sensors {
		log.Printf("sent instruction for: %s", s)
		reply := comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				if arduino_prev_stat[s] != 0 {
					mutex.Lock()
					arduino_stat[s] = arduino_prev_stat[s]
					mutex.Unlock()
					if conf.Zero_unreadable { arduino_prev_stat[s] = 0 }
					log.Printf("failed read, using cached value\n")
				} else {
					log.Printf("failed read, cache value is zero, writing zero\n")
					mutex.Lock()
					arduino_stat[s] = 0
					mutex.Unlock()
				}
			} else {
				lower := float32(arduino_prev_stat[s]) * conf.Lower_limit
				upper := float32(arduino_prev_stat[s]) * conf.Upper_limit
				_,ok := arduino_prev_stat[s]
				if ok && float32(output) < lower {
					log.Printf("%s value is lower than safe boundaries: %f , using cached value\n",s, lower)
					mutex.Lock()
					arduino_stat[s] = arduino_prev_stat[s]
					mutex.Unlock()
					arduino_prev_stat[s]= int(float32(arduino_prev_stat[s])*(conf.Lower_limit/2))
				} else if ok && float32(output) > upper {
					log.Printf("%s value is higher than safe boundaries: %f , using cached value\n",s, upper)
					mutex.Lock()
					arduino_stat[s] = arduino_prev_stat[s]
					mutex.Unlock()
					arduino_prev_stat[s]= int(float32(arduino_prev_stat[s])*(conf.Upper_limit/2))
				} else {
					mutex.Lock()
					arduino_prev_stat[s] = output
					log.Printf("value stored: %d\n", output)
					arduino_stat[s] = output
					mutex.Unlock()
				}
			}
		} else {
			if arduino_prev_stat[s] != 0 {
				mutex.Lock()
				arduino_stat[s] = arduino_prev_stat[s]
				mutex.Unlock()
				if conf.Zero_unreadable { arduino_prev_stat[s] = 0 }
				log.Printf("failed read, using cached value\n")
			} else {
				log.Printf("failed read, cache value is zero, writing zero\n")
				mutex.Lock()
				arduino_stat[s] = 0
				mutex.Unlock()
			}
		}

		time.Sleep(time.Second)
	}
	check := comm2_arduino("S")
	mutex.Lock()
	arduino_stat["check_error"] = 0
	if strings.Index(check,"ok") == -1 { // check if the reply is what we asked
		log.Printf("Periodic check failed (%q)!\n", check)
		arduino_stat["check_error"] = 1
	}
	mutex.Unlock()
}

func get_rpi_stat(verbose bool) {
	if verbose {
		log.Println("RPi stats")
	}
	mutex.Lock()
	rpi_stat["wifi-signal"] = get_wireless_signal()
	d,h := get_uptime()
	rpi_stat["rpi_uptime_days"] = d
	rpi_stat["rpi_uptime_hours"] = h
	rpi_stat["cput"] =  get_Cpu_temp()
	mutex.Unlock()
}

func prometheus_update() {
	mutex.Lock()
	for k, v := range arduino_stat {
		SensorStat.WithLabelValues(k).Set(float64(v))
	}
	for k, v := range rpi_stat {
		RPIStat.WithLabelValues(k).Set(float64(v))
	}
	mutex.Unlock()
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
	log.Printf("Metrics will be exposed on %s\n", conf.Listen)
	if *verbose {
		log.Printf("Verbose logging is enabled")
	}
	flush_serial()
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)
	var t time.Time
	go func() {
		for t = range ticker.C {
			get_rpi_stat(*verbose)
			read_arduino()
			time.Sleep(time.Second)
			prometheus_update()
		}
	}()
	go send_gpio1( gpio1)
	go send_gpio2( gpio2)
	go human_presence()
	go alarm_mgr()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}
