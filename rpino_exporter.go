package main

import (
	"encoding/json"
	//"fmt"
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stianeikeland/go-rpio"
	"log"
	"math"
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
	for _, s := range conf.Arduino_sensors {
		arduino_in <- s
		reply := <- arduino_out
		output,err := strconv.ParseFloat(reply,64)
		if err != nil {
			log.Printf("Failed conversion: %s\n",err)
		}
		log.Printf("value stored: %d",int(output))
		outputf := math.Round(output)
		time.Sleep(time.Millisecond*500)
		mutex.Lock()
			arduino_stat[s] = int(outputf)
		mutex.Unlock()
	}
	arduino_in <- "S" // arduino will blink its built-in led
}

func get_rpi_stat() {
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

func command_socket(w http.ResponseWriter, r *http.Request) {
	socket, ok := r.URL.Query()["s1"]
	if ok {
		if socket[0] == "on" {
			gpio1 <- "on"
			w.Write([]byte("Turning ON"))
		} else if socket[0] == "off" {
			gpio1 <- "off"
			w.Write([]byte("Turning OFF"))
		} else {
			w.Write([]byte("Specify 'on' or 'off'"))
		}
	}
	socket, ok = r.URL.Query()["s2"]
	if ok {
		if socket[0] == "on" {
			gpio2 <- "on"
			w.Write([]byte("Turning ON"))
		} else if socket[0] == "off" {
			gpio2 <- "off"
			w.Write([]byte("Turning OFF"))
		} else {
			w.Write([]byte("Specify 'on' or 'off'"))
		}
	}
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
	confPath := flag.String("conf", "cfg.cfg", "Configuration file")
	verbose := flag.Bool("verbose", false, "Enable logging")
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
			read_arduino(conf)
			get_rpi_stat()
			time.Sleep(time.Second)
			prometheus_update()
		}
	}()
	go send_gpio1(conf,gpio1)
	go send_gpio2(conf,gpio2)
	go comm_arduino()

	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/socket", command_socket)
	http.HandleFunc("/json", json_stats)
	http.HandleFunc("/main", mainpage)
	log.Fatal(http.ListenAndServe(conf.Listen, nil))

}
