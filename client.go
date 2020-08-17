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
 *
 */

package talkkonnect

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/comail/colog"
	"github.com/kennygrant/sanitize"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/gumble/gumbleutil"
	_ "github.com/talkkonnect/gumble/opus"
	htgotts "github.com/talkkonnect/htgo-tts"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
)

var (
	LcdText              = [4]string{"nil", "nil", "nil", "nil"}
	currentChannelID     uint32
	prevChannelID        uint32
	prevParticipantCount int    = 0
	prevButtonPress      string = "none"
	maxchannelid         uint32
	origVolume           int
	tempVolume           int
	ConfigXMLFile        string
	GPSTime              string
	GPSDate              string
	GPSLatitude          float64
	GPSLongitude         float64
	Streaming            bool
	AccountIndex         int = 0
	ServerHop            bool
	httpServRunning      bool
	message              string
	isrepeattx           bool = true
	NowStreaming         bool
)

type Talkkonnect struct {
	Config *gumble.Config
	Client *gumble.Client

	Name      string
	Address   string
	Username  string
	Ident     string
	TLSConfig tls.Config

	ConnectAttempts uint

	Stream *Stream

	ChannelName string
	Logging     string
	Daemonize   bool

	IsTransmitting bool
	IsPlayStream   bool

	GPIOEnabled        bool
	OnlineLED          gpio.Pin
	ParticipantsLED    gpio.Pin
	TransmitLED        gpio.Pin
	HeartBeatLED       gpio.Pin
	BackLightLED       gpio.Pin
	VoiceActivityLED   gpio.Pin
	TxButton           gpio.Pin
	TxButtonState      uint
	TxToggle           gpio.Pin
	TxToggleState      uint
	UpButton           gpio.Pin
	UpButtonState      uint
	DownButton         gpio.Pin
	DownButtonState    uint
	PanicButton        gpio.Pin
	PanicButtonState   uint
	CommentButton      gpio.Pin
	CommentButtonState uint
	ChimesButton       gpio.Pin
	ChimesButtonState  uint
}

type ChannelsListStruct struct {
	chanID     uint32
	chanName   string
	chanParent *gumble.Channel
	chanUsers  int
}

func reset() {
	term.Sync()
}

func PreInit0(file string) {
	err := term.Init()
	if err != nil {
		log.Println("alert: Cannot Initalize Terminal Error: ", err)
		log.Fatal("Exiting talkkonnect! ...... bye!\n")
	}

	colog.Register()
	colog.SetOutput(os.Stdout)

	ConfigXMLFile = file
	err = readxmlconfig(ConfigXMLFile)
	if err != nil {
		log.Println("alert: XML Parser Module Returned Error: ", err)
		log.Fatal("Please Make Sure the XML Configuration File is In the Correct Path with the Correct Format, Exiting talkkonnect! ...... bye\n")
	}

	if APEnabled {
		log.Println("info: Contacting http Provisioning Server Pls Wait")
		err := autoProvision()
		time.Sleep(5 * time.Second)
		if err != nil {
			log.Println("alert: Error from AutoProvisioning Module: ", err)
			log.Println("alert: Please Fix Problem with Provisioning Configuration or use Static File By Disabling AutoProvisioning ")
			log.Fatal("Exiting talkkonnect! ...... bye\n")
		} else {
			log.Println("info: Got New Configuration Reloading XML Config")
			ConfigXMLFile = file
			readxmlconfig(ConfigXMLFile)
		}
	}

	b := Talkkonnect{
		Config:      gumble.NewConfig(),
		Name:        Name[AccountIndex],
		Address:     Server[AccountIndex],
		Username:    Username[AccountIndex],
		Ident:       Ident[AccountIndex],
		ChannelName: Channel[AccountIndex],
		Logging:     Logging,
		Daemonize:   Daemonize,
	}

	b.PreInit1(false)
}

func (b *Talkkonnect) PreInit1(httpServRunning bool) {
	if len(b.Username) == 0 {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			log.Println("alert: Cannot Generate Random Name Error: ", err)
			log.Fatal("Exiting talkkonnect! ...... bye!\n")
		}

		buf[0] |= 2
		b.Config.Username = fmt.Sprintf("talkkonnect-%02x%02x%02x%02x%02x%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	} else {
		b.Config.Username = Username[AccountIndex]
	}

	b.Config.Password = Password[AccountIndex]

	if Insecure[AccountIndex] {
		b.TLSConfig.InsecureSkipVerify = true
	}
	if Certificate[AccountIndex] != "" {
		cert, err := tls.LoadX509KeyPair(Certificate[AccountIndex], Certificate[AccountIndex])
		if err != nil {
			log.Println("alert: Certificate Error: ", err)
			log.Fatal("Exiting talkkonnect! ...... bye!\n")
		}
		b.TLSConfig.Certificates = append(b.TLSConfig.Certificates, cert)
	}

	if APIEnabled && !httpServRunning {
		go func() {
			http.HandleFunc("/", b.httpHandler)

			if err := http.ListenAndServe(":"+APIListenPort, nil); err != nil {
				log.Println("alert: Problem With Starting HTTP API Server Error: ", err)
				log.Fatal("Please Fix Problem or Disable API in XML Config, Exiting talkkonnect! ...... bye!\n")
			}
		}()
	}

	b.Init()
	IsConnected = false

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitStatus := 0

	<-sigs
	b.CleanUp()
	os.Exit(exitStatus)
}

func (b *Talkkonnect) Init() {
	f, err := os.OpenFile(LogFilenameAndPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.Println("info: Trying to Open File ", LogFilenameAndPath)
	if err != nil {
		log.Println("alert: Problem opening talkkonnect.log file Error: ", err)
		log.Fatal("Exiting talkkonnect! ...... bye!\n")
	}

	if TargetBoard == "rpi" {
		b.LEDOffAll()
	}

	if b.Logging == "screen" {
		colog.SetOutput(os.Stdout)
	} else {
		wrt := io.MultiWriter(os.Stdout, f)
		log.SetOutput(wrt)
	}

	b.Config.Attach(gumbleutil.AutoBitrate)
	b.Config.Attach(b)

	log.Printf("info: %d Default Mumble Accounts Found\n", AccountCount)
	if TargetBoard == "rpi" {
		log.Println("info: Target Board Set as RPI (gpio enabled) ")
		b.initGPIO()
	} else {
		log.Println("info: Target Board Set as PC (gpio disabled) ")
	}

	talkkonnectBanner()

	err = volume.Unmute(OutputDevice)
	if err != nil {
		log.Println("warn: Unable to Unmute ", err)
	} else {
		log.Println("info: Speaker UnMuted Before Connect to Server")
	}

	if TTSEnabled && TTSTalkkonnectLoaded {
		err := PlayWavLocal(TTSTalkkonnectLoadedFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSTalkkonnectLoadedFilenameAndPath) Returned Error: ", err)
		}
	}

	b.Connect()

	pstream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(""), 0)

	if HeartBeatEnabled && TargetBoard == "rpi" {
		HeartBeat := time.NewTicker(time.Duration(PeriodmSecs) * time.Millisecond)

		go func() {
			for _ = range HeartBeat.C {
				timer1 := time.NewTimer(time.Duration(LEDOnmSecs) * time.Millisecond)
				timer2 := time.NewTimer(time.Duration(LEDOffmSecs) * time.Millisecond)
				<-timer1.C
				b.LEDOn(b.HeartBeatLED)
				<-timer2.C
				b.LEDOff(b.HeartBeatLED)

				if KillHeartBeat == true {
					HeartBeat.Stop()
				}

			}
		}()
	}

	if BeaconEnabled {
		BeaconTicker := time.NewTicker(time.Duration(BeaconTimerSecs) * time.Second)

		go func() {
			for _ = range BeaconTicker.C {
				IsPlayStream = true
				b.playIntoStream(BeaconFilenameAndPath, BVolume)
				IsPlayStream = false
				log.Println("warn: Beacon Enabled and Timed Out Auto Played File ", BeaconFilenameAndPath, " Into Stream")
			}
		}()
	}

	b.BackLightTimer()

	if OLEDEnabled == true {
		Oled.DisplayOn()
		LCDIsDark = false
	}

	if AudioRecordEnabled == true {

		if AudioRecordOnStart == true {

			if AudioRecordMode != "" {

				if AudioRecordMode == "traffic" {
					log.Println("info: Incoming Traffic will be Recorded with sox")
					AudioRecordTraffic()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "nil", "Traffic Recording ->"} // 4
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 6, 1, "Traffic Recording") // 6
						}
					}
				}
				if AudioRecordMode == "ambient" {
					log.Println("info: Ambient Audio from Mic will be Recorded with sox")
					AudioRecordAmbient()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "nil", "Mic Recording ->"} // 4
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 6, 1, "Mic Recording") // 6
						}
					}
				}
				if AudioRecordMode == "combo" {
					log.Println("info: Both Incoming Traffic and Ambient Audio from Mic will be Recorded with sox")
					AudioRecordCombo()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "nil", "Combo Recording ->"} // 4
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 6, 1, "Combo Recording") //6
						}
					}
				}

			}

		}
	}

keyPressListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				log.Println("--")
				log.Println("warn: ESC Key is Invalid")
				reset()
				break keyPressListenerLoop
				log.Println("--")
			case term.KeyDelete:
				b.commandKeyDel()
			case term.KeyF1:
				b.commandKeyF1()
			case term.KeyF2:
				b.commandKeyF2()
			case term.KeyF3:
				b.commandKeyF3()
			case term.KeyF4:
				b.commandKeyF4()
			case term.KeyF5:
				b.commandKeyF5()
			case term.KeyF6:
				b.commandKeyF6()
			case term.KeyF7:
				b.commandKeyF7()
			case term.KeyF8:
				b.commandKeyF8()
			case term.KeyF9:
				b.commandKeyF9()
			case term.KeyF10:
				b.commandKeyF10()
			case term.KeyF11:
				b.commandKeyF11()
			case term.KeyF12:
				b.commandKeyF12()
			case term.KeyCtrlC:
				talkkonnectAcknowledgements()
				b.commandKeyCtrlC()
			case term.KeyCtrlE:
				b.commandKeyCtrlE()
			case term.KeyCtrlF:
				b.commandKeyCtrlF()
			case term.KeyCtrlI: // New. Audio Recording. Traffic
				b.commandKeyCtrlI()
			case term.KeyCtrlJ: // New. Audio Recording. Mic
				b.commandKeyCtrlJ()
			case term.KeyCtrlK: // New/ Audio Recording. Combo
				b.commandKeyCtrlK()
			case term.KeyCtrlL:
				b.commandKeyCtrlL()
			case term.KeyCtrlO:
				b.commandKeyCtrlO()
			case term.KeyCtrlN:
				b.commandKeyCtrlN()
			case term.KeyCtrlP:
				b.commandKeyCtrlP()
			case term.KeyCtrlR:
				b.commandKeyCtrlR()
			case term.KeyCtrlS:
				b.commandKeyCtrlS()
			case term.KeyCtrlT:
				b.commandKeyCtrlT()
			case term.KeyCtrlV:
				b.commandKeyCtrlV()
			case term.KeyCtrlX:
				b.commandKeyCtrlX()
			default:
				log.Println("--")
				if ev.Ch != 0 {
					log.Println("warn: Invalid Keypress ASCII", ev.Ch)
				} else {
					log.Println("warn: Key Not Mapped")
				}
				log.Println("--")
			}
		case term.EventError:
			log.Println("alert: Terminal Error: ", ev.Err)
			log.Fatal("Exiting talkkonnect! ...... bye!\n")
		}

	}

}

func (b *Talkkonnect) CleanUp() {
	log.Println("warn: SIGHUP Termination of Program Requested...shutting down...bye!")

	if TargetBoard == "rpi" {
		t := time.Now()
		if LCDEnabled == true {
			LcdText = [4]string{"talkkonnect stopped", t.Format("02-01-2006 15:04:05"), "Please Visit", "www.talkkonnect.com"}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 1, "talkkonnect stopped")
			oledDisplay(false, 1, 1, t.Format("02-01-2006 15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
		b.LEDOffAll()
	}

	//b.Client.Disconnect()
	c := exec.Command("reset")
	c.Stdout = os.Stdout
	c.Run()
	os.Exit(0)
}

func (b *Talkkonnect) Connect() {
	IsConnected = false
	IsPlayStream = false
	NowStreaming = false
	KillHeartBeat = false
	var err error

	_, err = gumble.DialWithDialer(new(net.Dialer), b.Address, b.Config, &b.TLSConfig)

	if err != nil {
		log.Printf("warn: Connection Error %v  connecting to %v failed, attempting again...", err, b.Address)
		if !ServerHop {
			log.Println("warn: In the Connect Function & Trying With Username ", Username)
			log.Println("warn: DEBUG Serverhop  Not Set Reconnecting!!")
			b.ReConnect()
		}
	} else {
		b.OpenStream()
	}
}

func (b *Talkkonnect) ReConnect() {
	IsConnected = false
	IsPlayStream = false
	NowStreaming = false

	if b.Client != nil {
		log.Println("warn: Attempting Reconnection With Server")
		b.Client.Disconnect()
	}

	if ConnectAttempts < 3 {
		//go func() {
		if !ServerHop {
			ConnectAttempts++
			b.Connect()
		}
		//}()
	} else {
		log.Println("warn: Unable to connect, giving up")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"Failed to Connect!", "nil", "nil", "nil"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 2, 1, "Failed to Connect!")
			}
		}
		log.Fatal("Exiting talkkonnect! ...... bye!\n")
	}
}

func (b *Talkkonnect) OpenStream() {

	if os.Getenv("ALSOFT_LOGLEVEL") == "" {
		os.Setenv("ALSOFT_LOGLEVEL", "0")
	}

	if ServerHop {
		log.Println("warn: Server Hop Requested Will Now Destroy Old Server Stream")
		b.Stream.Destroy()
	}

	if stream, err := New(b.Client); err != nil {

		log.Println("warn: Stream open error ", err)
		if TargetBoard == "pi" {
			if LCDEnabled == true {
				LcdText = [4]string{"Stream Error!", "nil", "nil", "nil"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 2, 1, "Stream Error!!")
			}

		}
		log.Fatal("Exiting talkkonnect! ...... bye!\n")
	} else {
		b.Stream = stream
	}
}

func (b *Talkkonnect) ResetStream() {
	b.Stream.Destroy()
	time.Sleep(50 * time.Millisecond)
	b.OpenStream()
}

