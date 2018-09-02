package main

import (
	hd44780 "github.com/go-hd44780"
)

var lcdtext = [4]string{"uno","dos","tres","quatro"}

func main() {
        hd44780.LcdDisplay(lcdtext)
}

