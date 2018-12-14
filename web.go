package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"strconv"
	"time"
)

func json_stats(w http.ResponseWriter, r *http.Request) {
	all_data := make(map[string]int)
	lock.Lock()
	for k, v := range arduino_linear_stat {
		all_data[k] = v
	}
	for k, v := range arduino_exp_stat {
		all_data[k] = v
	}
	for k, v := range rpi_stat {
		all_data[k] = v
	}
	lock.Unlock()
	//add extra diagnostic fields
	t := time.Now()
	elapsed := t.Sub(start_time)
	hours := int(elapsed.Hours())%24
	days := int(elapsed.Hours())/24
	all_data["rpino_uptime_days"]=days
	all_data["rpino_uptime_hours"]=hours
	msg, _ := json.Marshal(all_data)
	w.Write(msg)
}

func api_router(w http.ResponseWriter, r *http.Request) {
	api_type := r.URL.Path
	switch api_type {
	case "/api/socket":
		socket, ok := r.URL.Query()["s1"]
		if ok {
			if socket[0] != "" {
				reply := command_socket(socket[0])
				w.Write([]byte(reply))
			}
		}
		socket2, ok := r.URL.Query()["s2"]
		if ok {
			if socket2[0] != "" {
				reply := command_socket(socket2[0])
				w.Write([]byte(reply))
			}
		}

	case "/api/arduino_reset":
		comm2_arduino("X")
		w.Write([]byte("ok"))

	case "/api/alarm_test":
		alarm_mgr()
		w.Write([]byte("ok"))

	case "/api/history_reset":
		history_setup()
		w.Write([]byte("ok"))

	case "/api/view_data":
		w.Write([]byte(view_history()))

	case "/api/help":
		w.Write([]byte("Available APIs:\n /socket\n/arduino_reset\n/alarm_test\n/history_reset\n/view_history\n"))

	default:
		log.Printf("Unknown Api (%s)!\n", api_type)
		w.Write([]byte("Unknown Api"))
	}
}

func view_history() (reply string) {
	reply = ""
	for _, sensor := range conf.Arduino_linear_sensors {
		reply = reply + sensor + ": actual= " + strconv.Itoa(arduino_linear_stat[sensor]) + ", prev: "
		for _,v := range arduino_prev_linear_stat[sensor]{
			reply = reply  + strconv.Itoa(v) + ", "
		}
		reply = reply + "\n"
	}
	for _, sensor := range conf.Arduino_exp_sensors {
		reply = reply + sensor + ": actual= " + strconv.Itoa(arduino_exp_stat[sensor]) + ", prev: "
		for _,v := range arduino_prev_exp_stat[sensor]{
			reply = reply  + strconv.Itoa(v) + ", "
		}
		used := strconv.Itoa(arduino_cache_stat[sensor])
		reply = reply + "; used " + used + " times\n"
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
         <p><a href='/api'><b>API endpoint</b></a></p>
         </body>
         </html>
         `))

}

