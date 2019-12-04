package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
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

func get_cpu_temp() (t int) {
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
	//fmt.Printf("Entropy: %d\n", int(r))
	return int(r)
}

func getCPUSample() (idle, total uint64) {
	contents, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	lines := strings.Split(string(contents), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == "cpu" {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			//fmt.Printf("total %d, idle %d\n",total, idle)
			return
		}
	}
	return
}

func get_cpu_usage() {
	Cticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * time.Second)
	for range Cticker.C {
		idle0, total0 := getCPUSample()
		time.Sleep(3 * time.Second)
		idle1, total1 := getCPUSample()
		idleTicks := float64(idle1 - idle0)
		totalTicks := float64(total1 - total0)
		lock.Lock()
		cpu_load = int(100 * (totalTicks - idleTicks) / totalTicks)
		lock.Unlock()
		//fmt.Printf("CPU usage is %f%% [busy: %f, total: %f]\n", cpuload, totalTicks-idleTicks, totalTicks)
	}

}

func get_git_info() (commit string) {
	p := ".git/refs/heads/master"
	contents, missing := ioutil.ReadFile(p)
	if missing != nil {
		fmt.Println("git info missing")
	}
	commit = string(contents)
	return commit
}
