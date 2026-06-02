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
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/allan-simon/go-singleinstance"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/gosshd"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleutil"
	_ "github.com/talkkonnect/gumble/opus"
	"github.com/talkkonnect/volume-go"
)

var (
	tmessage   string
	isrepeattx bool = true
	DaemonMode bool = false
)

type Talkkonnect struct {
	Config          *gumble.Config
	Client          *gumble.Client
	VoiceTarget     *gumble.VoiceTarget
	Name            string
	Address         string
	Username        string
	Ident           string
	TLSConfig       tls.Config
	ConnectAttempts uint
	Stream          *Stream
	ChannelName     string
	IsTransmitting  bool
	IsPlayStream    bool
	GPIOEnabled     bool

	// Hierarchical lifecycle: MasterCtx (daemon) -> ConnCtx (Mumble session) -> per-user stream contexts in StreamTracker.
	MasterCtx    context.Context
	masterCancel context.CancelFunc
	ConnCtx      context.Context
	connCancel   context.CancelFunc
}

// appTalkkonnect is the active Init() instance for shutdown hooks from CleanUp (GPIO, duplicate instance, etc.).
var appTalkkonnect *Talkkonnect

func (b *Talkkonnect) initDaemonLifecycle() {
	b.MasterCtx, b.masterCancel = context.WithCancel(context.Background())
}

func (b *Talkkonnect) shutdownDaemonLifecycle() {
	if b.masterCancel != nil {
		b.masterCancel()
		b.masterCancel = nil
	}
	b.MasterCtx = nil
}

func (b *Talkkonnect) startConnectionContext() {
	if b.connCancel != nil {
		b.connCancel()
		b.connCancel = nil
	}
	parent := b.MasterCtx
	if parent == nil {
		parent = context.Background()
	}
	b.ConnCtx, b.connCancel = context.WithCancel(parent)
}

func (b *Talkkonnect) cancelConnectionContext() {
	if b.connCancel != nil {
		b.connCancel()
		b.connCancel = nil
	}
	b.ConnCtx = nil
}

type ChannelsListStruct struct {
	chanIndex  int
	chanID     int
	chanName   string
	chanParent *gumble.Channel
	chanUsers  gumble.Users
}

