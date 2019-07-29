package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
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

var (
	verbose                  bool
	raising                  bool
	arduino_connected        bool
	arduino_comm_time        float64
	clock_offset             int
	cpu_load		 int
	iterations		 int64
	logfile                  string
	git_info		 string
	arduino_prev_linear_stat map[string][]int
	arduino_prev_exp_stat    map[string][]int
	arduino_linear_stat      map[string]int
	arduino_exp_stat         map[string]int
	arduino_cache_stat       map[string]int
	sensor_stat              map[string]int
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
	gpio1 = make(chan string)
	gpio2 = make(chan string)
	arduino_in = make(chan string)
	arduino_out = make(chan string)

	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	rpi_stat = make(map[string]int)
	sensor_stat = make(map[string]int, 10)
}

func get_rpi_stat() {
	if conf.Verbose {
		log.Println("RPi stats")
	}
	lock.Lock()
	wifi := get_wireless_signal()
	if wifi > 0 {
		rpi_stat["wifi-signal"] = get_wireless_signal()
	}
	d, h := get_uptime()
	rpi_stat["rpi_uptime_days"] = d
	rpi_stat["rpi_uptime_hours"] = h
	rpi_stat["cput"] = get_cpu_temp()
	rpi_stat["cpu_load"] = cpu_load
	rpi_stat["clock_offset"] = clock_offset
	rpi_stat["entropy"] = get_entropy()
	if conf.Serial.Tty != "none" {
		rpi_stat["arduino_present"] = 1
		if arduino_connected {
			rpi_stat["arduino_connected"] = 1
		} else {
			rpi_stat["arduino_connected"] = 0
		}
	} else {
		rpi_stat["arduino_present"] = 0
	}
	lock.Unlock()
}

func prometheus_update() {
	lock.Lock()
	adjusted := 0
	temp := false
	for k, v := range sensor_stat {
		SensorStat.WithLabelValues(k).Set(float64(v))
	}
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
	if conf.Serial.Tty != "none" {
		RPIStat.WithLabelValues("arduino_comm_time").Set(arduino_comm_time)
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
	RPIStat.WithLabelValues("iterations").Set(float64(iterations))
	iterations++
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
	git_info = get_git_info()
	initialize_arduino()
	flush_serial()
	iterations = 0
	f, err := os.OpenFile(conf.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file")
	}
	defer f.Close()
	log.SetOutput(f)
	log.SetPrefix("[RPino] ")
	log.Printf("Git version: %s\n", get_git_info())
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
		if conf.Serial.Tty != "none" {
			log.Printf("Arduino connected on port: %s ", conf.Serial.Tty)
		}
	}
	Mticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * time.Second)
	defer Mticker.Stop()
	go func() {
		for range Mticker.C {
			read_arduino()
			get_rpi_stat()
			if conf.Sensors.Bmp {
				bmp180()
			}
			if conf.Sensors.Dht {
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
	go water_mgr()
	go get_cpu_usage()
	if conf.Sensors.Sds11 { go sds11() }

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/notify", PostHandler)
	http.HandleFunc("/main", mainpage)
	fmt.Printf("Rpino is up\n")
	log.Fatal(http.ListenAndServe(conf.Listen, nil))
}
