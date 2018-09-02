// +build ignore

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/zlowred/embd"
	"github.com/zlowred/embd/sensor/ds18b20"
	_ "github.com/zlowred/embd/host/rpi"
)

func main() {
	if err := embd.InitW1(); err != nil {
		panic(err)
	}
	defer embd.CloseW1()

	w1 := embd.NewW1Bus(0)

	devs, err := w1.ListDevices()

	if err != nil {
		panic(err)
	}

	var name string = ""
	for _, dev := range devs {
		if strings.HasPrefix(dev, "28-") {
			name = dev
			break
		}
	}

	if name == "" {
		fmt.Println("No DS18B20 devices found")
	}

	fmt.Printf("Using DS18B20 device %s\n", name)
	w1d, err := w1.Open(name)

	if err != nil {
		panic(err)
	}

	sensor := ds18b20.New(w1d)

	err = sensor.SetResolution(ds18b20.Resolution_12bit)
	fmt.Println("Using 12-bit resolution")

	if err != nil {
		panic(err)
	}

	var timer time.Time

	for i := 0; i < 10; i++ {
		timer = time.Now()
		err = sensor.ReadTemperature()

		if err != nil {
			fmt.Printf("error %v\n", err)
			embd.CloseW1();
			time.Sleep(time.Second)
			w1 = embd.NewW1Bus(0)
			w1d, err = w1.Open(name)

			if err != nil {
				panic(err)
			}
			sensor = ds18b20.New(w1d)

			continue
		}


		fmt.Printf("%d milliseconds for conversion\n", time.Since(timer).Nanoseconds() / 1000000)
		fmt.Printf("Measured temperature: %vC\n", sensor.Celsius())
		fmt.Printf("Measured temperature: %vF\n", sensor.Fahrenheit())
	}

	err = sensor.SetResolution(ds18b20.Resolution_9bit)
	fmt.Println("Using 9-bit resolution")

	if err != nil {
		panic(err)
	}


	for i := 0; i < 10; i++ {
		timer = time.Now()
		err = sensor.ReadTemperature()

		if err != nil {
			fmt.Printf("error %v\n", err)
			embd.CloseW1();
			time.Sleep(time.Second)
			w1 = embd.NewW1Bus(0)
			w1d, err = w1.Open(name)

			if err != nil {
				panic(err)
			}
			sensor = ds18b20.New(w1d)
			continue
		}


		fmt.Printf("%d milliseconds for conversion\n", time.Since(timer).Nanoseconds() / 1000000)
		fmt.Printf("Measured temperature: %vC\n", sensor.Celsius())
		fmt.Printf("Measured temperature: %vF\n", sensor.Fahrenheit())
	}
}
