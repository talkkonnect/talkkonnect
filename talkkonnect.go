package talkkonnect

import (
	"fmt"
	"github.com/talkkonnect/volume-go"
	"log"
	"net"
)

func talkkonnectBanner() {
	log.Println("info: ┌────────────────────────────────────────────────────────────────┐")
	log.Println("info: │  _        _ _    _                               _             │")
	log.Println("info: │ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │")
	log.Println("info: │ | __/ _` | | |/ / |/ / _ \\| '_ \\| '_ \\ / _ \\/ __|  __|         │")
	log.Println("info: │ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │")
	log.Println("info: │  \\__\\__,_|_|_|\\_\\_|\\_\\___/|_| |_|_| |_|\\___|\\_ _|\\__|          │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │A Flexible Headless Mumble Transceiver/Gateway for Raspberry Pi │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │Created By : Suvir Kumar  <suvir@talkkonnect.com>               │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │Version 1.20 Released December 2018                             │")
	log.Println("info: │Additional Modifications Released under MPL 2.0 License         │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │visit us at www.talkkonnect.com and github.com/talkkonnect      │")
	log.Println("info: └────────────────────────────────────────────────────────────────┘")
	log.Println("info: Press the <Del> key for Menu Options or <Ctrl-c> to Quit talkkonnect")
}

func talkkonnectAcknowledgements() {
	log.Println("info: ┌───────────────────────────────────────────────────────────────────────────────────────────┐")
	log.Println("info: │Acknowledgements & Inspriation from the talkkonnect team of developers and maintainers     │")
	log.Println("info: ├───────────────────────────────────────────────────────────────────────────────────────────┤")
	log.Println("info: │talkkonnect is based on the works of many people and many open source projects             │")
	log.Println("info: │                                                                                           │")
	log.Println("info: │Thanks to :-                                                                               │")
	log.Println("info: │                                                                                           │")
	log.Println("info: │Organizations :-                                                                           │")
	log.Println("info: │The Mumble Development team, Raspberry Pi Foundation, Developers and Maintainers of Debian │")
	log.Println("info: │The Creators and Maintainers of Golang and all the libraries available on github.com       │")
	log.Println("info: │                                                                                           │")
	log.Println("info: │Individuals :-                                                                             │")
	log.Println("info: │Daniel Chote Creator of talkiepi and Tim Cooper Ceator of Barnard                          │")
	log.Println("info: │Tayeb Meftah and other people who wish to remain anonymous for their feedback and testing  │")
	log.Println("info: ├───────────────────────────────────────────────────────────────────────────────────────────┤")
	log.Println("info: │visit us at www.talkkonnect.com and github.com/talkkonnect <suvir@talkkonnect.com>         │")
	log.Println("info: └───────────────────────────────────────────────────────────────────────────────────────────┘")
}

func (b *Talkkonnect) talkkonnectMenu() {
	log.Println("info: ┌────────────────────────────────────────────────────────────────┐")
	log.Println("info: │                 _                                              │")
	log.Println("info: │ _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _           │")
	log.Println("info: │| '_ ` _ \\ / _` | | '_ \\  | '_ ` _ \\ / _ \\ '_ \\| | | |          │")
	log.Println("info: │| | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |          │")
	log.Println("info: │|_| |_| |_|\\__,_|_|_| |_| |_| |_| |_|\\___|_| |_|\\__,_|          │")
	log.Println("info: ├─────────────────────────────┬──────────────────────────────────┤")
	log.Println("info: │ <Del> to Display this Menu  | Ctrl-C to Quit talkkonnect       │")
	log.Println("info: ├─────────────────────────────┼──────────────────────────────────┤")
	log.Println("info: │ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)           │")
	log.Println("info: │ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level       │")
	log.Println("info: │ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)    │")
	log.Println("info: │ <F7>  List Server Channels  │ <F8>  Start Transmitting         │")
	log.Println("info: │ <F9>  Stop Transmitting     │ <F10> List Online Users          │")
	log.Println("info: │ <F11> Playback/Stop Chimes  │ <F12> For GPS Position           │")
	log.Println("info: │<Ctrl-P> Start/Stop Panic Sim│<Ctrx-X> Screen Dump XML Config   │")
	log.Println("info: │<Ctrl-E> Send Email          │                                  │")
	log.Println("info: ├─────────────────────────────┴──────────────────────────────────┤")
	log.Println("info: │   visit us at www.talkkonnect.com and github.com/talkkonnect   │")
	log.Println("info: └────────────────────────────────────────────────────────────────┘")

	log.Println("info: IP Address & Session Information")
	localAddresses()

	origMuted, _ := volume.GetMuted(OutputDevice)
	if origMuted {
		log.Println("info: Speaker Currently Muted")
	} else {
		log.Println("info: Speaker Currently Not Muted")
	}
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("error: localAddresses %v\n", err.Error()))
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Print(fmt.Errorf("error: localAddresses %v\n", err.Error()))
			continue
		}

		for _, a := range addrs {
			if i.Name != "lo" {
				log.Printf("info: %v %v\n", i.Name, a)
			}
		}
	}
}
