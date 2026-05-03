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
 * talkkonnect.go -> function in talkkonnect for printing banners to screen
 */

package talkkonnect

import (
	"log"
	"strconv"

	"github.com/talkkonnect/volume-go"
)

func talkkonnectBanner(backgroundcolor string) {
	var backgroundreset string = "\u001b[0m"
	log.Println("info: " + backgroundcolor + "┌────────────────────────────────────────────────────────────────┐" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│  _        _ _    _                               _             │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ | __/ _` | | |/ / |/ / _ \\| '_ \\| '_ \\ / _ \\/ __|  __|         │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│  \\__\\__,_|_|_|\\_\\_|\\_\\___/|_| |_|_| |_|\\___|\\_ _|\\__|          │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├────────────────────────────────────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│A Flexible Headless Mumble Transceiver/Gateway for RPi/PC/VM    │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├────────────────────────────────────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│Created By : Suvir Kumar  <suvir@talkkonnect.com>               │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├────────────────────────────────────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│Press the <Del> key for Menu or <Ctrl-c> to Quit talkkonnect    │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│Additional Modifications Released under MPL 2.0 License         │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│Blog at www.talkkonnect.com, source at github.com/talkkonnect   │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "└────────────────────────────────────────────────────────────────┘" + backgroundreset)
	releasedVersion := talkkonnectVersion
	if IsConnected {
		releasedVersion = checkGitHubVersion()
	}
	if talkkonnectVersion != releasedVersion {
		log.Printf("warn: Software Ver %v Rel %v (Different Ver %v Available!)\n", talkkonnectVersion, talkkonnectReleased, releasedVersion)
	} else {
		log.Printf("info: Software Ver %v Rel %v (Latest Release)\n", talkkonnectVersion, talkkonnectReleased)
	}
	boardVersion := checkSBCVersion()
	if boardVersion != "unknown" {
		log.Printf("info: Hardware Detected As %v\n", boardVersion)
	} else {
		log.Println("info: ")
	}
}

func talkkonnectAcknowledgements(backgroundcolor string) {
	var backgroundreset string = "\u001b[0m"
	ln := func(s string) {
		log.Println(s)
		sshRemoteReplyF(s + "\n")
	}
	ln("info: " + backgroundcolor + "┌──────────────────────────────────────────────────────────────────────────────────────────────┐" + backgroundreset)
	ln("info: " + backgroundcolor + "│Acknowledgements & Inspriation from the talkkonnect team of developers, maintainers & testers │" + backgroundreset)
	ln("info: " + backgroundcolor + "│talkkonnect is based on the works of many people and many open source projects                │" + backgroundreset)
	ln("info: " + backgroundcolor + "├──────────────────────────────────────────────────────────────────────────────────────────────┤" + backgroundreset)
	ln("info: " + backgroundcolor + "│Thanks to Organizations :-                                                                    │" + backgroundreset)
	ln("info: " + backgroundcolor + "│The Mumble Development team, Raspberry Pi Foundation, Developers and Maintainers of Debian    │" + backgroundreset)
	ln("info: " + backgroundcolor + "│The Creators and Maintainers of Golang and all the libraries available on github.com          │" + backgroundreset)
	ln("info: " + backgroundcolor + "│Global Coders Co., Ltd. For Sponsoring this project                                           │" + backgroundreset)
	ln("info: " + backgroundcolor + "│                                                                                              │" + backgroundreset)
	ln("info: " + backgroundcolor + "│Thanks to Individuals :-                                                                      │" + backgroundreset)
	ln("info: " + backgroundcolor + "│Daniel Chote Creator of talkiepi and Tim Cooper Creator of Barnard and gumble library         │" + backgroundreset)
	ln("info: " + backgroundcolor + "│Zoran Dimitrijevic for his commitment, building, testing, docummentation and kind feedback    │" + backgroundreset)
	ln("info: " + backgroundcolor + "│enabling us to take talkkonnect to use cases never originally imagined                        │" + backgroundreset)
	ln("info: " + backgroundcolor + "├──────────────────────────────────────────────────────────────────────────────────────────────┤" + backgroundreset)
	ln("info: " + backgroundcolor + "│visit us at www.talkkonnect.com and github.com/talkkonnect                                    │" + backgroundreset)
	ln("info: " + backgroundcolor + "│talkkonnect was created by Suvir Kumar <suvir@talkkonnect.com> & Released under MPLV2 License │" + backgroundreset)
	ln("info: " + backgroundcolor + "└──────────────────────────────────────────────────────────────────────────────────────────────┘" + backgroundreset)
}

func (b *Talkkonnect) talkkonnectMenu(backgroundcolor string) {
	var backgroundreset string = "\u001b[0m"
	log.Println("info: " + backgroundcolor + " ┌──────────────────────────────────────────────────────────────┐" + backgroundreset)
	log.Println("info: " + backgroundcolor + " │     _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _     │" + backgroundreset)
	log.Println("info: " + backgroundcolor + " │    | '_ ` _ \\ / _` | | '_ \\  | '_ ` _ \\ / _ \\ '_ \\| | | |    │" + backgroundreset)
	log.Println("info: " + backgroundcolor + " │    | | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |    │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│    |_| |_| |_|\\__,_|_|_| |_| |_| |_| |_|\\___|_| |_|\\__,_|    │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├─────────────────────────────┬────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <1> to Display this Menu    | <Ctrl-C> to Quit talkkonnect   │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├─────────────────────────────┼────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <2> Channel UP (+)          │ <3>  Channel Down (-)          │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <4> Mute/Unmute Speaker     │ <5>  Digital Volume Up (+)     │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <6> Digital Volume Down (-) │ <7>  Start Transmitting        │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <8> Stop Transmitting       │ <9> List Online Users          │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <0> Show Uptime             │                                │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├─────────────────────────────┼────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <a> List API Commands       │<b> Playback/Stop Stream        │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <d> Dump XML Config         │<e> Send Email                  │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <g> GPS Position            │<h> XML Config Checker          │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <i> Traffic Record          │<j> Mic Record                  │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <k> Traffic & Mic Record    │<l> Clear Screen                │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <m> Radio Channel (+)       │<n> Radio Channel (-)           │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <o> Ping Servers            │<p> Panic Simulation            │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <q> Repeat TX Loop Test     │<r> Scan Channels               │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <s> Thanks/Acknowledge      │<t> Show Uptime                 │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <u> Display Version         │<v> Online Radio On/Off         │" + backgroundreset)
    log.Println("info: " + backgroundcolor + "│ <w> Dump XML Config         │<x> Next Server                 │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│ <z> Next Server             │                                │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "├─────────────────────────────┼────────────────────────────────┤" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│  Visit us at www.talkkonnect.com and github.com/talkkonnect  │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "│  Thanks to Global Coders Co., Ltd. for their sponsorship     │" + backgroundreset)
	log.Println("info: " + backgroundcolor + "└──────────────────────────────────────────────────────────────┘" + backgroundreset)
	log.Println("info: IP Address & Session Information")
	b.pingconnectedserver()
	localAddresses()
	log.Println("info: Internet WAN IP Is", getOutboundIP())

	macaddress, err := getMacAddr()
	if err != nil {
		log.Println("error: Could Not Get Network Interface MAC Address")
	} else {
		for i, a := range macaddress {
			log.Println("info: Network Interface MAC Address (" + strconv.Itoa(i) + "): " + a)
		}
	}

	origMuted, _ := volume.GetMuted(Config.Global.Software.Settings.OutputDevice)
	if origMuted {
		log.Println("info: Speaker Currently Muted")
	} else {
		origVolume, err := volume.GetVolume(Config.Global.Software.Settings.OutputDevice)
		if err == nil {
			log.Printf("info: Speaker Not Muted & Current Volume at Level %v%%\n", origVolume)
		} else {
			log.Println("error: Can't Get Volume Level From Sound Card!")
		}
	}
	log.Printf("info: Talkkonnect Version %v Released %v\n", talkkonnectVersion, talkkonnectReleased)
}