func (b *Talkkonnect) TransmitStart() {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()
	t := time.Now()

	if SimplexWithMute {
		err := volume.Mute(OutputDevice)
		if err != nil {
			log.Println("warn: Unable to Mute ", err)
		} else {
			log.Println("info: Speaker Muted ")
		}
	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false
		time.Sleep(100 * time.Millisecond)
		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if TargetBoard == "rpi" {
		b.LEDOn(b.TransmitLED)
		if LCDEnabled == true {
			LcdText[0] = "Online/TX"
			LcdText[3] = "TX at " + t.Format("15:04:05")
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			//oledDisplay(true, 0, 0, "") // clear the screen
			oledDisplay(false, 0, 1, "Online/TX")
			oledDisplay(false, 3, 1, "TX at "+t.Format("15:04:05"))
			oledDisplay(false, 6, 1, "Please Visit       ")
			oledDisplay(false, 7, 1, "www.talkkonnect.com")
		}
	}

	b.IsTransmitting = true

	if RepeaterToneEnabled {
		b.RepeaterTone(RepeaterToneFilenameAndPath, RepeaterToneVolume)
	}

	if pstream.State() == gumbleffmpeg.StatePlaying {
		pstream.Stop()
	}

	b.Stream.StartSource()

}

func (b *Talkkonnect) TransmitStop(withBeep bool) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	if TargetBoard == "rpi" {
		b.LEDOff(b.TransmitLED)

		if LCDEnabled == true {
			LcdText[0] = b.Address
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			oledDisplay(false, 0, 1, b.Address)
		}
	}

	b.IsTransmitting = false
	b.Stream.StopSource()

	if SimplexWithMute {
		err := volume.Unmute(OutputDevice)
		if err != nil {
			log.Println("warn: Unable to Unmute ", err)
		} else {
			log.Println("info: Speaker UnMuted ")
		}
	}
}

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	if IsConnected == true {
		return
	}

	IsConnected = true
	b.BackLightTimer()
	b.Client = e.Client
	ConnectAttempts = 1

	log.Printf("info: Connected to %s Address %s on attempt %d index [%d]\n ", b.Name, b.Client.Conn.RemoteAddr(), b.ConnectAttempts, AccountIndex)
	if e.WelcomeMessage != nil {
		var message string = fmt.Sprintf(esc(*e.WelcomeMessage))
		log.Println("info: Welcome message: ")
		for _, line := range strings.Split(strings.TrimSuffix(message, "\n"), "\n") {
			log.Println("info: ", line)
		}
	}

	if TargetBoard == "rpi" {
		b.LEDOn(b.OnlineLED)

		if LCDEnabled == true {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 0, "") // clear the screen
		}

		b.ParticipantLEDUpdate(true)
	}

	if b.ChannelName != "" {
		b.ChangeChannel(b.ChannelName)
		prevChannelID = b.Client.Self.Channel.ID
	}
}

func (b *Talkkonnect) OnDisconnect(e *gumble.DisconnectEvent) {
	if !ServerHop {
		b.BackLightTimer()
	}

	var reason string

	switch e.Type {
	case gumble.DisconnectError:
		reason = "connection error"
	}

	IsConnected = false

	if TargetBoard == "rpi" {
		b.LEDOff(b.OnlineLED)
		b.LEDOff(b.ParticipantsLED)
		b.LEDOff(b.TransmitLED)
	}

	if !ServerHop {
		log.Println("warn: Attempting Reconnect in 10 seconds...")
		log.Println("warn: Connection to ", b.Address, "disconnected")
		log.Println("warn: Disconnection Reason ", reason)
		b.ReConnect()
	}

}

func (b *Talkkonnect) ChangeChannel(ChannelName string) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	channel := b.Client.Channels.Find(ChannelName)
	if channel != nil {

		b.Client.Self.Move(channel)

		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText[1] = "Joined " + ChannelName
				LcdText[2] = Username[AccountIndex]
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 0, 1, "Joined "+ChannelName)
				oledDisplay(false, 1, 1, Username[AccountIndex])
			}
		}

		log.Println("info: Joined Channel Name: ", channel.Name, " ID ", channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		log.Println("warn: Unable to Find Channel Name: ", ChannelName)
		prevChannelID = 0
	}
}

func (b *Talkkonnect) ParticipantLEDUpdate(verbose bool) {
	if !(IsConnected) {
		return
	}

	b.BackLightTimer()

	var participantCount = len(b.Client.Self.Channel.Users)

	if participantCount != prevParticipantCount {
		if EventSoundEnabled {
			err := PlayWavLocal(EventSoundFilenameAndPath, 100)
			if err != nil {
				log.Println("warn: PlayWavLocal(EventSoundFilenameAndPath) Returned Error: ", err)
			}
		}
	}

	if participantCount > 1 && participantCount != prevParticipantCount {

		if TTSEnabled && TTSParticipants {
			speech := htgotts.Speech{Folder: "audio", Language: "en"}
			speech.Speak("There Are Currently " + strconv.Itoa(participantCount) + " Users in The Channel " + b.Client.Self.Channel.Name)
		}

		prevParticipantCount = participantCount

		if verbose {
			log.Println("info: Current Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")
			b.ListUsers()
			if TargetBoard == "rpi" {
				if LCDEnabled == true {
					LcdText[0] = b.Address
					LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
					go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
					oledDisplay(false, 0, 1, b.Address)
					oledDisplay(false, 1, 1, b.Client.Self.Channel.Name+" ("+strconv.Itoa(participantCount)+" Users)")
					oledDisplay(false, 6, 1, "Please Visit")
					oledDisplay(false, 7, 1, "www.talkkonnect.com")
				}

			}
		}
	}

	if participantCount > 1 {
		if TargetBoard == "rpi" {
			b.LEDOn(b.ParticipantsLED)
			b.LEDOn(b.OnlineLED)
		}

	} else {

		if verbose {
			if TTSEnabled && TTSParticipants {
				speech := htgotts.Speech{Folder: "audio", Language: "en"}
				speech.Speak("You are Currently Alone in The Channel " + b.Client.Self.Channel.Name)
			}
			log.Println("info: Channel ", b.Client.Self.Channel.Name, " has no other participants")

			prevParticipantCount = 0

			if TargetBoard == "rpi" {

				b.LEDOff(b.ParticipantsLED)

				if LCDEnabled == true {
					LcdText = [4]string{b.Address, "Alone in " + b.Client.Self.Channel.Name, "", "nil"}
					go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
					oledDisplay(false, 0, 1, b.Address)
					oledDisplay(false, 1, 1, "Alone in "+b.Client.Self.Channel.Name)
				}
			}
		}
	}
}

func (b *Talkkonnect) OnTextMessage(e *gumble.TextMessageEvent) {
	b.BackLightTimer()

	if len(cleanstring(e.Message)) > 105 {
		log.Println(fmt.Sprintf("warn: Message Too Long to Be Displayed on Screen\n"))
		message = strings.TrimSpace(cleanstring(e.Message)[:105])
	} else {
		message = strings.TrimSpace(cleanstring(e.Message))
	}

	var sender string

	if e.Sender != nil {
		sender = strings.TrimSpace(cleanstring(e.Sender.Name))
		log.Println("alert: Sender Name is ", sender)
	} else {
		sender = ""
	}

	log.Println(fmt.Sprintf("alert: Message ("+strconv.Itoa(len(message))+") from %v %v\n", sender, message))

	if TargetBoard == "rpi" {
		if LCDEnabled == true {
			LcdText[0] = "Msg From " + sender
			LcdText[1] = message
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			oledDisplay(false, 2, 1, "Msg From "+sender)
			if len(message) <= 21 {
				oledDisplay(false, 3, 1, message)
				oledDisplay(false, 4, 1, "")
				oledDisplay(false, 5, 1, "")
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
			} else if len(message) <= 42 {
				oledDisplay(false, 3, 1, message[0:21])
				oledDisplay(false, 4, 1, message[21:len(message)])
				oledDisplay(false, 5, 1, "")
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
			} else if len(message) <= 63 {
				oledDisplay(false, 3, 1, message[0:21])
				oledDisplay(false, 4, 1, message[21:42])
				oledDisplay(false, 5, 1, message[42:len(message)])
				oledDisplay(false, 6, 1, "")
				oledDisplay(false, 7, 1, "")
			} else if len(message) <= 84 {
				oledDisplay(false, 3, 1, message[0:21])
				oledDisplay(false, 4, 1, message[21:42])
				oledDisplay(false, 5, 1, message[42:63])
				oledDisplay(false, 6, 1, message[63:len(message)])
				oledDisplay(false, 7, 1, "")
			} else if len(message) <= 105 {
				oledDisplay(false, 3, 1, message[0:20])
				oledDisplay(false, 4, 1, message[21:44])
				oledDisplay(false, 5, 1, message[42:63])
				oledDisplay(false, 6, 1, message[63:84])
				oledDisplay(false, 7, 1, message[84:105])
			}
		}
	}

	if EventSoundEnabled {
		err := PlayWavLocal(EventSoundFilenameAndPath, 100)
		if err != nil {
			log.Println("warn: PlayWavLocal(EventSoundFilenameAndPath) Returned Error: ", err)
		}
	}
}

