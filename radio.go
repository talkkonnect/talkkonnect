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
	DMOSetup.SerialOptions.PortName = Config.Global.Hardware.Radio.Sa818.Serial.Port
	DMOSetup.SerialOptions.BaudRate = Config.Global.Hardware.Radio.Sa818.Serial.Baud
	DMOSetup.SerialOptions.DataBits = Config.Global.Hardware.Radio.Sa818.Serial.Databits
	DMOSetup.SerialOptions.StopBits = Config.Global.Hardware.Radio.Sa818.Serial.Stopbits
	DMOSetup.SerialOptions.MinimumReadSize = 2
	DMOSetup.SerialOptions.InterCharacterTimeout = 200
	RadioModuleSA818Channel(Config.Global.Hardware.Radio.ConnectChannelID)
}

func RadioModuleSA818Channel(useChannelID string) {
	found, name := findChannelByID(useChannelID)
	if found {
		log.Printf("info: Found Channel ID %v Name %v\n", useChannelID, name)
		setFrequency()
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

func checkVersion() {
	err := sa818.Callsa818("CheckVersion", DMOSetup)
	log.Println("info: CheckVersion ", err)
}

func checkRSSI() {
	err := sa818.Callsa818("CheckRSSI", DMOSetup)
	log.Println("info: Check RSSI ", err)
}

func setFrequency() {
	err := sa818.Callsa818("DMOSetupGroup", DMOSetup)
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
