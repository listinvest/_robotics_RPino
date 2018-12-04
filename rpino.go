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

var SerialStat = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "SerialStats",
	Help: "Serial stats",
},
	[]string{"type"})

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
	prometheus.MustRegister(SerialStat)
	gpio1 = make(chan string)
	gpio2 = make(chan string)
	arduino_in = make(chan string)
	arduino_out = make(chan string)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
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
				failed_read++
				if arduino_prev_stat[s] != 0 {
					mutex.Lock()
					arduino_stat[s] = arduino_prev_stat[s]
					mutex.Unlock()
					log.Printf("failed read, using cached value\n")
				}
			} else {
				if arduino_prev_stat[s] == 0 {
					mutex.Lock()
					arduino_stat[s] = output
					mutex.Unlock()
					arduino_prev_stat[s] = output
				} else {
					lower := float32(arduino_prev_stat[s]) * conf.Lower_limit
					upper := float32(arduino_prev_stat[s]) * conf.Upper_limit
					if output <= 10 {
						log.Printf("Saving tiny value (%d)\n",output)
						mutex.Lock()
						arduino_stat[s] = output
						mutex.Unlock()
					} else if  float32(output) >= lower && float32(output) <= upper {
						log.Printf("%s value is within the safe boundaries( %f - %f )\n",s, lower, upper)
						mutex.Lock()
						arduino_stat[s] = output
						mutex.Unlock()
						arduino_prev_stat[s]= output

					} else {
						log.Printf("%s value is outside the safe boundaries( %f - %f ), using cached value %d\n", s, lower, upper, arduino_prev_stat[s])
						mutex.Lock()
						arduino_stat[s] = arduino_prev_stat[s]
						log.Printf("value stored: %d\n", output)
						mutex.Unlock()
					}
				}
			}
		} else {
			if arduino_prev_stat[s] != 0 {
				log.Printf("failed read, using cached value\n")
				mutex.Lock()
				arduino_stat[s] = arduino_prev_stat[s]
				mutex.Unlock()
			}
			failed_read++
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
	SerialStat.WithLabelValues("Good").Add(float64(good_read))
	good_read = 0
	SerialStat.WithLabelValues("Bad").Add(float64(failed_read))
	failed_read = 0
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

	n := len(conf.Arduino_sensors)
	arduino_stat = make(map[string]int,n)
	arduino_prev_stat = make(map[string]int,n)
	for k,_ := range arduino_stat {
		arduino_stat[k]=0
	}
	for k,_ := range arduino_prev_stat {
		arduino_prev_stat[k]=0
	}


	log.Printf("Metrics will be exposed on %s\n", conf.Listen)
	if *verbose {
		log.Printf("Verbose logging is enabled")
	}
	flush_serial()
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)
	go func() {
		for t := range ticker.C {
			get_rpi_stat(*verbose)
			read_arduino()
			time.Sleep(time.Second)
			prometheus_update()
			os.Stderr.WriteString(t.String())
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