func (b *Talkkonnect) OnUserChange(e *gumble.UserChangeEvent) {
	b.BackLightTimer()

	var info string

	switch e.Type {
	case gumble.UserChangeConnected:
		info = "conn"
	case gumble.UserChangeDisconnected:
		info = "disconnected!"
	case gumble.UserChangeKicked:
		info = "kicked"
	case gumble.UserChangeBanned:
		info = "banned"
	case gumble.UserChangeRegistered:
		info = "registered"
	case gumble.UserChangeUnregistered:
		info = "unregistered"
	case gumble.UserChangeName:
		info = "chg name"
	case gumble.UserChangeChannel:
		info = "chg channel"
		log.Println("info:", cleanstring(e.User.Name), " Changed Channel to ", e.User.Channel.Name)
		LcdText[2] = cleanstring(e.User.Name) + "->" + e.User.Channel.Name
		LcdText[3] = ""
	case gumble.UserChangeComment:
		info = "chg comment"
	case gumble.UserChangeAudio:
		info = "chg audio"
	case gumble.UserChangePrioritySpeaker:
		info = "is priority"
	case gumble.UserChangeRecording:
		info = "chg rec status"
	case gumble.UserChangeStats:
		info = "chg stats"

		if info != "chg channel" {
			if info != "" {
				log.Println("info: User ", cleanstring(e.User.Name), " ", info, "Event type=", e.Type, " channel=", e.User.Channel.Name)
				if TTSEnabled && TTSParticipants {
					speech := htgotts.Speech{Folder: "audio", Language: "en"}
					speech.Speak("User ")
				}
			}

		} else {
			log.Println("info: User ", cleanstring(e.User.Name), " Event type=", e.Type, " channel=", e.User.Channel.Name)
		}

		LcdText[2] = cleanstring(e.User.Name) + " " + info //+strconv.Atoi(string(e.Type))

	}

	b.ParticipantLEDUpdate(true)
}

func (b *Talkkonnect) OnPermissionDenied(e *gumble.PermissionDeniedEvent) {
	var info string

	switch e.Type {
	case gumble.PermissionDeniedOther:
		info = e.String
	case gumble.PermissionDeniedPermission:
		info = "insufficient permissions"
		LcdText[2] = "insufficient perms"

		// Set Upper Boundary
		if prevButtonPress == "ChannelUp" && b.Client.Self.Channel.ID == maxchannelid {
			log.Println("info: Can't Increment Channel Maximum Channel Reached")
		}

		// Set Lower Boundary
		if prevButtonPress == "ChannelDown" && currentChannelID == 0 {
			log.Println("info: Can't Increment Channel Minumum Channel Reached")
		}

		// Implement Seek Up Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelUp" && b.Client.Self.Channel.ID+1 < maxchannelid {
			prevChannelID++
			b.ChannelUp()
			LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
		}

		// Implement Seek Dwn Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelDown" && int(b.Client.Self.Channel.ID) > 0 {
			prevChannelID--
			b.ChannelDown()
			LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
		}

		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 1, 1, LcdText[1])
				oledDisplay(false, 2, 1, LcdText[2])
			}
		}

	case gumble.PermissionDeniedSuperUser:
		info = "cannot modify SuperUser"
	case gumble.PermissionDeniedInvalidChannelName:
		info = "invalid channel name"
	case gumble.PermissionDeniedTextTooLong:
		info = "text too long"
	case gumble.PermissionDeniedTemporaryChannel:
		info = "temporary channel"
	case gumble.PermissionDeniedMissingCertificate:
		info = "missing certificate"
	case gumble.PermissionDeniedInvalidUserName:
		info = "invalid user name"
	case gumble.PermissionDeniedChannelFull:
		info = "channel full"
	case gumble.PermissionDeniedNestingLimit:
		info = "nesting limit"
	}

	LcdText[2] = info

	log.Println("warn: Permission denied  ", info)
}

func (b *Talkkonnect) OnChannelChange(e *gumble.ChannelChangeEvent) {
}

func (b *Talkkonnect) OnUserList(e *gumble.UserListEvent) {
}

func (b *Talkkonnect) OnACL(e *gumble.ACLEvent) {
}

func (b *Talkkonnect) OnBanList(e *gumble.BanListEvent) {
}

func (b *Talkkonnect) OnContextActionChange(e *gumble.ContextActionChangeEvent) {
}

func (b *Talkkonnect) OnServerConfig(e *gumble.ServerConfigEvent) {
}

func (b *Talkkonnect) OnAudioStream(e *gumble.AudioStreamEvent) {
}

func esc(str string) string {
	return sanitize.HTML(str)
}

func cleanstring(str string) string {
	return sanitize.Name(str)
}

func (b *Talkkonnect) ListUsers() {
	if !(IsConnected) {
		return
	}

	item := 0
	for _, usr := range b.Client.Users {
		if usr.Channel.ID == b.Client.Self.Channel.ID {
			item++
			log.Println(fmt.Sprintf("info: %d. User %#v is online. [%v]", item, usr.Name, usr.Comment))
		}
	}

}

func (b *Talkkonnect) ListChannels(verbose bool) {
	if !(IsConnected) {
		return
	}

	var records = int(len(b.Client.Channels))
	var channelsList [100]ChannelsListStruct
	counter := 0

	for _, ch := range b.Client.Channels {
		channelsList[counter].chanID = ch.ID
		channelsList[counter].chanName = ch.Name
		channelsList[counter].chanParent = ch.Parent
		channelsList[counter].chanUsers = len(ch.Users)

		if ch.ID > maxchannelid {
			maxchannelid = ch.ID
		}

		counter++
	}

	for i := 0; i < int(records); i++ {
		if channelsList[i].chanID == 0 || channelsList[i].chanParent.ID == 0 {
			if verbose {
				log.Println(fmt.Sprintf("info: Parent -> ID=%2d | Name=%-12v (%v) Users | ", channelsList[i].chanID, channelsList[i].chanName, channelsList[i].chanUsers))
			}
		} else {
			if verbose {
				log.Println(fmt.Sprintf("info: Child  -> ID=%2d | Name=%-12v (%v) Users | PID =%2d | PName=%-12s", channelsList[i].chanID, channelsList[i].chanName, channelsList[i].chanUsers, channelsList[i].chanParent.ID, channelsList[i].chanParent.Name))
			}
		}
	}

}

