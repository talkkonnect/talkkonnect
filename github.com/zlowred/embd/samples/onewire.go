// +build ignore

package main

import (
	"fmt"

	"github.com/zlowred/embd"
	_ "github.com/zlowred/embd/host/all"
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

	for _, dev := range devs {
		fmt.Printf("OneWire device: %s\n", dev)
	}

	w1d, err := w1.Open("28-011572120bff")

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", w1d)

	err = w1d.WriteByte(0x44)

	if err != nil {
		panic(err)
	}

	for ret, err := w1d.ReadByte(); ret == 0 && err != nil; {}

	if err != nil {
		panic(err)
	}

	err = w1d.WriteByte(0xBE)

	if err != nil {
		panic(err)
	}

	res, err := w1d.ReadBytes(9)

	if err != nil {
		panic(err)
	}

	fmt.Print("res: ")
	for _, val := range res {
		fmt.Printf("0x%02X ", val)
	}
	fmt.Println()

	var temp float64 = float64(float64(res[1]) * 256. + float64(res[0])) / 16.
	fmt.Printf("%f\n", temp)

	fmt.Println("Done")
}
