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

	"github.com/talkkonnect/sa818"
)

func radioSetup() {
	DMOSetup.PortName = Config.Global.Hardware.Radio.Sa818.Serial.Port
	DMOSetup.BaudRate = Config.Global.Hardware.Radio.Sa818.Serial.Baud
	DMOSetup.DataBits = Config.Global.Hardware.Radio.Sa818.Serial.Databits
	DMOSetup.StopBits = Config.Global.Hardware.Radio.Sa818.Serial.Stopbits
	moduleResponding, message := RadioModuleSA818InitComm(DMOSetup)
	if !moduleResponding {
		log.Println("error: ", message)
	} else {
		radioChannelID := "01"
		found, name := findChannelByID(radioChannelID)
		if found {
			log.Printf("info: Found Channel ID %v Name %v\n", radioChannelID, name)
			//RadioModuleSA818InitCheckVersion()
			//RadioModuleSA818InitCheckRSSI()
			RadioModuleSA818SetDMOGroup()
			RadioModuleSA818SetDMOFilter()
			RadioModuleSA818SetVolume()
		}
	}

}

func RadioModuleSA818Channel(useChannelID string) {

	found, name := findChannelByID(useChannelID)
	if found {
		log.Printf("info: Found Channel ID %v Name %v\n", useChannelID, name)

	} else {
		log.Printf("error: Not Found Channel ID %v\n", useChannelID)
	}
}

func findChannelByID(findChannelID string) (bool, string) {
	for _, channel := range Config.Global.Hardware.Radio.Sa818.Channels.Channel {
		if channel.ID == findChannelID && channel.Enabled {
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
	return false, "not found channel"
}

func RadioModuleSA818InitComm(DMOSetup sa818.DMOSetupStruct) (bool, string) {
	message, err := sa818.Callsa818("InitComm", "(DMOCONNECT:0)", DMOSetup)
	if err != nil {
		return false, "sa818 communication error"
	} else {
		return true, message
	}
}

func RadioModuleSA818InitCheckVersion() {
	message, err := sa818.Callsa818("CheckVersion", "(VERSION:)", DMOSetup)
	if err != nil {
		log.Println("error: From Module ", err)
	} else {
		log.Println("info: sa818 says ", message)
	}
}

func RadioModuleSA818InitCheckRSSI() {
	message, err := sa818.Callsa818("CheckRSSI", "(RSSI)", DMOSetup)
	if err != nil {
		log.Println("error: From Module ", err)
	} else {
		log.Println("info: sa818 says ", message)
	}
}

func RadioModuleSA818SetVolume() {
	message, err := sa818.Callsa818("SetVolume", "(DMOSETVOLUME:0)", DMOSetup)
	if err != nil {
		log.Println("error: From Module ", err)
	} else {
		log.Println("info: sa818 says ", message)
	}
}

func RadioModuleSA818SetDMOFilter() {
	message, err := sa818.Callsa818("DMOSetupFilter", "(DMOSETFILTER:0)", DMOSetup)
	if err != nil {
		log.Println("error: From Module ", err)
	} else {
		log.Println("info: sa818 says ", message)
	}
}

func RadioModuleSA818SetDMOGroup() {
	message, err := sa818.Callsa818("DMOSetupGroup", "(DMOSETGROUP:0)", DMOSetup)
	if err != nil {
		log.Println("error: From Module ", err)
	} else {
		log.Println("info: sa818 says ", message)
	}

}
