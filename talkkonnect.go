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
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/volume-go"
)

func talkkonnectBanner() {
	log.Println("info: ┌────────────────────────────────────────────────────────────────┐")
	log.Println("info: │  _        _ _    _                               _             │")
	log.Println("info: │ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │")
	log.Println("info: │ | __/ _` | | |/ / |/ / _ \\| '_ \\| '_ \\ / _ \\/ __|  __|         │")
	log.Println("info: │ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │")
	log.Println("info: │  \\__\\__,_|_|_|\\_\\_|\\_\\___/|_| |_|_| |_|\\___|\\_ _|\\__|          │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │A Flexible Headless Mumble Transceiver/Gateway for RPi/PC/VM    │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │Created By : Suvir Kumar  <suvir@talkkonnect.com>               │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │Press the <Del> key for Menu or <Ctrl-c> to Quit talkkonnect    │")
	log.Println("info: │Additional Modifications Released under MPL 2.0 License         │")
	log.Println("info: │visit us at www.talkkonnect.com and github.com/talkkonnect      │")
	log.Println("info: └────────────────────────────────────────────────────────────────┘")
	log.Printf("info: Talkkonnect Version %v Released %v", talkkonnectVersion, talkkonnectReleased)
	log.Printf("info: ")
}

func talkkonnectAcknowledgements() {
	log.Println("info: ┌──────────────────────────────────────────────────────────────────────────────────────────────┐")
	log.Println("info: │Acknowledgements & Inspriation from the talkkonnect team of developers, maintainers & testers │")
	log.Println("info: │talkkonnect is based on the works of many people and many open source projects                │")
	log.Println("info: ├───────────────────────────────────────────────────────────────────────────────────────────── ┤")
	log.Println("info: │Thanks to Organizations :-                                                                    │")
	log.Println("info: │The Mumble Development team, Raspberry Pi Foundation, Developers and Maintainers of Debian    │")
	log.Println("info: │The Creators and Maintainers of Golang and all the libraries available on github.com          │")
	log.Println("info: │                                                                                              │")
	log.Println("info: │Thanks to Individuals :-                                                                      │")
	log.Println("info: │Daniel Chote Creator of talkiepi and Tim Cooper Creator of Barnard and gumble library         │")
	log.Println("info: │Zoran Dimitrijevic for his commitment, building, testing, docummentation and kind feedback    │")
	log.Println("info: │enabling us to take talkkonnect to use cases never orignially imagined                        │")
	log.Println("info: ├──────────────────────────────────────────────────────────────────────────────────────────────┤")
	log.Println("info: │visit us at www.talkkonnect.com and github.com/talkkonnect                                    │")
	log.Println("info: │talkkonnect was created by Suvir Kumar <suvir@talkkonnect.com> & Released under MPLV2 License │")
	log.Println("info: └──────────────────────────────────────────────────────────────────────────────────────────────┘")
}

func (b *Talkkonnect) talkkonnectMenu() {
	log.Println("info: ┌──────────────────────────────────────────────────────────────┐")
	log.Println("info: │ _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _         │")
	log.Println("info: │| '_ ` _ \\ / _` | | '_ \\  | '_ ` _ \\ / _ \\ '_ \\| | | |        │")
	log.Println("info: │| | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |        │")
	log.Println("info: │|_| |_| |_|\\__,_|_|_| |_| |_| |_| |_|\\___|_| |_|\\__,_|        │")
	log.Println("info: ├─────────────────────────────┬────────────────────────────────┤")
	log.Println("info: │ <Del> to Display this Menu  | Ctrl-C to Quit talkkonnect     │")
	log.Println("info: ├─────────────────────────────┼────────────────────────────────┤")
	log.Println("info: │ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)         │")
	log.Println("info: │ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level     │")
	log.Println("info: │ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)  │")
	log.Println("info: │ <F7>  List Server Channels  │ <F8>  Start Transmitting       │")
	log.Println("info: │ <F9>  Stop Transmitting     │ <F10> List Online Users        │")
	log.Println("info: │ <F11> Playback/Stop Chimes  │ <F12> For GPS Position         │")
	log.Println("info: ├─────────────────────────────┼────────────────────────────────┤")
	log.Println("info: │<Ctrl-E> Send Email          │<Ctrl-N> Conn Next Server       │")
	log.Println("info: │<Ctrl-F> Conn Previous Server│<Ctrl-P> Panic Simulation       │")
	log.Println("info: │<Ctrl-Q> Reserved            │<Ctrl-S> Scan Channels          │")
	log.Println("info: │<Ctrl-V> Display Version     │<Ctrl-T> Thanks/Acknowledgements│")
	log.Println("info: ├─────────────────────────────┼────────────────────────────────┤")
	log.Println("info: │<Ctrl-L> Clear Screen        │<Ctrl-O> Ping Servers           │")
	log.Println("info: │<Ctrl-R> Repeat TX Loop Test │<Ctrl-X> Dump XML Config        │")
	log.Println("info: ├─────────────────────────────┼────────────────────────────────┤")
	log.Println("info: │<Ctrl-I> Traffic Record      │<Ctrl-J> Mic Record             │")
	log.Println("info: │<Ctrl-K> Traffic & Mic Record│<Ctrl-U> Show Uptime            │")
	log.Println("info: ├─────────────────────────────┼────────────────────────────────┤")
	log.Println("info: │  visit us at www.talkkonnect.com and github.com/talkkonnect  │")
	log.Println("info: └──────────────────────────────────────────────────────────────┘")

	log.Println("info: IP Address & Session Information")
	b.pingconnectedserver()
	localAddresses()

	origMuted, _ := volume.GetMuted(OutputDevice)
	if origMuted {
		log.Println("info: Speaker Currently Muted")
	} else {
		origVolume, err := volume.GetVolume(OutputDevice)
		if err == nil {
			log.Printf("info: Speaker Not Muted & Current Volume at Level %v%%\n", origVolume)
		} else {
			log.Println("alert: Can't Get Volume Level From Sound Card!")
		}
	}
	hostname, err1 := os.Hostname()
	if err1 != nil {
		log.Printf("alert: Cannot Get Hostname\n")
	} else {
		log.Printf("info: Hostname is %s\n", hostname)
	}

	log.Printf("info: Talkkonnect Version %v Released %v\n", talkkonnectVersion, talkkonnectReleased)
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("error: localAddresses %v", err.Error()))
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Print(fmt.Errorf("error: localAddresses %v", err.Error()))
			continue
		}

		for _, a := range addrs {
			if i.Name != "lo" {
				log.Printf("info: %v %v\n", i.Name, a)
			}
		}
	}
}

func (b *Talkkonnect) pingconnectedserver() {

	resp, err := gumble.Ping(b.Address, time.Second*1, time.Second*5)

	if err != nil {
		log.Println(fmt.Sprintf("error: Ping Error %s", err))
		return
	}

	major, minor, patch := resp.Version.SemanticVersion()

	log.Println("info: Server Address:         ", resp.Address)
	log.Println("info: Current Channel:        ", b.Client.Self.Channel.Name)
	log.Println("info: Server Ping:            ", resp.Ping)
	log.Println("info: Server Version:         ", major, ".", minor, ".", patch)
	log.Println("info: Server Users:           ", resp.ConnectedUsers, "/", resp.MaximumUsers)
	log.Println("info: Server Maximum Bitrate: ", resp.MaximumBitrate)
}
