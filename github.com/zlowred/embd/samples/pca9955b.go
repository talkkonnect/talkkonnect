// +build ignore

package main

import (
	"flag"

	"github.com/zlowred/embd"
	"github.com/zlowred/embd/controller/pca9955b"
	_ "github.com/zlowred/embd/host/all"
	"fmt"
)

func main() {
	flag.Parse()

	if err := embd.InitI2C(); err != nil {
		panic(err)
	}
	defer embd.CloseI2C()

	bus := embd.NewI2CBus(1)

	pca9955b := pca9955b.New(bus, 0x0B)

	pca9955b.Reset()

	fmt.Println("running in cycle")
	for {
		for x := 0; x < 256; x++ {
			if err := pca9955b.SetOutput(byte(7), byte(x)); err != nil {
				panic(err)
			}
		}
		for x := 255; x > 0; x-- {
			if err := pca9955b.SetOutput(byte(7), byte(x)); err != nil {
				panic(err)
			}
		}
	}
}
