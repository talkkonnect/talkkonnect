/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * talkkonnect is the based on talkiepi and barnard by Daniel Chote and Tim Cooper
 *
 * The Initial Developer of the Original Code is
 * Suvir Kumar <suvir@talkkonnect.com>
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * oleddisplay.go modified to work with talkkonnect
 */

package talkkonnect

import (
	"log"
	"strings"
	"sync"

	hd44780 "github.com/talkkonnect/go-hd44780"
)

var mutex = &sync.Mutex{}

func oledDisplay(OledClear bool, OledRow int, OledColumn int, OledOriginalText string) {
	if !OLEDEnabled {
		log.Println("error: OLED Function Called in Error!")
		return
	}

	if OLEDInterfacetype != "i2c" {
		log.Println("error: Only i2c OLED Screens Supported Now!")
		return
	}

	OledText := stripRegex(OledOriginalText)

	if !OledClear && len(OledText) > 0 && LCDIsDark {
		Oled.DisplayOn()
	}

	if OledClear {
		Oled.Clear()
		log.Println("debug: OLED Clearing Screen")
		if len(OledText) == 0 {
			return
		}
	}

	var rpadding = int(OLEDDisplayColumns)

	if len(OledText) <= int(OLEDDisplayColumns) {
		rpadding = int(OLEDDisplayColumns) - len(OledText)
	}

	var text string = OledText + strings.Repeat(" ", rpadding)

	mutex.Lock()

	Oled.SetCursor(OledRow, OLEDStartColumn)

	if len(OledText) >= int(OLEDDisplayColumns) {
		Oled.Write(OledText[:OLEDDisplayColumns])
		//log.Printf("alert: Over  Length=%v Text [%v ", len(OledText), OledText[:OLEDDisplayColumns]+"]")
	} else {
		Oled.Write(text)
		//log.Printf("alert: Under Length=%v Text [%v", len(OledText), text+"]")
	}

	mutex.Unlock()
}

func LcdDisplay(lcdtextshow [4]string, PRSPin int, PEPin int, PD4Pin int, PD5Pin int, PD6Pin int, PD7Pin int, LCDInterfaceType string, LCDI2CAddress byte) {
	go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
}

func (b *Talkkonnect) sevenSegment(function string, value string) {
	if Config.Global.Hardware.IO.Max7219.Enabled {
		if function == "mumblechannel" {
			prefix := "c"
			if b.findEnabledRotaryEncoderFunction("mumblechannel") {
				Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix+value)
			}
		}
		if function == "localvolume" {
			prefix := "u"
			if b.findEnabledRotaryEncoderFunction("localvolume") {
				Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix+value)
			}
		}
		if function == "radiochannel" {
			prefix := "r"
			if b.findEnabledRotaryEncoderFunction("radiochannel") {
				Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix+value)
			}
		}
		if function == "voicetarget" {
			prefix := "t"
			if b.findEnabledRotaryEncoderFunction("voicetarget") {
				Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix+value)
			}
		}
		if function == "hello" {
			prefix := "hello"
			Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix)
		}
		if function == "bye" {
			prefix := "bye"
			Max7219(Config.Global.Hardware.IO.Max7219.Max7219Cascaded, Config.Global.Hardware.IO.Max7219.SPIBus, Config.Global.Hardware.IO.Max7219.SPIDevice, Config.Global.Hardware.IO.Max7219.Brightness, prefix)
		}
	} else {
		log.Println("debug: Max7219 Seven Segment Not Enabled")
	}
}
