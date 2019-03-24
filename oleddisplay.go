package talkkonnect

import (
	goled "github.com/talkkonnect/go-oled-i2c"
	"log"
	"strings"
	"sync"
)

var mutex = &sync.Mutex{}

func oledDisplay(OledClear bool, OledRow int, OledColumn int, OledText string) {
	mutex.Lock()
	defer mutex.Unlock()

	if OLEDEnabled == false {
		log.Println("warn: OLED Function Called in Error!")
		return
	}

	if OLEDInterfacetype != "i2c" {
		log.Println("warn: Only i2c OLED Screens Supported Now!")
		return
	}

	oled, err := goled.BeginOled(OLEDDefaultI2cAddress, OLEDDefaultI2cBus, OLEDScreenWidth, OLEDScreenHeight, OLEDDisplayRows, OLEDDisplayColumns, OLEDStartColumn, OLEDCharLength, OLEDCommandColumnAddressing, OLEDAddressBasePageStart)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer oled.Close()

	// clear oled screen command
	if OledClear == true {
		oled.Clear()
		log.Println("warn: OLED Clearing Screen")
	}

	oled.SetCursor(OledRow, 0)

	var rpadding = int(OLEDDisplayColumns)

	if len(OledText) <= int(OLEDDisplayColumns) {
		rpadding = int(OLEDDisplayColumns) - len(OledText)
	}

	var text string = OledText + strings.Repeat(" ", rpadding)

	oled.SetCursor(OledRow, OLEDStartColumn)

	if len(OledText) >= int(OLEDDisplayColumns) {
		oled.Write(OledText[:OLEDDisplayColumns])
	} else {
		oled.Write(text)
	}
}
