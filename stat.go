package main

import (
	"fmt"
	"sort"
)

/*
var (
	verbose      bool
	percentile   float32
	depth	     int
	arduino_prev_stat map[string][]int
	Arduino_sensors = []string{"a","d","c"}
)

func init() {
	verbose = true
	percentile = 0.6
	depth = 3
	n := len(Arduino_sensors)

	arduino_prev_stat = make(map[string][]int,n)

	for _,k := range Arduino_sensors {
		arduino_prev_stat[k]=[]int{0}
	}
	
}


func main() {

	fmt.Printf("initial: %v\n",arduino_prev_stat)
	add("a",10)
	add("a",41)
	s := nsamples("a")
	fmt.Printf("elements: %d\n",s)
	add("a",31)
	l := last("a")
	fmt.Printf("last: %d\n",l)
	add("a",22)
	fmt.Printf("final: %v\n",arduino_prev_stat)
	kk := reference("a")
	fmt.Printf("ref: %d\n",kk)
}
*/

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
	if lenght >= conf.Depth {
		arduino_prev_linear_stat[sensor] = arduino_prev_linear_stat[sensor][1:]
	}

}

func add_exp(sensor string, value int) {
	lenght := len(arduino_prev_exp_stat[sensor])
	arduino_prev_exp_stat[sensor] = append(arduino_prev_exp_stat[sensor],value)
	//removing oldest value
	if lenght >= conf.Depth {
		arduino_prev_exp_stat[sensor] = arduino_prev_exp_stat[sensor][1:]
	}

}

func reference(sensor string) (ref int) {
	lenght := float32(len(arduino_prev_linear_stat[sensor]))
	sort.Ints(arduino_prev_linear_stat[sensor])
	ref = int(lenght * conf.Percentile)
	if verbose { fmt.Printf("index %d, value: %d\n",ref,arduino_prev_linear_stat[sensor][ref]) }
	return arduino_prev_linear_stat[sensor][ref]
}


