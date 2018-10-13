package talkkonnect

import (
	"crypto/tls"
	"fmt"
	"github.com/talkkonnect/beep"
	"github.com/talkkonnect/beep/speaker"
	"github.com/talkkonnect/beep/wav"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/gumble/gumbleopenal"
	"log"
	"net"
	"os"
	"time"
)

const (
	//  Modified By Suvir Kumar to Match GPIO Pins I used for my Hardware Implentation
	ParticipantsLEDPin uint = 3  // GPIO 3 ->  Raspberry Pi Physical Pin 5
	TransmitLEDPin     uint = 4  // GPIO 4 ->  Raspberry Pi Physical Pin 7
	OnlineLEDPin       uint = 5  // GPIO 5 ->  Raspberry Pi Physical Pin 3
	TxButtonPin        uint = 26 // GPIO 26 -> Raspberry Pi Physical Pin 37
	UpButtonPin        uint = 19 // GPIO 19 -> Raspberry Pi Physical Pin 35
	DownButtonPin      uint = 13 // GPIO 13 -> Raspberry Pi Physical Pin 33
)

type Talkkonnect struct {
	Config *gumble.Config
	Client *gumble.Client

	Address   string
	TLSConfig tls.Config

	ConnectAttempts uint

	Stream *gumbleopenal.Stream

	ChannelName string
	Logging     string
	Daemonize   string

	IsConnected    bool
	IsTransmitting bool

	GPIOEnabled     bool
	OnlineLED       gpio.Pin
	ParticipantsLED gpio.Pin
	TransmitLED     gpio.Pin
	TxButton        gpio.Pin
	TxButtonState   uint
	UpButton        gpio.Pin
	UpButtonState   uint
	DownButton      gpio.Pin
	DownButtonState uint
}

func (b *Talkkonnect) RogerBeep() {
	var stream *gumbleffmpeg.Stream
	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
		time.Sleep(2 * time.Second)
		return
	}
	stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile("/home/mumble/gocode/src/github.com/talkkonnect/talkkonnect/soundfiles/rogerbeep.wav"))
	if err := stream.Play(); err != nil {
		log.Println("alert: Can't Play File", err)
	} else {
		log.Println("info: Roger Beep Playing!")
	}
}

func EventDing() {
	log.Println("info: Ding!")
	// Open first sample File (Harded Code Path and File Should be removed to config later

	f, err := os.Open("/home/mumble/gocode/src/github.com/talkkonnect/talkkonnect/soundfiles/tone.wav")

	// Check for errors when opening the file
	if err != nil {
		log.Fatal("alert: Fatal Error ", err)
	}

	defer f.Close()

	// Decode the .mp3 File, if you have a .wav file, use wav.Decode(f)
	s, format, _ := wav.Decode(f)

	// Init the Speaker with the SampleRate of the format and a buffer size of 1/10s
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	// Now we Play our Streamer on the Speaker
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing
}

func talkkonnectBanner() {
	log.Println("info: ┌────────────────────────────────────────────────────────────────┐")
	log.Println("info: │  _        _ _    _                               _             │")
	log.Println("info: │ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │")
	log.Println("info: │ | __/ _` | | |/ / |/ / _ \\| '_ \\| '_ \\ / _ \\/ __|  __|         │")
	log.Println("info: │ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │")
	log.Println("info: │  \\__\\__,_|_|_|\\_\\_|\\_\\___/|_| |_|_| |_|\\___|\\_ _|\\__|          │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │talKKonnect (http://www.talkkonnect.com)                        │")
	log.Println("info: │A Headless Mumble Transceiver With a LCD Screen and Up/Dwn Bttns│")
	log.Println("info: │Created By Suvir Kumar <suvir@talkkonnect.com>                  │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │Based on talkiepi By Daniel Chote and Barnard by Tim Cooper     │")
	log.Println("info: │Developed using The Go Programming Language and gumble libraries│")
	log.Println("info: │Version 1.00 Released August 2018                               │")
	log.Println("info: │Additional Modifications Released under MPL 2.0 License         │")
	log.Println("info: └────────────────────────────────────────────────────────────────┘")
	log.Println("info: Press the <Del> key for Menu or press <F10> to Quit talKKonnect")
}

func (b *Talkkonnect) talkkonnectMenu() {
	log.Println("info: ┌────────────────────────────────────────────────────────────────┐")
	log.Println("info: │                 _ __ ___   ___ _ __  _   _                     │")
	log.Println("info: │                | '_ ` _ \\ / _ \\ '_ \\| | | |                    │")
	log.Println("info: │                | | | | | |  __/ | | | |_| |                    │")
	log.Println("info: │                |_| |_| |_|\\___|_| |_|\\__,_|                    │")
	log.Println("info: │                                                                │")
	log.Println("info: ├────────────────────────────────────────────────────────────────┤")
	log.Println("info: │ <Del> to Display this Menu                                     │")
	log.Println("info: ├─────────────────────────────┬──────────────────────────────────┤")
	log.Println("info: │ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)           │")
	log.Println("info: │ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level       │")
	log.Println("info: │ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)    │")
	log.Println("info: │ <F7>  Channels Information  │ <F8>  Start Transmitting         │")
	log.Println("info: │ <F9>  Stop Transmitting     │ <F10> Terminate Program          │")
	log.Println("info: └─────────────────────────────┴──────────────────────────────────┘")

	log.Println("info: IP Address & Session Information")
	localAddresses()
	log.Println("info: Server ", b.Address)
	log.Println("info: Username ", b.Config.Username)
	log.Println("info: Logging Mode ", b.Logging)
	log.Println("info: Daemon Mode ", b.Daemonize)
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("error: localAddresses: %v\n", err.Error()))
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Print(fmt.Errorf("error: localAddresses: %v\n", err.Error()))
			continue
		}

		for _, a := range addrs {
			if i.Name != "lo" {
				log.Printf("%v %v\n", i.Name, a)
			}
		}
	}
}
