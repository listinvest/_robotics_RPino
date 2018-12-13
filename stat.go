package main

import (
	"fmt"
	"sort"
)

var (
	raising bool
)

func nsamples(sensor string)(num int){
	num = len(arduino_prev_exp_stat[sensor])
	return num
}

func last_linear(sensor string)(num int){
	lenght := len(arduino_prev_linear_stat[sensor])-1
	num = arduino_prev_linear_stat[sensor][lenght]
	return num
}

func last_exp(sensor string)(num int){
	lenght := len(arduino_prev_exp_stat[sensor])-1
	num = arduino_prev_exp_stat[sensor][lenght]
	return num
}


func add_linear(sensor string, value int) {
	lenght := len(arduino_prev_linear_stat[sensor])
	arduino_prev_linear_stat[sensor] = append(arduino_prev_linear_stat[sensor],value)
	//removing oldest value
	if lenght >= conf.Analysis.Depth {
		arduino_prev_linear_stat[sensor] = arduino_prev_linear_stat[sensor][1:]
	}

}

func add_exp(sensor string, value int) {
	lenght := len(arduino_prev_exp_stat[sensor])
	arduino_prev_exp_stat[sensor] = append(arduino_prev_exp_stat[sensor],value)
	//removing oldest value
	if lenght >= conf.Analysis.Depth {
		arduino_prev_exp_stat[sensor] = arduino_prev_exp_stat[sensor][1:]
	}

}

func reference(sensor string, value int) (ref int) {
	lenght := float32(len(arduino_prev_linear_stat[sensor]))
	if lenght == 1 {
		if conf.Verbose { fmt.Printf("history is emtpy, returning %d\n",value)}
		ref = value
		return ref
	}
	sort.Ints(arduino_prev_linear_stat[sensor])
	ref = int(lenght * conf.Analysis.Percentile)
	if verbose { fmt.Printf("index %d, value: %d\n",ref,arduino_prev_linear_stat[sensor][ref]) }
	ref = arduino_prev_linear_stat[sensor][ref]
	return ref
}

// multiplied moving average
func mma(sensor string, value int, ff int, sf int) (avg float32) {
        lenght := len(arduino_prev_exp_stat[sensor])
	if lenght <= 1 {
		if conf.Verbose { fmt.Printf("history is emtpy, returning %d\n",value)}
		avg = float32(value)
		return avg
	}
        items := 0
        total := 0

        for i,v := range arduino_prev_exp_stat[sensor]{
		if v == 0 { continue }
                if i == lenght - 2{
                        total = total + v*sf
                        items = items + sf
                }
                if i == lenght -1 {
                        total = total + v*ff
                        items = items + ff
                }
                total = total + v
                items = items + 1
        }
        avg = float32(total)/float32(items)
	return avg
}

func dutycycle(sensor string) (up int) {
	up = 0
	num := arduino_linear_stat[sensor]
	cache_count:= len(arduino_prev_linear_stat[sensor])
	if cache_count < 2 { return up }
	prev := arduino_prev_linear_stat[sensor][cache_count-2]
        if num > prev {
                up = 1
		raising = true
        } else if num == prev && raising {
                up = 1
	} else {
		raising = false
	}
	if conf.Verbose { fmt.Printf("rampup status: %d, raising: %v\n",up,raising)}
	return up
}

func history_setup() {
	for k, _ := range arduino_linear_stat {
		arduino_linear_stat[k] = 0
	}
	for _, k := range conf.Arduino_linear_sensors {
		arduino_prev_linear_stat[k] = []int{0}
	}
	for k, _ := range arduino_exp_stat {
		arduino_exp_stat[k] = 0
	}
	for k, _ := range arduino_exp_stat {
		arduino_cache_stat[k] = 0
	}
	for _, k := range conf.Arduino_exp_sensors {
		arduino_prev_exp_stat[k] = []int{0}
	}
	if conf.Verbose { fmt.Printf("reset history successfull\n")}
}
