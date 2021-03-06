package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	hostname string
)

func init() {
	hostname, _ = os.Hostname()
}
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
	ver, _  := strconv.Atoi(version)
	all_data["arduino_version"] = ver
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
		initialize_arduino()
		w.Write([]byte("ok"))

	case "/api/view_data":
		w.Write([]byte(view_data()))

	case "/api/wifi_stats":
		w.Write([]byte(view_wifi()))

	case "/api/view_conf":
		w.Write([]byte(view_conf()))

	case "/api/help":
		w.Write([]byte("<html><body><h1>Available APIs:</h1><br><a href='/api/socket'>/socket</a><br><a href='/api/arduino_reset'>/arduino_reset</a><br><a href='/api/alarm_test'>/alarm_test</a><br><a href='/api/history_reset'>/history_reset</a><br><a href='/api/view_data'>/view_data</a><br><a href='/api/view_conf'>/view_conf</a><br><a href='/api/wifi_stats'>/WiFi stats</a></body></html>"))

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
	raw,_ := ioutil.ReadFile(confPath)
	return string(raw)
}

func view_wifi() (reply string) {
	raw, err := exec.Command("iwlist", "wlan0", "scanning").Output()
	log.Printf("Cannot run iwlist command: %s\n", err)
	return string(raw)
}

// PostHandler converts post request body to string
func PostHandler(w http.ResponseWriter, r *http.Request) {
	type Msg struct {
		Siren bool
	}
	if r.Method == "POST" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body",
				http.StatusInternalServerError)
		}
		var json_message Msg
		err = json.Unmarshal(body, &json_message)
		if err == nil {
			log.Printf("received a good json %t", json_message.Siren)
		} else {
			fmt.Fprintf(w, "invalid JSON - %s", err)
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
         <h2>git commit '` + git_info + `'</h2>

         <p><a href='/metrics'><b>Prometheus Metrics</b></a></p>
         <p><a href='/json'><b>JSON Metrics</b></a></p>
         <p><a href='/api/help'><b>API endpoint</b></a></p>
         <p><a href='/debug/pprof/'><b>PProf endpoint</b></a></p>
         </body>
         </html>
         `))

}
