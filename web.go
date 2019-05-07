package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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
	if arduino_linear_stat["T"] < conf.Alarms.Critical_temp {
		all_data["siren"] = 1
	} else {
		all_data["siren"] = 0
	}
	lock.Unlock()
	//add extra diagnostic fields
	t := time.Now()
	elapsed := t.Sub(start_time)
	hours := int(elapsed.Hours()) % 24
	days := int(elapsed.Hours()) / 24
	all_data["rpino_uptime_days"] = days
	all_data["rpino_uptime_hours"] = hours
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

	case "/api/arduino_reset":
		comm2_arduino("X")
		w.Write([]byte("ok"))

	case "/api/alarm_test":
		test_siren()
		w.Write([]byte("ok"))

	case "/api/history_reset":
		history_setup()
		w.Write([]byte("ok"))

	case "/api/view_data":
		w.Write([]byte(view_data()))

	case "/api/view_conf":
		w.Write([]byte(view_conf()))

	case "/api/help":
		w.Write([]byte("Available APIs:\n/socket\n/arduino_reset\n/alarm_test\n/history_reset\n/view_data\n"))

	default:
		log.Printf("Unknown Api (%s)!\n", api_type)
		w.Write([]byte("Unknown Api"))
	}
}

func view_data() (reply string) {
	reply = "\nLinear sensors:\n"
	for _, sensor := range conf.Sensors.Arduino_linear {
		reply = reply + sensor + ": actual= " + strconv.Itoa(arduino_linear_stat[sensor]) + ", prev: "
		for _, v := range arduino_prev_linear_stat[sensor] {
			reply = reply + strconv.Itoa(v) + ", "
			reply = reply + "\n "
		}
	}
	reply = reply + "\n\nExponential sensors:\n"
	for _, sensor := range conf.Sensors.Arduino_exp {
		last := len(arduino_prev_exp_stat[sensor]) - 1
		reply = reply + sensor + ": actual= " + strconv.Itoa(arduino_prev_exp_stat[sensor][last]) + ", prev: "
		for _, v := range arduino_prev_exp_stat[sensor] {
			reply = reply + strconv.Itoa(v) + ", "
		}
		reply = reply + "\n "
	}

	return reply
}

func view_conf() (reply string) {
	reply = fmt.Sprintf("%q", conf.Sensors.Arduino_linear)
	reply = reply + "\n" + fmt.Sprintf("%q", conf.Sensors.Arduino_exp)
	return reply
}

// PostHandler converts post request body to string
func PostHandler(w http.ResponseWriter, r *http.Request) {
	type Msg struct {
		Alarm bool
	}
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		if !json.Valid(body) {
			var json_message Msg
			json.Unmarshal(body, json_message)
			log.Printf("good %s", json_message.Alarm)
		} else {
			fmt.Fprint(w, "invalid JSON received by POST")
		}

	} else {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

func mainpage(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	w.Write([]byte(`
         <html>
         <head><title>RPino</title></head>
         <body>
         <h1>Rpino Web Interface running on ` + hostname + `</h1>
         <h2>parameters '` + strings.Join(os.Args, " ") + `'</h2>
         <p><a href='/metrics'><b>Prometheus Metrics</b></a></p>
         <p><a href='/json'><b>JSON Metrics</b></a></p>
         <p><a href='/api/help'><b>API endpoint</b></a></p>
         <p><a href='/debug/pprof/'><b>PProf endpoint</b></a></p>
         </body>
         </html>
         `))

}
