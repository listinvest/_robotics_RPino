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
	arduino_prev_linear_stat map[string][]int
	arduino_prev_exp_stat map[string][]int
	arduino_linear_stat      map[string]int
	arduino_exp_stat      map[string]int
	rpi_stat          map[string]int
	arduino_in        chan (string) // questions to  Arduino
	arduino_out       chan (string) // replies from Arduino
	start_time        time.Time
	good_read         int = 1
	failed_read       int = 1
	conf              *config
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

}

func read_arduino() {
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	reply := ""
	for _, s := range conf.Arduino_linear_sensors {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				failed_read++
				validated = last_linear(s)
				log.Printf("failed read, using cached value\n")
			} else {
				ref_value := reference(s,output)
				lower := float32(ref_value) * conf.Lower_limit
				upper := float32(ref_value) * conf.Upper_limit
				if float32(output) >= lower && float32(output) <= upper {
					log.Printf("value for %s is %d, within the safe boundaries( %f - %f )\n", s, output, lower, upper)
					validated = output
					add_linear(s,output)

				} else {
					validated = last_linear(s)
					log.Printf("value for %s is %d, which outside the safe boundaries( %f - %f ), using cached value %d\n", s, output, lower, upper,validated)
				}
			}
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_linear(s)
			failed_read++
		}
		reply = ""
		lock.Lock()
		arduino_linear_stat[s] = validated
		lock.Unlock()
		time.Sleep(time.Second * 2)
	}
	for _, s := range conf.Arduino_exp_sensors {
		log.Printf("sent instruction for: %s", s)
		validated := 0
		reply = comm2_arduino(s)
		if reply != "null" {
			output, err := strconv.Atoi(reply)
			if err != nil {
				log.Printf("Failed conversion: %s\n", err)
				failed_read++
				validated = last_exp(s)
				log.Printf("failed read, using cached value\n")
			} else {
				validated = output
				ref_value := last_exp(s)
				lower := float32(ref_value) * (conf.Lower_limit - 0.1)
				upper := float32(ref_value) * (conf.Upper_limit + 0.1)
				if float32(output) >= lower && float32(output) <= upper {
					log.Printf("EXP: value for %s is %d, within the safe boundaries( %f - %f )\n", s, output, lower, upper)
				} else {
					log.Printf("EXP: value for %s is %d, which outside the safe boundaries( %f - %f ), using cached value %d\n", s, output, lower, upper, ref_value)
				}
				add_exp(s,output)
			}
		} else {
			log.Printf("failed read, using cached value\n")
			validated = last_exp(s)
			failed_read++
		}
		reply = ""
		lock.Lock()
		arduino_exp_stat[s] = validated
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
	for k, v := range rpi_stat {
		RPIStat.WithLabelValues(k).Set(float64(v))
	}
	SerialStat.WithLabelValues("Good").Add(float64(good_read))
	good_read = 0
	SerialStat.WithLabelValues("Bad").Add(float64(failed_read))
	failed_read = 0
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
	for k, _ := range arduino_linear_stat {
		arduino_linear_stat[k] = 0
	}

	arduino_prev_linear_stat = make(map[string][]int, n)
	for _, k := range conf.Arduino_linear_sensors {
		arduino_prev_linear_stat[k] = []int{0}
	}

	n = len(conf.Arduino_exp_sensors)
	arduino_exp_stat = make(map[string]int, n)
	for k, _ := range arduino_exp_stat {
		arduino_exp_stat[k] = 0
	}

	arduino_prev_exp_stat = make(map[string][]int, n)
	for _, k := range conf.Arduino_exp_sensors {
		arduino_prev_exp_stat[k] = []int{0}
	}
	log.Printf("Metrics will be exposed on %s\n", conf.Listen)
	if conf.Verbose {
		log.Printf("Verbose logging is enabled")
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
	go start_inputs()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}
