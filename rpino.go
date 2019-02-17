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
	verbose                  bool
	raising                  bool
	arduino_prev_linear_stat map[string][]int
	arduino_prev_exp_stat    map[string][]int
	arduino_linear_stat      map[string]int
	arduino_exp_stat         map[string]int
	arduino_cache_stat       map[string]int
	serial_stat              map[string]int
	rpi_stat                 map[string]int
	arduino_in               chan (string) // questions to  Arduino
	arduino_out              chan (string) // replies from Arduino
	start_time               time.Time
	conf                     *config
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
	adjusted := 0
	temp := false
	for k, v := range arduino_linear_stat {
		if k == "T" {
			temp = true
		}
		if conf.Sensors.Adj_T["value"] != 0 && k == "T" {
			adjusted = v + conf.Sensors.Adj_T["value"]
			SensorStat.WithLabelValues(k).Set(float64(adjusted))
		} else if conf.Sensors.Adj_H["value"] != 0 && k == "H" {
			adjusted = v + conf.Sensors.Adj_H["value"]
			SensorStat.WithLabelValues(k).Set(float64(adjusted))
		} else {
			SensorStat.WithLabelValues(k).Set(float64(v))
		}
	}
	for k, v := range arduino_exp_stat {
		SensorStat.WithLabelValues(k).Set(float64(v))
	}
	if temp {
		dutyc := dutycycle("T")
		SensorStat.WithLabelValues("dutycycle_T").Set(float64(dutyc))
	}
	for k, v := range rpi_stat {
		RPIStat.WithLabelValues(k).Set(float64(v))
	}

	for k, v := range serial_stat {
		SerialStat.WithLabelValues(k).Add(float64(v))
		serial_stat[k] = 0
	}
	if arduino_linear_stat["check_error"] == 1 {
		SerialStat.WithLabelValues("check_error").Add(float64(1))
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

	initialize_arduino()

	log.SetPrefix("[RPino] ")
	log.Printf("Prometheus metrics will be exposed on %s\n", conf.Listen)
	if conf.Verbose {
		log.Printf("Verbose logging is enabled")
		if conf.Alarms.Siren_enabled {
			log.Printf("Siren on pin %d for low temperature set on %d", conf.Outputs["alarm"].PIN, conf.Alarms.Critical_temp)
		}
		if conf.Alarms.Email_enabled {
			log.Printf("Email notification is for: %s ", conf.Alarms.Mailbox)
		}
		log.Printf("Adjustments: H %d, T %d ", conf.Sensors.Adj_H["value"], conf.Sensors.Adj_T["value"])
	}
	flush_serial()
	Mticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * time.Second)
	defer Mticker.Stop()
	go func() {
		for _ = range Mticker.C {
			get_rpi_stat()
			read_arduino()
			if conf.Sensors.Bmp > 0 {
				bmp180()
			}
			if conf.Sensors.Dht > 0 {
				dht11()
			}
			time.Sleep(time.Second)
			prometheus_update()
			light_mgr()
		}
	}()
	go send_gpio1(gpio1)
	go send_gpio2(gpio2)
	go input_presence()
	go alarm_mgr()
	go siren_mgr()
	go start_inputs()
	go get_time()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}
