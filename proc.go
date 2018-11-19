package main

import (
        "fmt"
	"io/ioutil"
	"math"
	"strings"
	"strconv"
)


func get_uptime() (days,hours int){
	p := "/proc/uptime"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
            fmt.Println("missing")
	}
	fields := strings.Fields(string(stats))
	uptime,_ := strconv.ParseFloat(fields[0],32)
	seconds := int(math.Round(uptime))
	days = seconds/86400
	hours = (seconds%86400)/3600
	//fmt.Printf("days:%d hours:%d\n",days,hours)
	return days,hours
}

func get_wireless_signal() (w int){
	p := "/proc/net/wireless"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
            fmt.Println("missing")
	}
	fields := strings.Fields(string(stats))
	w,_ = strconv.Atoi(strings.TrimSuffix(fields[29],"."))
	//fmt.Printf("wifi: %d \n",w)
	return w
}

func get_Cpu_temp() (t int){

	p := "/sys/class/thermal/thermal_zone0/temp"
	stats, missing := ioutil.ReadFile(p)
	if missing != nil {
            fmt.Println("missing")
	}
        temp := strings.TrimSpace(string(stats))
	//t, missing = strconv.ParseInt(temp,10,32)
	t, missing = strconv.Atoi(temp)
	t /= 1000
	if missing != nil {
            fmt.Println(missing)
	}
	//fmt.Printf("CPU T: %d\n",int(t))
	return int(t)
}
