package main

import (
	"fmt"
	"sort"
)

func nsamples(sensor string)(num int){
	num = len(arduino_prev_linear_stat[sensor])
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


