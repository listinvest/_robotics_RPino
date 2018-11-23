package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"
)

func json_stats(w http.ResponseWriter, r *http.Request) {
	all_data := make(map[string]int)
	for k, v := range arduino_stat {
		all_data[k] = v
	}
	for k, v := range rpi_stat {
		all_data[k] = v
	}
	//add extra diagnostic fields
	t := time.Now()
	elapsed := t.Sub(start_time)
	hours := int(elapsed.Hours())%24
	days := int(elapsed.Hours())/24
	all_data["failed_serial_read"]=failed_read
	all_data["good_serial_read"]=good_read
	all_data["quality_serial"]=(failed_read*100)/good_read
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

	case "/api/speak":
		speak()

	default:
		log.Printf("Unknown Api (%s)!\n", api_type)
		w.Write([]byte("Unknown Api"))
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
         <p><a href='/api'><b>API endpoint</b></a></p>
         </body>
         </html>
         `))

}

