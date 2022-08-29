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
 * radio.go -> talkkonnect function to interface to radio modules (currently sa818 only)

 */

package talkkonnect

import (
	"log"
	"time"

	"github.com/talkkonnect/sa818"
)

var CurrentChannelIndex = 0
var MoveChannelIndex = 0
var EnabledChannelCounter = 0

func radioSetChannel(channelID string) {
	DMOSetup.SerialOptions.PortName = Config.Global.Hardware.Radio.Sa818.Serial.Port
	DMOSetup.SerialOptions.BaudRate = Config.Global.Hardware.Radio.Sa818.Serial.Baud
	DMOSetup.SerialOptions.DataBits = Config.Global.Hardware.Radio.Sa818.Serial.Databits
	DMOSetup.SerialOptions.StopBits = Config.Global.Hardware.Radio.Sa818.Serial.Stopbits
	DMOSetup.SerialOptions.MinimumReadSize = 2
	DMOSetup.SerialOptions.InterCharacterTimeout = 200
	RadioModuleSA818Channel(channelID, true, true)
}

func radioChannelIncrement(command string) {
	DMOSetup.SerialOptions.PortName = Config.Global.Hardware.Radio.Sa818.Serial.Port
	DMOSetup.SerialOptions.BaudRate = Config.Global.Hardware.Radio.Sa818.Serial.Baud
	DMOSetup.SerialOptions.DataBits = Config.Global.Hardware.Radio.Sa818.Serial.Databits
	DMOSetup.SerialOptions.StopBits = Config.Global.Hardware.Radio.Sa818.Serial.Stopbits
	DMOSetup.SerialOptions.MinimumReadSize = 2
	DMOSetup.SerialOptions.InterCharacterTimeout = 200

	if command == "up" {
		if Config.Global.Hardware.Radio.Enabled {
			if len(radioChannels)-1 < CurrentChannelIndex+1 {
				MoveChannelIndex = 0
				CurrentChannelIndex = 0
				log.Printf("info: Moving %v To Channel ID %v Name %v\n", command, radioChannels[MoveChannelIndex].ID, radioChannels[MoveChannelIndex].Name)
				RadioModuleSA818Channel(radioChannels[MoveChannelIndex].ID, true, true)
				return
			}
			if len(radioChannels)-1 >= CurrentChannelIndex+1 {
				MoveChannelIndex = CurrentChannelIndex + 1
				CurrentChannelIndex++
				log.Printf("info: Moving %v To Channel ID %v Name %v\n", command, radioChannels[MoveChannelIndex].ID, radioChannels[MoveChannelIndex].Name)
				RadioModuleSA818Channel(radioChannels[MoveChannelIndex].ID, true, true)
				return
			}
		} else {
			log.Println("error: Radio Channel ID Up Requested But Radio Disabled in Config")
			return
		}
	}

	if command == "down" {
		if Config.Global.Hardware.Radio.Enabled {
			if CurrentChannelIndex-1 < 0 {
				MoveChannelIndex = len(radioChannels) - 1
				CurrentChannelIndex = len(radioChannels) - 1
				log.Printf("info: Moving %v To Channel ID %v Name %v\n", command, radioChannels[MoveChannelIndex].ID, radioChannels[MoveChannelIndex].Name)
				RadioModuleSA818Channel(radioChannels[MoveChannelIndex].ID, true, true)
				return
			}
			if CurrentChannelIndex-1 >= 0 {
				MoveChannelIndex = CurrentChannelIndex - 1
				CurrentChannelIndex--
				log.Printf("info: Moving %v To Channel ID %v Name %v\n", command, radioChannels[MoveChannelIndex].ID, radioChannels[MoveChannelIndex].Name)
				RadioModuleSA818Channel(radioChannels[MoveChannelIndex].ID, true, true)
				return
			}
		} else {
			log.Println("error: Radio Channel ID Down Requested But Radio Disabled in Config")
			return
		}
	}
}

func createEnabledRadioChannels() {
	for _, channel := range Config.Global.Hardware.Radio.Sa818.Channels.Channel {
		if channel.Enabled {
			EnabledChannelCounter++
			if channel.ID == Config.Global.Hardware.Radio.ConnectChannelID {
				CurrentChannelIndex = EnabledChannelCounter - 1
			}
			radioChannels = append(radioChannels, radioChannelsStruct{channel.ID, channel.Name, channel.ItemInList, channel.Bandwidth, channel.Rxfreq, channel.Txfreq, channel.Squelch, channel.Ctcsstone, channel.Dcstone, channel.Predeemph, channel.Highpass, channel.Lowpass, channel.Volume})
		}
	}
}

func RadioModuleSA818Channel(useChannelID string, setVolumeToo bool, setFilterToo bool) {
	found, name := findChannelNameByID(useChannelID)
	if found {
		log.Printf("info: Found Channel ID %v Name %v\n", useChannelID, name)
		setFrequency()
		if setVolumeToo {
			time.Sleep(500 * time.Millisecond)
			setVolume()
		}
		if setFilterToo {
			time.Sleep(700 * time.Millisecond)
			setFilter()
		}
	} else {
		log.Printf("error: Not Found Channel ID %v\n", useChannelID)
	}
}

func findChannelNameByID(findChannelID string) (bool, string) {
	var EnabledItemInList int = 0
	for Item, channel := range Config.Global.Hardware.Radio.Sa818.Channels.Channel {
		if channel.Enabled {
			EnabledItemInList++
			Config.Global.Hardware.Radio.Sa818.Channels.Channel[Item].ItemInList = EnabledItemInList
			if channel.ID == findChannelID {
				DMOSetup.Band = channel.Bandwidth
				DMOSetup.Rxfreq = channel.Rxfreq
				DMOSetup.Txfreq = channel.Txfreq
				DMOSetup.Ctsstone = channel.Ctcsstone
				DMOSetup.Squelch = channel.Squelch
				DMOSetup.Dcstone = channel.Dcstone
				DMOSetup.Predeemph = channel.Predeemph
				DMOSetup.Highpass = channel.Highpass
				DMOSetup.Lowpass = channel.Lowpass
				DMOSetup.Volume = channel.Volume
				return true, channel.Name
			}
		}
	}
	return false, "not found channel"
}

//func checkVersion() {
//	err := sa818.Callsa818("CheckVersion", DMOSetup)
//	log.Println("info: CheckVersion ", err)
//}

//func checkRSSI() {
//	err := sa818.Callsa818("CheckRSSI", DMOSetup)
//	log.Println("info: Check RSSI ", err)
//}

func setFrequency() {
	err := sa818.Callsa818("DMOSetupGroup", DMOSetup)
	//log.Printf("debug: actual data sent to module %#v", DMOSetup)
	if err != nil {
		log.Println("info: SAModule Set Frequecy Error ", err)
	} else {
		log.Println("info: SAModule Set Frequecy OK ")
	}
}

func setFilter() {
	err := sa818.Callsa818("DMOSetupFilter", DMOSetup)
	if err != nil {
		log.Println("info: SAModule Setup Filter Error ", err)
	} else {
		log.Println("info: SAModule Setup Filter OK ")
	}
}

func setVolume() {
	err := sa818.Callsa818("SetVolume", DMOSetup)
	if err != nil {
		log.Println("info: SAModule Set Volume Error ", err)
	} else {
		log.Println("info: SAModule Set Volume OK ")
	}
}