func (b *Talkkonnect) ChannelUp() {
	if !(IsConnected) {
		return
	}

	if prevChannelID == 0 {
		prevChannelID = b.Client.Self.Channel.ID
	}

	if TTSEnabled && TTSChannelUp {
		err := PlayWavLocal(TTSChannelUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelUp"

	b.ListChannels(false)

	// Set Upper Boundary
	if b.Client.Self.Channel.ID == maxchannelid {
		log.Println("info: Can't Increment Channel Maximum Channel Reached")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText[2] = "Max Chan Reached"
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 1, 1, "Max Chan Reached")
			}

		}
		return
	}

	// Implement Seek Up Avoiding any null channels
	if prevChannelID < maxchannelid {

		prevChannelID++

		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {

			channel := b.Client.Channels[i]

			if channel != nil {
				b.Client.Self.Move(channel)
				//displaychannel
				time.Sleep(500 * time.Millisecond)
				if TargetBoard == "rpi" {

					if len(b.Client.Self.Channel.Users) == 1 {
						LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
					} else {
						LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
					}

					if LCDEnabled == true {
						go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled == true {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}
				break
			}
		}
	}
	return
}

func (b *Talkkonnect) ChannelDown() {
	if !(IsConnected) {
		return
	}

	if prevChannelID == 0 {
		prevChannelID = b.Client.Self.Channel.ID
	}

	if TTSEnabled && TTSChannelDown {
		err := PlayWavLocal(TTSChannelDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSChannelDownFilenameAndPath) Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelDown"
	b.ListChannels(false)

	// Set Lower Boundary
	if int(b.Client.Self.Channel.ID) == 0 {
		log.Println("info: Can't Decrement Channel Root Channel Reached")
		channel := b.Client.Channels[uint32(AccountIndex)]
		b.Client.Self.Move(channel)
		//displaychannel
		time.Sleep(500 * time.Millisecond)
		if TargetBoard == "rpi" {

			if len(b.Client.Self.Channel.Users) == 1 {
				LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
			} else {
				LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
			}

			if LCDEnabled == true {
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 1, 1, LcdText[1])
			}
		}

		return
	}

	// Implement Seek Down Avoiding any null channels
	if int(prevChannelID) > 0 {

		prevChannelID--

		for i := uint32(prevChannelID); uint32(i) < maxchannelid; i-- {
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				//displaychannel
				time.Sleep(500 * time.Millisecond)
				if TargetBoard == "rpi" {

					if len(b.Client.Self.Channel.Users) == 1 {
						LcdText[1] = "Alone in " + b.Client.Self.Channel.Name
					} else {
						LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(len(b.Client.Self.Channel.Users)) + " Users)"
					}

					if LCDEnabled == true {
						go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled == true {
						oledDisplay(false, 1, 1, LcdText[1])
					}
				}

				break
			}
		}
	}
	return
}

func (b *Talkkonnect) Scan() {
	if !(IsConnected) {
		return
	}

	b.ListChannels(false)

	if b.Client.Self.Channel.ID+1 > maxchannelid {
		prevChannelID = 0
		channel := b.Client.Channels[prevChannelID]
		b.Client.Self.Move(channel)
		return
	}

	if prevChannelID < maxchannelid {
		prevChannelID++

		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				time.Sleep(1000 * time.Millisecond)
				if len(b.Client.Self.Channel.Users) == 1 {
					b.Scan()
					break
				} else {

					log.Println("warn: Found Someone Online Stopped Scan on Channel ", b.Client.Self.Channel.Name)
					return
				}
			}
		}
	}
	return
}

func (b *Talkkonnect) httpHandler(w http.ResponseWriter, r *http.Request) {
	commands, ok := r.URL.Query()["command"]
	if !ok || len(commands[0]) < 1 {
		log.Println("warn: URL Param 'command' is missing")
		return
	}

	command := commands[0]
	log.Println("info: http command " + string(command))
	b.BackLightTimer()

	switch string(command) {
	case "DEL":
		if APIDisplayMenu {
			b.commandKeyDel()
			fmt.Fprintf(w, "API Display Menu Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Display Menu Request Denied\n")
		}
	case "F1":
		if APIChannelUp {
			b.commandKeyF1()
			fmt.Fprintf(w, "API Channel Up Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Channel Up Request Denied\n")
		}
	case "F2":
		if APIChannelDown {
			b.commandKeyF2()
			fmt.Fprintf(w, "API Channel Down Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Channel Down Request Denied\n")
		}
	case "F3":
		if APIMute {
			b.commandKeyF3()
			fmt.Fprintf(w, "API Mute/UnMute Speaker Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Mute/Unmute Speaker Request Denied\n")
		}
	case "F4":
		if APICurrentVolumeLevel {
			b.commandKeyF4()
			fmt.Fprintf(w, "API Current Volume Level Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Current Volume Level Request Denied\n")
		}
	case "F5":
		if APIDigitalVolumeUp {
			b.commandKeyF5()
			fmt.Fprintf(w, "API Digital Volume Up Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Digital Volume Up Request Denied\n")
		}
	case "F6":
		if APIDigitalVolumeDown {
			b.commandKeyF6()
			fmt.Fprintf(w, "API Digital Volume Down Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Digital Volume Down Request Denied\n")
		}
	case "F7":
		if APIListServerChannels {
			b.commandKeyF7()
			fmt.Fprintf(w, "API List Server Channels Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API List Server Channels Request Denied\n")
		}
	case "F8":
		if APIStartTransmitting {
			b.commandKeyF8()
			fmt.Fprintf(w, "API Start Transmitting Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Start Transmitting Request Denied\n")
		}
	case "F9":
		if APIStopTransmitting {
			b.commandKeyF9()
			fmt.Fprintf(w, "API Stop Transmitting Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Stop Transmitting Request Denied\n")
		}
	case "F10":
		if APIListOnlineUsers {
			b.commandKeyF10()
			fmt.Fprintf(w, "API List Online Users Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API List Online Users Request Denied\n")
		}
	case "F11":
		if APIPlayChimes {
			b.commandKeyF11()
			fmt.Fprintf(w, "API Play/Stop Chimes Request Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Play/Stop Chimes Request Denied\n")
		}
	case "F12":
		if APIRequestGpsPosition {
			b.commandKeyF12()
			fmt.Fprintf(w, "API Request GPS Position Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request GPS Position Denied\n")
		}
	case "commandKeyCtrlE":
		if APIEmailEnabled {
			b.commandKeyCtrlE()
			fmt.Fprintf(w, "API Send Email Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Send Email Config Denied\n")
		}
	case "commandKeyCtrlF":
		if APINextServer {
			b.commandKeyCtrlF()
			fmt.Fprintf(w, "API Previous Server Processed Successfully\n")
		} else {
			fmt.Fprintf(w, "API Previous Server Denied\n")
		}
	case "commandKeyCtrlN":
		if APINextServer {
			b.commandKeyCtrlN()
			fmt.Fprintf(w, "API Next Server Processed Successfully\n")
		} else {
			fmt.Fprintf(w, "API Next Server Denied\n")
		}

	case "commandKeyCtrlL":
		if APIClearScreen {
			b.commandKeyCtrlL()
			fmt.Fprintf(w, "API Clear Screen Processed Successfully\n")
		} else {
			fmt.Fprintf(w, "API Clear Screen Denied\n")
		}
	case "commandKeyCtrlO":
		if APIEmailEnabled {
			b.commandKeyCtrlO()
			fmt.Fprintf(w, "API Ping Servers Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Ping Servers Denied\n")
		}
	case "commandKeyCtrlP":
		if APIPanicSimulation {
			b.commandKeyCtrlP()
			fmt.Fprintf(w, "API Request Panic Simulation Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Panic Simulation Denied\n")
		}
	case "commandKeyCtrlR":
		if APIRepeatTxLoopTest {
			b.commandKeyCtrlR()
			fmt.Fprintf(w, "API Request Repeat Tx Loop Test Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Repeat Tx Loop Test Denied\n")
		}
	case "commandKeyCtrlS":
		if APIScanChannels {
			b.commandKeyCtrlS()
			fmt.Fprintf(w, "API Request Scan Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Scan Denied\n")
		}
	case "commandKeyCtrlT":
		if true {
			b.commandKeyCtrlT()
			fmt.Fprintf(w, "API Request Show Acknowledgements Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Show Acknowledgements Denied\n")
		}
	case "commandKeyCtrlV":
		if APIDisplayVersion {
			b.commandKeyCtrlV()
			fmt.Fprintf(w, "API Request Current Version Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Current Version Denied\n")
		}
	case "commandKeyCtrlX":
		if APIPrintXmlConfig {
			b.commandKeyCtrlX()
			fmt.Fprintf(w, "API Print XML Config Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Print XML Congfig Denied\n")
		}
	default:
		fmt.Fprintf(w, "API Command Not Defined\n")
	}
}

