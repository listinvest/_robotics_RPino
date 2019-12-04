package main

import (
	"log"
)

func nsamples(sensor string) (num int) {
	num = len(arduino_prev_exp_stat[sensor])
	return num
}

func last_linear(sensor string) (num int) {
	lenght := len(arduino_prev_linear_stat[sensor]) - 1
	num = arduino_prev_linear_stat[sensor][lenght]
	return num
}

func last_exp(sensor string) (num int) {
	lenght := len(arduino_prev_exp_stat[sensor]) - 1
	num = arduino_prev_exp_stat[sensor][lenght]
	return num
}

func add_linear(sensor string, value int) {
	lenght := len(arduino_prev_linear_stat[sensor])
	arduino_prev_linear_stat[sensor] = append(arduino_prev_linear_stat[sensor], value)
	//removing oldest value
	if lenght >= conf.Analysis.Depth {
		arduino_prev_linear_stat[sensor] = arduino_prev_linear_stat[sensor][1:]
	}

}

func add_exp(sensor string, value int) {
	lenght := len(arduino_prev_exp_stat[sensor])
	arduino_prev_exp_stat[sensor] = append(arduino_prev_exp_stat[sensor], value)
	//removing oldest value
	if lenght >= conf.Analysis.Depth {
		arduino_prev_exp_stat[sensor] = arduino_prev_exp_stat[sensor][1:]
	}

}

func dutycycle(sensor string) (up int) {
	up = 0
	num := arduino_linear_stat[sensor]
	cache_count := len(arduino_prev_linear_stat[sensor])
	if cache_count < 2 {
		return up
	}
	prev := arduino_prev_linear_stat[sensor][cache_count-2]
	if conf.Verbose {
		log.Printf("num %d, prev %d, raising: %v\n", num, prev, raising)
	}
	if num >= prev {
		up = 1
		raising = true
	} else {
		raising = false
	}

	if num == prev && raising {
		up = 1
	}
	return up
}

func history_setup() {
	for k := range arduino_linear_stat {
		arduino_linear_stat[k] = 0
	}
	for _, k := range conf.Sensors.Arduino_linear {
		arduino_prev_linear_stat[k] = []int{0}
	}
	for k := range arduino_exp_stat {
		arduino_exp_stat[k] = 0
	}
	for k := range arduino_exp_stat {
		arduino_cache_stat[k] = 0
	}
	for _, k := range conf.Sensors.Arduino_exp {
		arduino_prev_exp_stat[k] = []int{0}
	}
	if conf.Verbose {
		log.Printf("reset history successfull\n")
	}
}
