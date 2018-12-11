package main

import (
	"fmt"
	"sort"
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

func median(sensor string, value int) (ref int) {
	lenght := float32(len(arduino_prev_exp_stat[sensor]))
	if lenght == 1 {
		if conf.Verbose { fmt.Printf("history is emtpy, returning %d\n",value)}
		ref = value
		return ref
	}
	sort.Ints(arduino_prev_exp_stat[sensor])
	ref = int(lenght * 0.6)
	if verbose { fmt.Printf("median: %d\n",arduino_prev_linear_stat[sensor][ref]) }
	ref = arduino_prev_exp_stat[sensor][ref]
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
	num := arduino_linear_stat[sensor]
	prev := last_linear(sensor)
        if num >= prev {
                up = 1
        } else if num == prev {
		up = 0
	}
	return up
}
