package main

import (
	"github.com/d2r2/go-bsbmp"
	dht "github.com/d2r2/go-dht"
	"github.com/d2r2/go-i2c"
	"github.com/makers-bierzo/sds011"
	"github.com/stianeikeland/go-rpio"
	"github.com/tarm/serial"
	"log"
	"time"
)

var (
	first_time = true
	prev_temp  = float32(0)
)

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

	for range Gticker.C {
		pin.Toggle()                     //flick
		time.Sleep(5 * time.Millisecond) //wait
		res := pin.Read()                //read
		//log.Printf("detected: %d",res)
		lock.Lock()
		sensor_stat[sensor] = int(res)
		lock.Unlock()
		if res == 1 {
			input <- true
		}
	}
}

func bmp180() {
	if conf.Verbose {
		log.Println("Reading BMP180")
	}
	// Use 'i2cdetect -y 1' utility to find the device address
	i2c, err := i2c.NewI2C(uint8(0x77), 1)
	if err != nil {
		log.Println(err)
		return
	}
	defer i2c.Close()
	sensor, err := bsbmp.NewBMP(bsbmp.BMP180, i2c)
	if err != nil {
		log.Println(err)
		return
	}
	// Read temperature in celsius degree
	t, err := sensor.ReadTemperatureC(bsbmp.ACCURACY_STANDARD)
	if err != nil {
		log.Println(err)
		return
	}
	if conf.Verbose {
		log.Printf("Temperature = %f*C", t)
	}
	// Read atmospheric pressure in millibar
	p, err := sensor.ReadPressurePa(bsbmp.ACCURACY_STANDARD)
	if err != nil {
		log.Println(err)
		return
	}
	if conf.Verbose {
		log.Printf("Pressure = %f millibar", p/100)
	}
	lock.Lock()
	sensor_stat["bmp180_T"] = int(t)
	sensor_stat["bmp180_P"] = int(p / 100)
	lock.Unlock()

}

func dht11() {
	if conf.Verbose {
		log.Println("Reading DHT11")
	}
	temperature, humidity, _, err := dht.ReadDHTxxWithRetry(dht.DHT11, conf.Sensors.Dht_pin, true, 3)
	if err != nil {
		log.Println(err)
		return
	}
	if conf.Verbose {
		log.Printf("Temperature = %f*C, Humidity = %f%%", temperature, humidity)
	}
	lock.Lock()
	if temperature > (float32(prev_temp)*conf.Analysis.Lower_limit) && temperature <= (float32(prev_temp)*conf.Analysis.Upper_limit) {
		sensor_stat["dht_T"] = int(temperature)
		prev_temp = float32(temperature)
	} else {
		sensor_stat["dht_T"] = int(prev_temp)
		if conf.Verbose {
			log.Printf("Temperature outside boundaries (centered on %f)", prev_temp)
		}
	}
	if first_time {
		prev_temp = temperature
		first_time = false
	}
	// humidity seems to not suffer from false readings
	sensor_stat["dht_H"] = int(humidity)
	lock.Unlock()
}

func sds11() {
	if conf.Verbose {
		log.Println("Reading SDS11")
	}
	c := &serial.Config{Name: "/dev/ttyUSB0", Baud: 9600}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	sensor := sds011.NewSensor(s)
	if err != nil {
		log.Printf("1: %s", err)
		return
	}

	err = sensor.Sleep(false)
	if err != nil {
		log.Printf("2: %s", err)
		return
	}
	err = sensor.SetWorkingPeriod(1)
	if err != nil {
		log.Printf("3: %s", err)
		return
	}
	err = sensor.SetMode(sds011.ActiveMode)
	if err != nil {
		log.Printf("4: %s", err)
		return
	}

	measureChannel := make(chan sds011.Measurement)
	sensor.OnQuery(measureChannel)

	sensor.Listen()
	for true {
		measure := <-measureChannel
		lock.Lock()
		sensor_stat["pm2"] = int(measure.PM2_5)
		sensor_stat["pm10"] = int(measure.PM10)
		lock.Unlock()
		if conf.Verbose {
			log.Printf("[%s]\nPM 2.5 => %f μg/m³\nPM 10 => %f μg/m³\n", time.Now().Format("2006-01-02 15:04:05"), measure.PM2_5, measure.PM10)
		}
	}
}
