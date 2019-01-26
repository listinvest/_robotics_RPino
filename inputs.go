package main

import (
	"github.com/stianeikeland/go-rpio"
	"log"
	"time"
	"github.com/d2r2/go-bsbmp"
	"github.com/d2r2/go-i2c"
	dht "github.com/d2r2/go-dht"
)

//var lock = &sync.Mutex{}

func start_inputs() {
	if len(conf.Inputs) == 0 {
		log.Printf("No GPIO to monitor")
	}
	for sensor, detail := range conf.Inputs {
		if detail.PIN != 0 {
			log.Printf("Starting to monitor sensor: %s on pin %d", sensor, detail.PIN)
			go gpio_watch(sensor, detail.PIN)
		}
	}
}

func gpio_watch(sensor string, Spin int) {
	// Open and map memory to access gpio, check for errors
	pin := rpio.Pin(Spin)
	if err := rpio.Open(); err != nil {
		log.Fatal(err)
	}
	pin.Input()
	defer rpio.Close()
	//set a x seconds ticker
	Gticker := time.NewTicker(time.Duration(conf.Sensors.Poll_interval) * time.Second / 2)
	defer Gticker.Stop()

	for _ = range Gticker.C {
		res := pin.Read()
		//log.Printf("detected: %d",res)
		lock.Lock()
		rpi_stat[sensor] = int(res)
		lock.Unlock()
		if res == 1 {
			input <- true
		}
	}
}


func bmp180() {
        if conf.Verbose { log.Println("Reading BMP180") }
        // Use 'i2cdetect -y 1' utility to find the device address 
        i2c, err := i2c.NewI2C(0x77, 1)
        if err != nil {
                log.Println(err)
        }
        defer i2c.Close()
        sensor, err := bsbmp.NewBMP(bsbmp.BMP180, i2c)
        if err != nil {
                log.Println(err)
        }
        // Read temperature in celsius degree
        t, err := sensor.ReadTemperatureC(bsbmp.ACCURACY_STANDARD)
        if err != nil {
		log.Println(err)
        }
        if conf.Verbose {
		log.Printf("Temperature = %f*C", t)
        }
	// Read atmospheric pressure in millibar
        p, err := sensor.ReadPressurePa(bsbmp.ACCURACY_STANDARD)
        if err != nil {
                log.Println(err)
        }
        if conf.Verbose {
		log.Printf("Pressure = %f millibar", p/100 )
	}
	lock.Lock()
	rpi_stat["bmp180_T"] = int(t)
	rpi_stat["bmp180_P"] = int(p/100)
	lock.Unlock()
}

func dht11() {
        if conf.Verbose { log.Println("Reading DHT11") }
	temperature, humidity, _, err :=
		dht.ReadDHTxxWithRetry(dht.DHT11, 14, false, 2)
	if err != nil {
		log.Println(err)
	}
        if conf.Verbose {
		log.Printf("Temperature = %f*C, Humidity = %f%%", temperature, humidity)
        }
	lock.Lock()
	rpi_stat["dht_T"] = int(temperature)
	rpi_stat["dht_H"] = int(humidity)
	lock.Unlock()
}

