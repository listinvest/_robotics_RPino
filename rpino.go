package main

import (
	"encoding/json"
	//"fmt"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio"
	"log"
	//"math"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	//"sort"
	"sync"
	//"sync/atomic"
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
	verbose bool
	arduino_stat map[string]int
	rpi_stat   map[string]int
	gpio1      chan (string)
	gpio2      chan (string)
	arduino_in chan (string) // questions to  Arduino
	arduino_out chan (string) // replies from Arduino
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
	rpi_stat = make(map[string]int)
}

func read_arduino(conf *config) {
	if conf.Verbose {
		log.Println("Arduino stats")
	}
	for _, s := range conf.Arduino_sensors {
		log.Printf("sent instruction for: %s",s)
		reply := comm2_arduino(s)
		output,err := strconv.Atoi(reply)
		if err != nil {
			log.Printf("Failed conversion: %s\n",err)
		}
		log.Printf("value stored: %d\n",output)
		mutex.Lock()
			arduino_stat[s] = output
		mutex.Unlock()
		time.Sleep(time.Second)
	}
	check := comm2_arduino("S")
	mutex.Lock()
	arduino_stat["error"]=0
	if check != "ok" {
		log.Printf("Periodic check failed (%q)!\n",check)
		arduino_stat["error"]=1
	}
	if check != "S?" {
		arduino_stat["error"]=1
	}
	mutex.Unlock()
}

func get_rpi_stat(verbose bool) {
	if verbose {
		log.Println("RPi stats")
	}
	cmd_load := "uptime | cut -d ' ' -f 11|cut -d '.' -f 1"
	oneminload, err := exec.Command("bash", "-c", cmd_load).Output()
	if err != nil {
		log.Fatal(err)
	}
	oneminload_rounded, _ := strconv.Atoi(string(oneminload[0]))

	mutex.Lock()
	rpi_stat["wifi-signal"] = rand.Intn(100)
	rpi_stat["1minload"] = oneminload_rounded
	mutex.Unlock()
}

func speak() {
	sermon := "espeak -g 5 \"Please listen to the following stats:\n"
	for k, v := range arduino_stat {
		val := strconv.Itoa(v)
		sermon = sermon + k +" is " + val + "\n"
	}
	sermon = sermon +"\""
	log.Printf("%s\n",sermon)
	_, err := exec.Command("bash", "-c", sermon).Output()
	if err != nil {
		log.Fatal(err)
	}
}

func human_presence() {
        mutex.Lock()
        presence := arduino_stat["P"]
        mutex.Unlock()
	if presence == 1 {
		speak()
	}
        time.Sleep(time.Minute)
}

func alarm_mgr(conf *config) {
        mutex.Lock()
        actual_temp := arduino_stat["T"]
        mutex.Unlock()
        if actual_temp < conf.Critical_temp  {
                log.Printf("Alarm triggered!!\n")
                pin := rpio.Pin(conf.Socket1)
                pin.Output()
                pin.High()
                time.Sleep(time.Second)
                pin.Low()
        }
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

func json_stats(w http.ResponseWriter, r *http.Request) {
	all_data := make(map[string]int)
	for k, v := range arduino_stat {
		all_data[k] = v
	}
	for k, v := range rpi_stat {
		all_data[k] = v
	}
	msg, _ := json.Marshal(all_data)
	w.Write(msg)
}

func api_router(w http.ResponseWriter, r *http.Request) {

	api_type := r.URL.Path
	switch api_type {
	case "/api/socket" :
		socket,ok := r.URL.Query()["s1"]
		if ok {
			if socket[0] != "" {
				reply := command_socket(socket[0])
				w.Write([]byte(reply))
			}
		}
		socket2,ok := r.URL.Query()["s2"]
		if ok {
			if socket2[0] != "" {
				reply := command_socket(socket2[0])
				w.Write([]byte(reply))
			}
		}

	case  "/api/arduino_reset" :
		comm2_arduino("X")

	default :
		log.Printf("Unknown Api (%s)!\n",api_type)
		w.Write([]byte("Unknown Api"))
	}
}

func command_socket(socket string) (reply string){
	if socket == "on" {
		gpio1 <- "on"
		reply = "Turning ON"
	} else if socket == "off" {
		gpio1 <- "off"
		reply = "Turning OFF"
	} else {
		reply = "Specify 'on' or 'off'"
	}
	return reply
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
         <html>
         <head><title>Rpino Exporter</title></head>
         <body>
         <h1>Rpino Exporter</h1>
         <h2>parameters '` + strings.Join(os.Args, " ") + `'</h2>
         <p><a href='/metrics'><b>Prometheus Metrics</b></a></p>
         <p><a href='/json'><b>JSON Metrics</b></a></p>
         <p><a href='/socket'><b>Socket API endpoint</b></a></p>
         </body>
         </html>
         `))

}

func send_gpio1(conf *config, gpio1 <-chan string) {
	pin := rpio.Pin(conf.Socket1)
	pin.Output()
	for {
		status := <-gpio1
		log.Printf("Sending %s to GPIO1", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

func send_gpio2(conf *config, gpio2 <-chan string) {
	pin := rpio.Pin(conf.Socket2)
	pin.Output()
	for {
		status := <-gpio2
		log.Printf("Sending %s to GPIO2", status)
		if status == "on" {
			pin.High()
		}
		if status == "off" {
			pin.Low()
		}
	}
}

func main() {
	confPath := flag.String("c", "cfg.cfg", "Configuration file")
	verbose := flag.Bool("v", false, "Enable logging")
	flag.Parse()

	conf, err := loadConfig(*confPath)
	if err != nil {
		log.Fatalln(err)
	}

	if *verbose {
		conf.Verbose = true
	}
	log.Printf("Metrics will be exposed on %s\n", conf.Listen)
	//set a x seconds ticker
	ticker := time.NewTicker(time.Duration(conf.Poll_interval) * time.Second)

	go func() {
		for t := range ticker.C {
			if *verbose {
				log.Println("\nStats at", t)
			}
			get_rpi_stat(*verbose)
			read_arduino(conf)
			time.Sleep(time.Second)
			prometheus_update()
		}
	}()
	go send_gpio1(conf,gpio1)
	go send_gpio2(conf,gpio2)
	go human_presence()
	go alarm_mgr(conf)

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/api/", api_router)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))

}