func (b *Talkkonnect) commandKeyDel() {
	log.Println("--")
	log.Println("info: Delete Key Pressed Menu and Session Information Requested")

	if TTSEnabled && TTSDisplayMenu {
		err := PlayWavLocal(TTSDisplayMenuFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSDisplayMenuFilenameAndPath) Returned Error: ", err)
		}

	}

	b.talkkonnectMenu()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF1() {
	log.Println("--")
	log.Println("info: F1 pressed Channel Up (+) Requested")
	b.ChannelUp()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF2() {
	log.Println("--")
	log.Println("info: F2 pressed Channel Down (-) Requested")
	b.ChannelDown()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF3() {
	log.Println("--")
	log.Println("info: ", TTSMuteUnMuteSpeakerFilenameAndPath)

	origMuted, err := volume.GetMuted(OutputDevice)

	if err != nil {
		log.Println("warn: get muted failed: %+v", err)
	}

	if origMuted {
		err := volume.Unmute(OutputDevice)

		if err != nil {
			log.Println("warn: unmute failed: %+v", err)
		}

		log.Println("info: F3 pressed Mute/Unmute Speaker Requested Now UnMuted")
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("warn: PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath) Returned Error: ", err)
			}

		}
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Unmuted")
			}

		}
	} else {
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("warn: PlayWavLocal(TTSMuteUnMuteSpeakerFilenameAndPath) Returned Error: ", err)
			}

		}
		err = volume.Mute(OutputDevice)
		if err != nil {
			log.Println("warn: Mute failed: %+v", err)
		}

		log.Println("info: F3 pressed Mute/Unmute Speaker Requested Now Muted")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Muted"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Muted")
			}

		}
	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF4() {
	log.Println("--")
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("warn: Unable to get current volume: %+v", err)
	}

	log.Println("info: F4 pressed Volume Level Requested and is at", origVolume, "%")

	if TTSEnabled && TTSCurrentVolumeLevel {
		err := PlayWavLocal(TTSCurrentVolumeLevelFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSCurrentVolumeLevelFilenameAndPath) Returned Error: ", err)
		}

	}
	if TargetBoard == "rpi" {
		if LCDEnabled == true {
			LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(origVolume)}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled == true {
			oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(origVolume))
		}

	}
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF5() {
	log.Println("--")
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("warn: unable to get original volume: %+v", err)
	}

	if origVolume < 100 {
		err := volume.IncreaseVolume(+1, OutputDevice)
		if err != nil {
			log.Println("warn: F5 Increase Volume Failed! ", err)
		}

		log.Println("info: F5 pressed Volume UP (+) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Volume + " + strconv.Itoa(origVolume)}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Volume "+strconv.Itoa(origVolume))
			}
		}
	} else {
		log.Println("info: F5 Increase Volume Already at Maximum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Max Vol"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Max Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeUp {
		err := PlayWavLocal(TTSDigitalVolumeUpFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSDigitalVolumeUpFilenameAndPath) Returned Error: ", err)
		}

	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF6() {
	log.Println("--")
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("warn: unable to get original volume: %+v", err)
	}

	if origVolume > 0 {
		origVolume--
		err := volume.IncreaseVolume(-1, OutputDevice)
		if err != nil {
			log.Println("warn: F6 Decrease Volume Failed! ", err)
		}

		log.Println("info: F6 pressed Volume Down (-) Now At ", origVolume, "%")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Volume - " + strconv.Itoa(origVolume)}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Volume -")
			}

		}
	} else {
		log.Println("info: F6 Increase Volume Already at Minimum Possible Volume")
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText = [4]string{"nil", "nil", "nil", "Min Vol"}
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 6, 1, "Min Vol")
			}
		}
	}

	if TTSEnabled && TTSDigitalVolumeDown {
		err := PlayWavLocal(TTSDigitalVolumeDownFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSDigitalVolumeDownFilenameAndPath) Returned Error: ", err)
		}

	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF7() {
	log.Println("--")
	log.Println("info: F7 pressed Channel List Requested")

	if TTSEnabled && TTSListServerChannels {
		err := PlayWavLocal(TTSListServerChannelsFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSListServerChannelsFilenameAndPath) Returned Error: ", err)
		}

	}

	b.ListChannels(true)
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF8() {
	log.Println("--")
	log.Println("info: F8 pressed TX Mode Requested (Start Transmitting)")

	if TTSEnabled && TTSStartTransmitting {
		err := PlayWavLocal(TTSStartTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSStartTransmittingFilenameAndPath) Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if !b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStart()
	} else {
		log.Println("warn: Already in Transmitting Mode")
	}
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF9() {
	log.Println("--")
	log.Println("info: F9 pressed RX Mode Request (Stop Transmitting)")

	if TTSEnabled && TTSStopTransmitting {
		err := PlayWavLocal(TTSStopTransmittingFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: Play Wav Local Module Returned Error: ", err)
		}

	}

	if IsPlayStream {
		IsPlayStream = false
		NowStreaming = false

		b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}

	if b.IsTransmitting {
		time.Sleep(100 * time.Millisecond)
		b.TransmitStop(true)
	} else {
		log.Println("warn: Already Stopped Transmitting")
	}
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF10() {
	log.Println("--")
	log.Println("info: F10 pressed Online User(s) in Current Channel Requested")

	if TTSEnabled && TTSListOnlineUsers {
		err := PlayWavLocal(TTSListOnlineUsersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSListOnlineUsersFilenameAndPath) Returned Error: ", err)
		}

	}

	log.Println(fmt.Sprintf("info: Channel %#v Has %d Online User(s)", b.Client.Self.Channel.Name, len(b.Client.Self.Channel.Users)))
	b.ListUsers()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF11() {
	log.Println("--")
	log.Println("info: F11 pressed Start/Stop Chimes Stream into Current Channel Requested")

	b.BackLightTimer()

	if TTSEnabled && TTSPlayChimes {
		err := PlayWavLocal(TTSPlayChimesFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSPlayChimesFilenameAndPath) Returned Error: ", err)

		}

	}

	IsPlayStream = !IsPlayStream
	NowStreaming = IsPlayStream

	if IsPlayStream {
		b.SendMessage(fmt.Sprintf("%s Streaming", b.Username), false)
	}

	go b.playIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)

}

