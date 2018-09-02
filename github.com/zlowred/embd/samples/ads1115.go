// +build ignore

package main

import (
	"flag"
	"fmt"

	"github.com/zlowred/embd"

	_ "github.com/zlowred/embd/host/all"
)

func main() {
	flag.Parse()

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	sensor := ads1115.New(bus, 0x48)

	if res, err := sensor.Read(); err != nil {
		panic(err)
	} else {
		fmt.Printf("Converted value: %04X\n", res)
	}
}
