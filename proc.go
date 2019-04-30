package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

func get_uptime() (days, hours int) {
	p := "/proc/uptime"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
		fmt.Println("missing")
	}
	fields := strings.Fields(string(stats))
	uptime, _ := strconv.ParseFloat(fields[0], 32)
	seconds := int(uptime)
	days = seconds / 86400
	hours = (seconds % 86400) / 3600
	//fmt.Printf("days:%d hours:%d\n",days,hours)
	return days, hours
}

func get_wireless_signal() (w int) {
	p := "/proc/net/wireless"
	stats, missing := ioutil.ReadFile(p)
	w = 0
	if missing != nil {
		fmt.Println("missing")
	}
	fields := strings.Fields(string(stats))
	if len(fields) > 29 {
		w, _ = strconv.Atoi(strings.TrimSuffix(fields[29], "."))
	}
	return w
}

func get_Cpu_temp() (t int) {
	p := "/sys/class/thermal/thermal_zone0/temp"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
		fmt.Println("missing")
	}
	temp := strings.TrimSpace(string(stats))
	t, missing = strconv.Atoi(temp)
	t /= 1000
	if missing != nil {
		fmt.Println(missing)
	}
	//fmt.Printf("CPU T: %d\n",int(t))
	return int(t)
}

func get_entropy() (r int) {
	p := "/proc/sys/kernel/random/entropy_avail"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
		fmt.Println("missing")
	}
	ent := strings.TrimSpace(string(stats))
	r, missing = strconv.Atoi(ent)
	if missing != nil {
		fmt.Println(missing)
	}
	fmt.Printf("Entropy: %d\n", int(r))
	return int(r)
}
