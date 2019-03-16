package talkkonnect

import (
	goled "github.com/talkkonnect/go-oled-i2c"
	"log"
	"sync"
)

var mutex = &sync.Mutex{}

func oledDisplay (OledClear bool,OledRow int, OledColumn int, OledText string){
        mutex.Lock()
        defer mutex.Unlock()
        oled, err := goled.BeginOled()

        if err != nil {
                log.Fatal(err)
                return
        }
        defer oled.Close()

	// clear oled screen command
        if OledClear == true {
                oled.Clear()
		for i := 0; i < OledDisplayRows; i++ {
	        	oled.SetCursor(i,0)
			oled.Write("                       \n")
		}
                return
        }

	//set row and column and print to oled display command
        oled.SetCursor(OledRow, 0)
	oled.Write("                       \n")
        oled.SetCursor(OledRow, 1)
	if len(OledText) >= OledDisplayColumns {
        	oled.Write(OledText[:OledDisplayColumns])
	} else {
        	oled.Write(OledText)
	}
}