func Init(file string, ServerIndex string) int {
	defer func() {
		if r := recover(); r != nil {
			internetRadioShutdownKill()
			atomic.StoreInt32(&shutdownExitCode, 1)
			panic(r)
		}
	}()

	prefixLogRegister()
	prefixLogSetOutput(os.Stdout)

	ConfigXMLFile = file
	err := readxmlconfig(ConfigXMLFile, false)
	if err != nil {
		message := err.Error()
		FatalCleanUp(message)
		return 0
	}

	if Config.Global.Software.Settings.SingleInstance {
		lockFile, err := singleinstance.CreateLockFile("talkkonnect.lock")
		if err != nil {
			log.Println("error: Another Instance of talkkonnect is already running!!, Killing this Instance")
			time.Sleep(5 * time.Second)
			TTSEvent("quittalkkonnect")
			CleanUp(false)
		}
		defer lockFile.Close()
	}

	if Config.Global.Software.Settings.Logging == "screen" {
		prefixLogSetFlags(log.Ldate | log.Ltime)
	}

	if Config.Global.Software.Settings.Logging == "screenwithlineno" || Config.Global.Software.Settings.Logging == "screenandfilewithlineno" {
		prefixLogSetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	ApplyCologMinLevelFromConfig()

	if Config.Global.Software.AutoProvisioning.Enabled {
		log.Println("info: Contacting http Provisioning Server Pls Wait")
		err := autoProvision()
		time.Sleep(5 * time.Second)
		if err != nil {
			FatalCleanUp("Error from AutoProvisioning Module " + err.Error())
			return 0
		} else {
			log.Println("info: Loading XML Config")
			ConfigXMLFile = file
			readxmlconfig(ConfigXMLFile, false)
		}
	}

	if Config.Global.Software.Settings.NextServerIndex > 0 {
		AccountIndex = Config.Global.Software.Settings.NextServerIndex
	} else {
		AccountIndex, _ = strconv.Atoi(ServerIndex)
	}

	b := Talkkonnect{
		Config:      gumble.NewConfig(),
		Name:        Name[AccountIndex],
		Address:     Server[AccountIndex],
		Username:    Username[AccountIndex],
		Ident:       Ident[AccountIndex],
		ChannelName: Channel[AccountIndex],
	}
	b.initDaemonLifecycle()
	appTalkkonnect = &b
	b.startSignalShutdownBridge()

	b.setupCologOutputAndStartBottomCLI()

	if Config.Global.Software.RemoteControl.MQTT.Enabled {
		log.Printf("info: Attempting to Contact MQTT Server")
		log.Printf("info: MQTT Broker      : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker)
		log.Printf("info: Subscribed topic : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTSubTopic)
		go b.mqttsubscribe()
	} else {
		log.Printf("info: MQTT Server Subscription Disabled in Config")
	}

	MACName := ""
	if len(b.Username) == 0 {
		macaddress, err := getMacAddr()
		if err != nil {
			log.Println("error: Could Not Get Network Interface MAC Address")
		} else {
			for _, a := range macaddress {
				tmacname := a
				MACName = strings.Replace(tmacname, ":", "", -1)
			}
		}
		if len(MACName) == 0 {
			buf := make([]byte, 6)
			_, err := rand.Read(buf)
			if err != nil {
				FatalCleanUp("Cannot Generate Random Number Error " + err.Error())
				return 0
			}
			buf[0] |= 2
			b.Config.Username = fmt.Sprintf("talkkonnect-%02x%02x%02x%02x%02x%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
		} else {
			b.Config.Username = fmt.Sprintf("talkkonnect-%v", MACName)
		}
	} else {
		b.Config.Username = Username[AccountIndex]
	}

	log.Printf("info: Connecting to Server %v Identified As %v With Username %v\n", Config.Accounts.Account[AccountIndex].ServerAndPort, Config.Accounts.Account[AccountIndex].Name, b.Config.Username)
	b.Config.Password = Password[AccountIndex]

	if Insecure[AccountIndex] {
		b.TLSConfig.InsecureSkipVerify = true
	}
	if Certificate[AccountIndex] != "" {
		cert, err := tls.LoadX509KeyPair(Certificate[AccountIndex], Certificate[AccountIndex])
		if err != nil {
			FatalCleanUp("Certificate Error " + err.Error())
			return 0
		}
		b.TLSConfig.Certificates = append(b.TLSConfig.Certificates, cert)
	}

	if Config.Global.Software.RemoteControl.HTTP.Enabled && !HTTPServRunning {
		SafeGo(func() {
			http.HandleFunc("/", b.httpAPI)
			http.HandleFunc("/config", b.httpConfig)
			http.HandleFunc("/uistatus", b.httpUIStatus)
			if err := http.ListenAndServe(":"+Config.Global.Software.RemoteControl.HTTP.ListenPort, nil); err != nil {
				FatalCleanUp("Problem Starting HTTP API Server " + err.Error())
			}
		})
	}

	b.ClientStart()

	IsConnected = false

	performCleanup(false)
	return int(atomic.LoadInt32(&shutdownExitCode))
}

// setupCologOutputAndStartBottomCLI opens the log file, wires prefix-level logging (screen+file + optional bottom CLI), and starts the bottom CLI goroutine.
// Called from Init as soon as Talkkonnect exists so the prompt and scroll region are ready before MQTT, HTTP, and ClientStart.
func (b *Talkkonnect) setupCologOutputAndStartBottomCLI() {
	f, err := os.OpenFile(Config.Global.Software.Settings.LogFilenameAndPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	log.Println("info: Trying to Open File ", Config.Global.Software.Settings.LogFilenameAndPath)
	if err != nil {
		FatalCleanUp("Problem Opening talkkonnect.log file " + err.Error())
		return
	}

	logOut := io.Writer(os.Stdout)
	switch Config.Global.Software.Settings.Logging {
	case "screenandfile":
		log.Println("info: Logging is set to: ", Config.Global.Software.Settings.Logging)
		logOut = io.MultiWriter(os.Stdout, f)
		prefixLogSetFlags(log.Ldate | log.Ltime)
	case "screenandfilewithlineno":
		log.Println("info: Logging is set to: ", Config.Global.Software.Settings.Logging)
		logOut = io.MultiWriter(os.Stdout, f)
		prefixLogSetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}
	if bottomTerminalCLIShouldWrap() {
		logOut = newBottomCLILogWriter(logOut)
		SafeGo(b.runBottomTerminalCLI)
	} else if Config.Global.Software.RemoteSSHConsole.Enabled && DaemonMode {
		// Daemon: stdout is usually not a TTY, so the bottom CLI log wrapper is skipped.
		// Still mirror each log line to embedded SSH sessions (same DECSTBM + prompt as local).
		logOut = newBottomCLISSHMirrorLogWriter(logOut)
	}
	prefixLogSetOutput(logOut)
}

func (b *Talkkonnect) ClientStart() {
	mainLoopRunning.Store(true)

	// Daemon-mode SSH console must listen before Mumble connects; otherwise a down
	// server blocks forever on FatalCleanUp and nothing binds on the remote port.
	if Config.Global.Software.RemoteSSHConsole.Enabled && DaemonMode {
		go b.runRemoteSSHConsoleDaemon()
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {
		if Config.Global.Hardware.GPIOOffset > 0 {
			for item, pins := range Config.Global.Hardware.IO.Pins.Pin {
				if pins.Enabled {
					newPinNo := Config.Global.Hardware.GPIOOffset + pins.PinNo
					log.Printf("info: Offsetting GPIO PinNo=%v -> %v Name=%v\n", pins.PinNo, newPinNo, pins.Name)
					Config.Global.Hardware.IO.Pins.Pin[item].PinNo = newPinNo
				}
			}
		}
	}

	b.Config.Attach(gumbleutil.AutoBitrate)
	b.Config.Attach(b)

	log.Printf("info: [%d] Default Mumble Accounts Found in XML config\n", AccountCount)

	if Config.Global.Hardware.TargetBoard == "rpi" {
		log.Println("info: Target Board Set as RPI (gpio enabled) ")
		b.initGPIO()
		GPIOOutAll("led/relay", "off")
		// if Config.Global.Hardware.LedStripEnabled {
		// 	MyLedStrip, _ = NewLedStrip()
		// 	log.Printf("info: Led Strip %v %s\n", MyLedStrip.buf, MyLedStrip.display)
		// }
	} else {
		log.Println("debug: Target Board Set as PC (gpio disabled) ")
	}

	if (Config.Global.Hardware.TargetBoard == "rpi" && Config.Global.Hardware.LCD.BacklightTimerEnabled) && (OLEDEnabled || Config.Global.Hardware.LCD.Enabled) {

		log.Println("debug: Backlight Timer Enabled by Config")
		BackLightTime = *BackLightTimePtr
		BackLightTime = time.NewTicker(time.Duration(Config.Global.Hardware.LCD.BackLightTimeoutSecs) * time.Second)

		go func() {
			for {
				<-BackLightTime.C
				log.Printf("debug: LCD Backlight Ticker Timed Out After %d Seconds", LCDBackLightTimeout)
				LCDIsDark = true
				if LCDInterfaceType == "parallel" {
					GPIOOutPin("backlight", "off")
				}
				if LCDInterfaceType == "i2c" {
					lcd := hd44780.NewI2C4bit(LCDI2CAddress)
					if err := lcd.Open(); err != nil {
						log.Println("error: Can't open lcd: " + err.Error())
						return
					}
					lcd.ToggleBacklight()
				}
				if OLEDEnabled && OLEDInterfacetype == "i2c" {
					Oled.DisplayOff()
					LCDIsDark = true
				}
			}
		}()
	} else {
		log.Println("debug: Backlight Timer Disabled by Config")
	}

	talkkonnectBanner("\x1b[0;44m") // add blue background to banner reference https://www.lihaoyi.com/post/BuildyourownCommandLinewithANSIescapecodes.html#background-colors

	err := volume.Unmute(Config.Global.Software.Settings.OutputDevice)

	if err != nil {
		log.Println("error: Unable to Unmute ", err)
	} else {
		log.Println("debug: Speaker UnMuted Before Connect to Server")
	}

	log.Println("info: Playing startup sound (talkkonnectloaded TTS event; may take a few seconds)")
	//TTSEvent("talkkonnectloaded")

	IsConnected = false
	IsPlayStream = false
	NowStreaming = false
	KillHeartBeat = false

	parentCtx := b.MasterCtx
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	if err := b.connectMumbleWithBackoff(parentCtx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}
		FatalCleanUp("Exceed Connection Threshold Reached! Giving Up trying to reach " + b.Address + ": " + err.Error())
		return
	}

	if (Config.Global.Hardware.HeartBeat.Enabled) && (Config.Global.Hardware.TargetBoard == "rpi") {
		HeartBeat := time.NewTicker(time.Duration(Config.Global.Hardware.HeartBeat.Periodmsecs) * time.Millisecond)

		go func() {
			for range HeartBeat.C {
				timer1 := time.NewTimer(time.Duration(Config.Global.Hardware.HeartBeat.LEDOnmsecs) * time.Millisecond)
				timer2 := time.NewTimer(time.Duration(Config.Global.Hardware.HeartBeat.LEDOffmsecs) * time.Millisecond)
				<-timer1.C
				if Config.Global.Hardware.HeartBeat.Enabled {
					GPIOOutPin("heartbeat", "on")
				}
				<-timer2.C
				if Config.Global.Hardware.HeartBeat.Enabled {
					GPIOOutPin("heartbeat", "off")
				}
				if KillHeartBeat {
					HeartBeat.Stop()
				}

			}
		}()
	}

	if Config.Global.Software.Beacon.Enabled {
		b.beaconPlay()
	}

	b.BackLightTimer()

	if LCDEnabled {
		GPIOOutPin("backlight", "on")
		LCDIsDark = false
	}

	if OLEDEnabled {
		Oled.DisplayOn()
		LCDIsDark = false
	}

	if Config.Global.Hardware.AudioRecordFunction.Enabled {

		if Config.Global.Hardware.AudioRecordFunction.RecordOnStart {

			if Config.Global.Hardware.AudioRecordFunction.RecordMode != "" {

				if Config.Global.Hardware.AudioRecordFunction.RecordMode == "traffic" {
					log.Println("info: Incoming Traffic will be Recorded with sox")
					AudioRecordTraffic()
					if Config.Global.Hardware.TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "nil", "Traffic Recording ->"} // 4
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
							oledDisplay(false, 6, OLEDStartColumn, "Traffic Recording") // 6
						}
					}
				}
				if Config.Global.Hardware.AudioRecordFunction.RecordMode == "ambient" {
					log.Println("info: Ambient Audio from Mic will be Recorded with sox")
					AudioRecordAmbient()
					if Config.Global.Hardware.TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "nil", "Mic Recording ->"} // 4
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
							oledDisplay(false, 6, OLEDStartColumn, "Mic Recording") // 6
						}
					}
				}
				if Config.Global.Hardware.AudioRecordFunction.RecordMode == "combo" {
					log.Println("info: Both Incoming Traffic and Ambient Audio from Mic will be Recorded with sox")
					AudioRecordCombo()
					if Config.Global.Hardware.TargetBoard == "rpi" {
						if LCDEnabled {
							LcdText = [4]string{"nil", "nil", "nil", "Combo Recording ->"} // 4
							LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
							oledDisplay(false, 6, OLEDStartColumn, "Combo Recording") //6
						}
					}
				}

			}

		}
	}

	if Config.Global.Hardware.USBKeyboard.Enabled && len(Config.Global.Hardware.USBKeyboard.USBKeyboardPaths) > 0 {
		go b.USBKeyboard()
	}

	if Register[AccountIndex] && !b.Client.Self.IsRegistered() {
		b.Client.Self.Register()
		log.Println("alert: Client Is Now Registered")
	} else {
		log.Println("info: Client Is Already Registered")

	}

	go func() {
		var RXLEDStatus bool
		for {
			select {
			case v := <-Talking:
				if LastSpeaker != v.WhoTalking {
					LastSpeaker = v.WhoTalking
				}
				if !RXLEDStatus {
					selfChName := ""
					if b.Client != nil && b.Client.Self != nil && b.Client.Self.Channel != nil {
						selfChName = b.Client.Self.Channel.Name
					}
					if selfChName == v.OnChannel {
						log.Printf("info: Speaking -> %v\n", v.WhoTalking)
					} else {
						log.Printf("info: Listening-> %v \033[31m[%v]\033[0m\n", v.WhoTalking, v.OnChannel)
					}
					RXLEDStatus = true
					ReceivingVoice = true
					txlockout := &TXLockOut
					*txlockout = true
					GPIOOutPin("voiceactivity", "on")
					//					MyLedStripVoiceActivityLEDOn()
					go rxScreen(LastSpeaker)
				}
				AnalogRelayZonesOnTalking(v)
			case <-TalkedTicker.C:
				if RXLEDStatus {
					RXLEDStatus = false
					ReceivingVoice = false
					txlockout := &TXLockOut
					*txlockout = false
					GPIOOutPin("voiceactivity", "off")
					//MyLedStripVoiceActivityLEDOff()
					//TalkedTicker.Stop()
				}
				AnalogRelayZonesOnSilence()
			}
		}
	}()

	if Config.Global.Hardware.GPS.Enabled {
		if Config.Global.Hardware.GPS.GpsInfoVerbose {
			go consoleScreenLogging()
		}

		if Config.Global.Hardware.TargetBoard == "rpi" && Config.Global.Hardware.Traccar.DeviceScreenEnabled && (Config.Global.Hardware.LCD.Enabled || Config.Global.Hardware.OLED.Enabled) {
			go gpsDisplayShow()
		}

		if Config.Global.Hardware.Traccar.Enabled {
			if Config.Global.Hardware.Traccar.Track && Config.Global.Hardware.Traccar.Protocol.Name == "osmand" {
				go httpSendTraccar("osmand")
			}

			if Config.Global.Hardware.Traccar.Track && Config.Global.Hardware.Traccar.Protocol.Name == "opengts" {
				go httpSendTraccar("opengts")
			}

			if Config.Global.Hardware.Traccar.Track && Config.Global.Hardware.Traccar.Protocol.Name == "t55" {
				go tcpSendT55Traccar()
			}
		}
	}

	if Config.Global.Hardware.Radio.Enabled {
		if !(Config.Global.Hardware.Radio.Sa818.Enabled && Config.Global.Hardware.Radio.Sa818.Serial.Enabled) {
			log.Println("error: Radio Module Not Configured Properly")
		} else {
			createEnabledRadioChannels()
			go radioSetChannel(Config.Global.Hardware.Radio.ConnectChannelID)
		}
	}

	if Config.Global.Software.Settings.StreamOnStart {
		time.Sleep(Config.Global.Software.Settings.StreamOnStartAfter * time.Second)
		b.cmdPlayback()
	}

	if Config.Global.Software.Settings.TXOnStart {
		time.Sleep(Config.Global.Software.Settings.TXOnStartAfter * time.Second)
		b.cmdStartTransmitting()
	}

	if Config.Global.Hardware.IO.RotaryEncoder.Enabled {
		createEnabledRotaryEncoderFunctions()
		if len(RotaryFunctions) > 0 {
			RotaryFunction.Item = RotaryFunctions[0].Item
			RotaryFunction.Function = RotaryFunctions[0].Function
		} else {
			RotaryFunction.Item = 0
			RotaryFunction.Function = "undefined"
		}
		log.Printf("info: Current Rotary Item %v Function %v\n", RotaryFunction.Item, RotaryFunction.Function)
	}

	b.ListChannels(true)

	// Set VT index to Zero
	Config.Accounts.Account[AccountIndex].Voicetargets.ID[0].IsCurrent = true
	b.sevenSegment("mumblechannel", strconv.Itoa(int(b.Client.Self.Channel.ID)))

	//Channel Listening on Startup
	if Config.Global.Software.Settings.ListenToChannelsOnStart {
		b.listeningToChannels("start")
	}

	AnalogRelayZonesInit()

	if Config.Global.Software.RemoteSSHConsole.Enabled && !DaemonMode {
		go gosshd.SSHDaemon(Config.Global.Software.RemoteSSHConsole.Username, Config.Global.Software.RemoteSSHConsole.Password, Config.Global.Software.RemoteSSHConsole.IDRSAFile, Config.Global.Software.RemoteSSHConsole.Listen)
	}

	// Block until MasterCtx canceled (signal bridge, FatalCleanUp, or reconnect exhaustion).
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.MasterCtx.Done():
			log.Println("info: Lifecycle context canceled; exiting ClientStart control loop")
			return
		case <-ticker.C:
			// idle tick — same cadence as the former sleep(1s) control loop
		}
	}
}
