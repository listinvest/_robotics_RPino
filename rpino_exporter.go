package main
 
import (
		"encoding/json"
                //"fmt"
                "flag"
                "github.com/prometheus/client_golang/prometheus"
                "github.com/prometheus/client_golang/prometheus/promhttp"
                "log"
		"math/rand"
                _ "net/http/pprof"
                "net/http"
                "os"
                "os/exec"
                //"sort"
		"sync"
    		//"sync/atomic"
                "strings"
		"strconv"
                "time"
)

var SensorStat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                	Name: "SensorStat",
                	Help: "Arduino sensors stats",
		},
                []string{"sensor"},)
 
var RPIStat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
                	Name: "RPIStat",
                	Help: "RPI stats",
		},
                []string{"rpi"},)
 
var (
        verbose        bool
        //elements       int
	arduino_in  map[string]int	
	rpi_stat  map[string]int	
	gpio1	chan(string)
	gpio2	chan(string)
)


var mutex = &sync.Mutex{}
 
func init() {
                prometheus.MustRegister(SensorStat)
                prometheus.MustRegister(RPIStat)
		gpio1 = make(chan string)
		gpio2 = make(chan string)
		arduino_in = make(map[string]int)
		rpi_stat = make(map[string]int)
}

func read_arduino(conf *config) {
        mutex.Lock()
	for _,s := range conf.Arduino_sensors {	
        	arduino_in[s] = rand.Intn(50)
	}
        mutex.Unlock()
}


func get_rpi_stat() {
	cmd_load := "uptime | cut -d ' ' -f 11|cut -d '.' -f 1"
	oneminload, err := exec.Command("bash","-c",cmd_load).Output()
	if err != nil {
		log.Fatal(err)
	}
	oneminload_rounded,_ :=  strconv.Atoi(string(oneminload[0])) 

        mutex.Lock()
        rpi_stat["wifi-signal"] = rand.Intn(100)
        rpi_stat["1minload"] = oneminload_rounded
        mutex.Unlock()
}
 
func prometheus_update() {
        mutex.Lock()
	for k,v := range arduino_in {
        	SensorStat.WithLabelValues(k).Set(float64(v))
	}
	for k,v := range rpi_stat {
        	RPIStat.WithLabelValues(k).Set(float64(v))
	}
        mutex.Unlock()
}

func json_stats(w http.ResponseWriter, r *http.Request) { 
        all_data := make(map[string]int)	
	for k,v := range arduino_in {
		all_data[k]=v
	}
	for k,v := range rpi_stat {
		all_data[k]=v
	}
	msg,_ := json.Marshal(all_data)
	w.Write(msg)
}

func command_socket(w http.ResponseWriter, r *http.Request) { 
	socket,ok := r.URL.Query()["s1"]
	if ok {
		if socket[0] == "on" {
			gpio1 <- "on"
			w.Write([]byte("Turning ON"))
		}else if socket[0] == "off" {
			gpio1 <- "off"
			w.Write([]byte("Turning OFF"))
		} else {
			w.Write([]byte("Specify 'on' or 'off'"))
		}
	}
	socket,ok = r.URL.Query()["s2"]
	if ok {
		if socket[0] == "on" {
			gpio2 <- "on"
			w.Write([]byte("Turning ON"))
		}else if socket[0] == "off" {
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


func send_gpio1(gpio1 <-chan string) {
	for {
		status := <-gpio1
		log.Printf("Sending %s to GPIO1", status)
	}
}

func send_gpio2(gpio2 <-chan string) {
	for {
		status := <-gpio2
		log.Printf("Sending %s to GPIO2", status)
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
				go send_gpio1(gpio1)
				go send_gpio2(gpio2)
 
                http.Handle("/metrics", promhttp.Handler())
                http.HandleFunc("/socket", command_socket)
                http.HandleFunc("/json", json_stats)
                http.HandleFunc("/main", mainpage) 
          	log.Fatal(http.ListenAndServe(conf.Listen, nil))
 
}

