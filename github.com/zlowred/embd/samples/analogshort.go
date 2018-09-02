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

	embd.InitGPIO()
	defer embd.CloseGPIO()

	val, _ := embd.AnalogRead(0)
	fmt.Printf("Reading: %v\n", val)
}
