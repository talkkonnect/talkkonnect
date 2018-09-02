package talkkonnect

import (
	"crypto/tls"

	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/fatih/color"
	"github.com/suvirkumar/gpio"
	"github.com/suvirkumar/gumble/gumble"
	"github.com/suvirkumar/gumble/gumbleffmpeg"
	"github.com/suvirkumar/gumble/gumbleopenal"
	"log"
	"os"
	"time"
)

const (
	//  Modified By Suvir Kumar to Match GPIO Pins I used for my Hardware Implentation
	OnlineLEDPin       uint = 2  // GPIO 2 ->  Raspberry Pi Physical Pin 3
	ParticipantsLEDPin uint = 3  // GPIO 3 ->  Raspberry Pi Physical Pin 5
	TransmitLEDPin     uint = 4  // GPIO 4 ->  Raspberry Pi Physical Pin 7
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

	ChannelName    string
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
	stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile("/home/mumble/gocode/src/github.com/suvirkumar/talkkonnect/soundfiles/rogerbeep.wav"))
	if err := stream.Play(); err != nil {
		fmt.Printf("%s\n", err)
	} else {
		color.Yellow(time.Now().Format(time.Stamp) + " Event   : Roger Beep Playing!\n")
	}
}

func EventDing() {
	color.Yellow(time.Now().Format(time.Stamp) + " Event   : Ding!\n")

	// Open first sample File

	f, err := os.Open("/home/mumble/gocode/src/github.com/suvirkumar/talkkonnect/soundfiles/tone.wav")

	// Check for errors when opening the file
	if err != nil {
		log.Fatal(err)
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
	color.Set(color.FgGreen)
	fmt.Printf("┌────────────────────────────────────────────────────────────────┐\n")
	fmt.Printf("│  _        _ _    _                               _             │\n")
	fmt.Printf("│ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │\n")
	fmt.Printf("│ | __/ _` | | |/ / |/ / _ \\| '_ \\| '_ \\ / _ \\/ __|  __|         │\n")
	fmt.Printf("│ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │\n")
	fmt.Printf("│  \\__\\__,_|_|_|\\_\\_|\\_\\___/|_| |_|_| |_|\\___|\\_ _|\\__|          │\n")
	fmt.Printf("├────────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│TalKKonect (http://www.talkkonnect.com)                         │\n")
	fmt.Printf("│A Headless Mumble Transceiver With a LCD Screen and Up/Dwn Bttns│\n")
	color.Set(color.FgCyan)
	fmt.Printf("│Created By Suvir Kumar <suvir@talkkonnect.com>                  │\n")
	color.Set(color.FgGreen)
	fmt.Printf("├────────────────────────────────────────────────────────────────┤\n")
	fmt.Printf("│Based on talkiepi By Daniel Chote and Barnard by Tim Cooper     │\n")
	fmt.Printf("│Developed using The Go Programming Language and gumble libraries│\n")
	fmt.Printf("│Version 1.00 Released August 2018                               │\n")
	fmt.Printf("│Additional Modifications Released under MPL 2.0 License         │\n")
	fmt.Printf("└────────────────────────────────────────────────────────────────┘\n")
	color.Unset()
}
