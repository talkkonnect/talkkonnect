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
 * The source code for ledstrip was copied from github.com/CustomMachines and the code was written by Ben Lewis licensed under MPL Version 2 License
 * The Initial Developer of the Original Code is
 * Ben Lewis
 *
 * Contributor(s):
 *
 * Ben Lewis
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * ledstrip.go -> function in talkkonnect to control the led strip on respeaker hat
 */


package talkkonnect
/*
// uncomment code for working leds on spi leds on respeaker, its a mine field you have been warned!

import (
	"errors"
	"log"
	"strconv"

	"github.com/talkkonnect/periph/x/periph/conn/physic"
	"github.com/talkkonnect/periph/x/periph/conn/spi"
 	"github.com/talkkonnect/periph/x/periph/conn/spi/spireg"
	"github.com/talkkonnect/periph/x/periph/devices/apa102"
	"github.com/talkkonnect/periph/x/periph/host"
)

const (
	numLEDs           int    = 3
	SOnlineLED        int    = 0
	SVoiceActivityLED int    = 1
	STransmitLED      int    = 2
	OnlineCol         string = "00FF00" //Green
	VoiceActivityCol  string = "0000FF" //Blue
	TransmitCol       string = "FF0000" //Red
	OffCol            string = "000000" //Off
)


type LedStrip struct {
	buf          []byte
	display      *apa102.Dev
	spiInterface spi.PortCloser
}


func NewLedStrip() (*LedStrip, error) {
	var spiID string = "SPI0.0"   //SPI port to use
	var intensity uint8 = 16      //light intensity [1-255]
	var temperature uint16 = 5000 //light temperature in °Kelvin [3500-7500]
	var hz physic.Frequency       //SPI port speed
	var globalPWM bool = false

	if _, err := host.Init(); err != nil {
		return nil, err
	}

	// Open the display device.
	s, err := spireg.Open(spiID)
	if err != nil {
		return nil, err
	}
	//Set port speed
	if hz != 0 {
		if err := s.LimitSpeed(hz); err != nil {
			return nil, err
		}
	}
	if p, ok := s.(spi.Pins); ok {
		log.Printf("debug: Using pins CLK: %s  MOSI: %s  MISO: %s", p.CLK(), p.MOSI(), p.MISO())
	}
	o := apa102.DefaultOpts
	o.NumPixels = numLEDs
	o.Intensity = intensity
	o.Temperature = temperature
	o.DisableGlobalPWM = globalPWM
	display, err := apa102.New(s, &o)
	if err != nil {
		return nil, err
	}
	log.Printf("debug: init display: %s\n", display)

	buf := make([]byte, numLEDs*3)

	return &LedStrip{
		buf:          buf,
		display:      display,
		spiInterface: s,
	}, nil
}

func (ls *LedStrip) ledCtrl(num int, color string) error {
	if !Config.Global.Hardware.LedStripEnabled {
		return errors.New("LedStrip Not Enabled in Config")
	}
	rgb, err := strconv.ParseUint(color, 16, 32)
	if err != nil {
		return err
	}
	r := byte(rgb >> 16)
	g := byte(rgb >> 8)
	b := byte(rgb)
	ls.buf[num*numLEDs+0] = r
	ls.buf[num*numLEDs+1] = g
	ls.buf[num*numLEDs+2] = b

	_, err = ls.display.Write(ls.buf)

	return err
}
*/