func (b *Talkkonnect) commandKeyF12() {
	log.Println("--")
	log.Println("info: F12 pressed GPS details requested")

	if TTSEnabled && TTSRequestGpsPosition {
		err := PlayWavLocal(TTSRequestGpsPositionFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSRequestGpsPositionFilenameAndPath) Returned Error: ", err)
		}

	}

	var i int = 0
	var tries int = 10

	for i = 0; i < tries; i++ {
		goodGPSRead, err := getGpsPosition(true)

		if err != nil {
			log.Println("warn: GPS Function Returned Error Message", err)
			break
		}

		if goodGPSRead == true {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlC() {
	log.Println("--")
	log.Println("info: Ctrl-C Terminate Program Requested")

	if TTSEnabled && TTSQuitTalkkonnect {
		err := PlayWavLocal(TTSQuitTalkkonnectFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSQuitTalkkonnectFilenameAndPath) Returned Error: ", err)
		}

	}
	ServerHop = true
	b.CleanUp()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlE() {
	log.Println("--")
	log.Println("info: Ctrl-E Pressed Send Email Requested")

	var i int = 0
	var tries int = 10

	for i = 0; i < tries; i++ {
		goodGPSRead, err := getGpsPosition(false)

		if err != nil {
			log.Println("warn: GPS Function Returned Error Message", err)
			break
		}

		if goodGPSRead == true {
			break
		}

	}

	if i == tries {
		log.Println("warn: Could Not Get a Good GPS Read")
		log.Println("--")
		return
	}

	if TTSEnabled && TTSSendEmail {
		err := PlayWavLocal(TTSSendEmailFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSSendEmailFilenameAndPath) Returned Error: ", err)
		}

	}

	if EmailEnabled {

		emailMessage := fmt.Sprintf(EmailMessage + "\n")
		emailMessage = emailMessage + fmt.Sprintf("Ident: %s \n", b.Ident)
		emailMessage = emailMessage + fmt.Sprintf("Mumble Username: %s \n", b.Username)

		if EmailGpsDateTime {
			emailMessage = emailMessage + fmt.Sprintf("Date "+GPSDate+" UTC Time "+GPSTime+"\n")
		}

		if EmailGpsLatLong {
			emailMessage = emailMessage + fmt.Sprintf("Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)+" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64)+"\n")
		}

		if EmailGoogleMapsURL {
			emailMessage = emailMessage + "http://www.google.com/maps/place/" + strconv.FormatFloat(GPSLatitude, 'f', 6, 64) + "," + strconv.FormatFloat(GPSLongitude, 'f', 6, 64)
		}

		err := sendviagmail(EmailUsername, EmailPassword, EmailReceiver, EmailSubject, emailMessage)
		if err != nil {
			log.Println("alert: Error from Email Module: ", err)
		}
	} else {
		log.Println("info: Sending Email Disabled in Config")
	}
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlF() {
	log.Println("--")
	log.Println("info: Ctrl-F Connect Previous Server Requested")

	if TTSEnabled && TTSPreviousServer {
		err := PlayWavLocal(TTSPreviousServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSPreviousServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {

		if AccountIndex > 0 {
			AccountIndex--
		} else {
			AccountIndex = AccountCount - 1
		}

		ConnectAttempts = 0
		log.Printf("info: Connecting to Account Index [%d] \n", AccountIndex)

		ServerHop = true
		KillHeartBeat = true
		b.Client.Disconnect()

		b.Name = Name[AccountIndex]
		b.Address = Server[AccountIndex]
		b.Username = Username[AccountIndex]
		b.Ident = Ident[AccountIndex]
		b.ChannelName = Channel[AccountIndex]
		b.PreInit1(true)

	} else {
		log.Println("info: talkkonnect will remain connected to the same server as only 1 Account is Set as Default")
	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlL() {
	reset()
	log.Println("info: Ctrl-L Pressed Cleared Screen")
	if TargetBoard == "rpi" {
		if LCDEnabled == true {
			LcdText = [4]string{"nil", "nil", "nil", "nil"}
			go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}

		if OLEDEnabled == true {
			Oled.DisplayOn()
			LCDIsDark = false
			oledDisplay(true, 0, 0, "") // clear the screen
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlO() {
	log.Println("info: Ctrl-O Ping Servers ")

	if TTSEnabled && TTSPingServers {
		err := PlayWavLocal(TTSPingServersFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: PlayWavLocal(TTSPingServersFilenameAndPath) Returned Error: ", err)
		}

	}

	b.pingServers()
}

func (b *Talkkonnect) commandKeyCtrlN() {
	log.Println("--")
	log.Println("info: Ctrl-N Connect Next Server Requested")

	if TTSEnabled && TTSNextServer {
		err := PlayWavLocal(TTSNextServerFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("alert: PlayWavLocal(TTSNextServerFilenameAndPath) Returned Error: ", err)
		}

	}

	if AccountCount > 1 {
		if AccountIndex < AccountCount-1 {
			AccountIndex++
		} else {
			AccountIndex = 0
		}

		ConnectAttempts = 0
		log.Printf("info: Connecting to Account Index [%d] \n", AccountIndex)

		ServerHop = true
		KillHeartBeat = true
		b.Client.Disconnect()

		b.Name = Name[AccountIndex]
		b.Address = Server[AccountIndex]
		b.Username = Username[AccountIndex]
		b.Ident = Ident[AccountIndex]
		b.ChannelName = Channel[AccountIndex]
		b.PreInit1(true)
	} else {
		log.Println("info: talkkonnect will remain connected to the same server as only 1 Account is Set as Default")
	}

	log.Println("--")
}

//New
func (b *Talkkonnect) commandKeyCtrlI() {
	log.Println("--")
	log.Println("info: Ctrl-I Traffic Recording Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "traffic" {
		log.Println("warn: Traffic Recording Not Enabled")
	}

	/* if TTSEnabled && TTSTrafficRecoding {
		        err := PlayWavLocal(TTSTrafficRecordingFilenameAndPath, TTSVolumeLevel)
		        if err != nil {
	log.Println("PlayWavLocal(TTSTrafficRecordingFilenameAndPath) Returned Error: ", err)
		        }
		} */ // Create TTS

	if AudioRecordEnabled == true {
		if AudioRecordMode == "traffic" {
			if AudioRecordFromOutput != "" {
				if AudioRecordSoft == "sox" {
					//log.Println("info: sox is Recording Traffic")
					go AudioRecordTraffic()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Traffic Audio Rec ->", "nil"} // 4 or 3
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Traffic Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("info: Traffic Recording is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlJ() {
	log.Println("--")
	log.Println("info: Ctrl-J Ambient (Mic) Recording Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "ambient" {
		log.Println("warn: Ambient (Mic) Recording Not Enabled")
	}

	/* if TTSEnabled && TTMicRecoding {
		        err := PlayWavLocal(TTSMicRecordingFilenameAndPath, TTSVolumeLevel)
		        if err != nil {
	  log.Println("PlayWavLocal(TTSMicRecordingFilenameAndPath) Returned Error: ", err)
		        }
		} */ // Create TTS

	if AudioRecordEnabled == true {
		if AudioRecordMode == "ambient" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					// log.Println("info: sox is Recording Traffic")
					go AudioRecordAmbient()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Mic Audio Rec ->", "nil"} // 4 or 3
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Mic Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("info: Ambient (Mic) Recording is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlK() {
	log.Println("--")
	log.Println("info: Ctrl-K Combo Recording (Traffic and Mic) Requested")
	if AudioRecordEnabled != true {
		log.Println("warn: Audio Recording Function Not Enabled")
	}
	if AudioRecordMode != "combo" {
		log.Println("warn: Combo Recording (Traffic and Mic) Not Enabled")
	}

	/* if TTSEnabled && TTSComboRecoding {
		        err := PlayWavLocal(TTSComboRecordingFilenameAndPath, TTSVolumeLevel)
		        if err != nil {
	log.Println("PlayWavLocal(TTSComboRecordingFilenameAndPath) Returned Error: ", err)
		        }
		} */ // Create TTS

	if AudioRecordEnabled == true {
		if AudioRecordMode == "combo" {
			if AudioRecordFromInput != "" {
				if AudioRecordSoft == "sox" {
					// log.Println("info: sox is Recording Traffica and Mic (Combo)")
					go AudioRecordCombo()
					if TargetBoard == "rpi" {
						if LCDEnabled == true {
							LcdText = [4]string{"nil", "nil", "Combo Audio Rec ->", "nil"} // 4 or 3
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled == true {
							oledDisplay(false, 5, 1, "Combo Audio Rec ->") // 6 or 5
						}
					}
				} else {
					log.Println("info: Combo Recording (Traffic and Mic) is not Enabled or sox Encountered Problems")
				}
			}
		}
	}
}

func (b *Talkkonnect) commandKeyCtrlP() {
	if !(IsConnected) {
		return
	}
	b.BackLightTimer()
	log.Println("--")
	log.Println("info: Ctrl-P pressed Panic Button Start/Stop Simulation Requested")

	if TTSEnabled && TTSPanicSimulation {
		err := PlayWavLocal(TTSPanicSimulationFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSPanicSimulationFilenameAndPath) Returned Error: ", err)
		}

	}

	if PEnabled {

		if b.IsTransmitting {
			log.Println("--")
			b.TransmitStop(false)
		} else {
			b.IsTransmitting = true
			b.SendMessage(PMessage, PRecursive)

		}

		if PSendIdent {
			b.SendMessage(fmt.Sprintf("My Username is %s and Ident is %s", b.Username, b.Ident), PRecursive)
		}

		if PSendGpsLocation && GpsEnabled {

			var i int = 0
			var tries int = 10

			for i = 0; i < tries; i++ {
				goodGPSRead, err := getGpsPosition(false)

				if err != nil {
					log.Println("warn: GPS Function Returned Error Message", err)
					break
				}

				if goodGPSRead == true {
					break
				}
			}

			if i == tries {
				log.Println("warn: Could Not Get a Good GPS Read")
			}

			if goodGPSRead == true && i != tries {
				log.Println("info: Sending GPS Info My Message")
				gpsMessage := "My GPS Coordinates are " + fmt.Sprintf(" Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)) + fmt.Sprintf(" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64))
				b.SendMessage(gpsMessage, PRecursive)
			}

			IsPlayStream = true
			b.playIntoStream(PFilenameAndPath, PVolume)
			if TargetBoard == "rpi" {
				if LCDEnabled == true {
					LcdText = [4]string{"nil", "nil", "nil", "Panic Message Sent!"}
					go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
				}
				if OLEDEnabled == true {
					oledDisplay(false, 6, 1, "Panic Message Sent!")
				}
			}
			if PTxLockEnabled && PTxlockTimeOutSecs > 0 {
				b.TxLockTimer()
			}

		} else {
			log.Println("info: Panic Function Disabled in Config")
		}
		IsPlayStream = false
		b.IsTransmitting = false
		b.LEDOff(b.TransmitLED)
		log.Println("--")
	}
}

func (b *Talkkonnect) commandKeyCtrlR() {
	log.Println("--")
	log.Println("info: Ctrl-R Repeat TX test")
	isrepeattx = !isrepeattx
	go b.repeatTx()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlS() {
	log.Println("--")
	log.Println("info: Ctrl-S Scan Channels ")

	if TTSEnabled && TTSScan {
		err := PlayWavLocal(TTSScanFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSScanFilenameAndPath) Returned Error: ", err)
		}

	}

	b.Scan()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlT() {
	log.Println("--")
	log.Println("info: Ctrl-T Thanks and Acknowledgements Screen Request ")
	talkkonnectAcknowledgements()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlV() {
	log.Println("--")
	log.Println("info: Ctrl-V Version Request ")
	log.Printf("info: Talkkonnect Version %v Released %v\n", talkkonnectVersion, talkkonnectReleased)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlX() {
	log.Println("--")
	log.Println("info: Ctrl-X Print XML Config " + ConfigXMLFile)

	if TTSEnabled && TTSPrintXmlConfig {
		err := PlayWavLocal(TTSPrintXmlConfigFilenameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("warn: PlayWavLocal(TTSPrintXmlConfigFilenameAndPath) Returned Error: ", err)
		}

	}

	printxmlconfig()
	log.Println("--")
}

func (b *Talkkonnect) SendMessage(textmessage string, PRecursive bool) {
	if !(IsConnected) {
		return
	}
	b.Client.Self.Channel.Send(textmessage, PRecursive)
}

func (b *Talkkonnect) SetComment(comment string) {
	if IsConnected {
		b.BackLightTimer()
		b.Client.Self.SetComment(comment)
		t := time.Now()
		if TargetBoard == "rpi" {
			if LCDEnabled == true {
				LcdText[2] = "Status at " + t.Format("15:04:05")
				time.Sleep(500 * time.Millisecond)
				LcdText[3] = b.Client.Self.Comment
				go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
			}
			if OLEDEnabled == true {
				oledDisplay(false, 1, 1, "Status at "+t.Format("15:04:05"))
				oledDisplay(false, 4, 1, b.Client.Self.Comment)
			}
		}
	}
}

func (b *Talkkonnect) BackLightTimer() {

	BackLightTime = *BackLightTimePtr

	if TargetBoard != "rpi" || LCDBackLightTimerEnabled == false {
		return
	}

	BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)

	//log.Printf("debug: LCD Backlight Timer Address %v", BackLightTime, " On\n")
	b.LEDOn(b.BackLightLED)

	go func() {
		<-BackLightTime.C
		//log.Printf("debug: LCD Backlight Timer Address %v", BackLightTime, " Off Timed Out After", LCDBackLightTimeoutSecs, " Seconds\n")
		LCDIsDark = true
		time.Sleep(100 * time.Millisecond)
		if LCDInterfaceType == "parallel" {
			b.LEDOff(b.BackLightLED)
		}
		if LCDInterfaceType == "i2c" {
			lcd := hd44780.NewI2C4bit(LCDI2CAddress)
			if err := lcd.Open(); err != nil {
				log.Println("alert: Can't open lcd: " + err.Error())
				return
			}
			lcd.ToggleBacklight()
		}
		if OLEDInterfacetype == "i2c" {
			Oled.DisplayOff()
			LCDIsDark = true
		}
	}()
}

func (b *Talkkonnect) TxLockTimer() {
	if PTxLockEnabled {
		TxLockTicker := time.NewTicker(time.Duration(PTxlockTimeOutSecs) * time.Second)
		log.Println("warn: TX Locked for ", PTxlockTimeOutSecs, " seconds")
		b.TransmitStop(false)
		b.TransmitStart()

		go func() {
			<-TxLockTicker.C
			b.TransmitStop(true)
			log.Println("warn: TX UnLocked After ", PTxlockTimeOutSecs, " seconds")
		}()
	}
}

func (b *Talkkonnect) pingServers() {

	currentconn := " Not Connected "

	for i := 0; i < len(Server); i++ {
		resp, err := gumble.Ping(Server[i], time.Second*1, time.Second*5)

		if b.Address == Server[i] {
			currentconn = " ** Connected ** "
		} else {
			currentconn = ""
		}

		log.Println("info: Server # ", i+1, "["+Name[i]+"]"+currentconn)

		if err != nil {
			log.Println(fmt.Sprintf("warn: Ping Error ", err))
			log.Println("--")
			continue
		}

		major, minor, patch := resp.Version.SemanticVersion()

		log.Println("info: Server Address:         ", resp.Address)
		log.Println("info: Server Ping:            ", resp.Ping)
		log.Println("info: Server Version:         ", major, ".", minor, ".", patch)
		log.Println("info: Server Users:           ", resp.ConnectedUsers, "/", resp.MaximumUsers)
		log.Println("info: Server Maximum Bitrate: ", resp.MaximumBitrate)
		log.Println("info: --")
	}
}

func (b *Talkkonnect) repeatTx() {
	for i := 0; i < 100; i++ {
		b.TransmitStart()
		b.IsTransmitting = true
		time.Sleep(1 * time.Second)
		b.IsTransmitting = false
		time.Sleep(1 * time.Second)
		if i > 0 {
			log.Println("info: TX Cycle ", i)
			if isrepeattx {
				log.Println("warn: Repeat Tx Loop Text Forcefully Stopped")
			}
		}

		if isrepeattx {
			break
		}
	}
}
