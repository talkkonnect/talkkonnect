package talkkonnect

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/comail/colog"
	"github.com/hegedustibor/htgo-tts"
	"github.com/kennygrant/sanitize"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/gpio"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleopenal"
	"github.com/talkkonnect/gumble/gumbleutil"
	_ "github.com/talkkonnect/gumble/opus"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
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
)

var (
	LcdText                     = [4]string{"nil", "nil", "nil", "nil"}
	currentChannelID     uint32 = 0
	prevChannelID        uint32 = 0
	prevParticipantCount        = 0
	prevButtonPress             = "none"
	maxchannelid         uint32 = 0
	origVolume           int    = 0
	tempVolume           int    = 0
	ConfigXMLFile        string
	GPSTime              string
	GPSDate              string
	GPSLatitude          float64
	GPSLongitude         float64
	Streaming            bool
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
	Daemonize   bool

	IsConnected    bool
	IsTransmitting bool

	GPIOEnabled        bool
	OnlineLED          gpio.Pin
	ParticipantsLED    gpio.Pin
	TransmitLED        gpio.Pin
	HeartBeatLED       gpio.Pin
	BackLightLED       gpio.Pin
	TxButton           gpio.Pin
	TxButtonState      uint
	UpButton           gpio.Pin
	UpButtonState      uint
	DownButton         gpio.Pin
	DownButtonState    uint
	PanicButton         gpio.Pin
	PanicButtonState    uint
	CommentButton       gpio.Pin
	CommentButtonState  uint
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

func PreInit(file string) {

	// read xml config file
	ConfigXMLFile = file
	err := readxmlconfig(ConfigXMLFile)
	if err != nil {
		log.Println("XML Parser Module Returned Error: ", err)
		log.Fatal("Please Make Sure the XML Configuration File is In the Correct Path with the Correct Format, Exiting talkkonnect! ...... bye\n")
	}

	// for auto provisioning
	if APEnabled {
		log.Println("info: Contacting http Provisioning Server Pls Wait")
		err := AutoProvision()
		time.Sleep(5 * time.Second)
		if err != nil {
			log.Println("alert: Error from AutoProvisioning Module: ", err)
			log.Println("Please Fix Problem with Provisioning Configuration or use Static File By Disabling AutoProvisioning ")
			log.Fatal("Exiting talkkonnect! ...... bye\n")
		} else {
			log.Println("info: Got New Configuration Reloading XML Config")
			ConfigXMLFile = file
			readxmlconfig(ConfigXMLFile)
		}
	}

	// Initialize
	b := Talkkonnect{
		Config:      gumble.NewConfig(),
		Address:     Server,
		ChannelName: Channel,
		Logging:     Logging,
		Daemonize:   Daemonize,
	}

	// if no username specified, lets just autogen a random one
	if len(Username) == 0 {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			log.Println("alert: Cannot Generate Random Name Error: ", err)
			log.Fatal("Exiting talkkonnect! ...... bye\n")
		}

		buf[0] |= 2
		b.Config.Username = fmt.Sprintf("talkkonnect-%02x%02x%02x%02x%02x%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	} else {
		b.Config.Username = Username
	}

	b.Config.Password = Password

	if Insecure {
		b.TLSConfig.InsecureSkipVerify = true
	}
	if Certificate != "" {
		cert, err := tls.LoadX509KeyPair(Certificate, Certificate)
		if err != nil {
			log.Println("alert: Certificate Error: ", err)
			log.Fatal("Exiting talkkonnect! ...... bye\n")
		}
		b.TLSConfig.Certificates = append(b.TLSConfig.Certificates, cert)
	}

	b.Init()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitStatus := 0

	<-sigs
	b.CleanUp()

	os.Exit(exitStatus)
}

func (b *Talkkonnect) Init() {

	f, err := os.OpenFile(LogFileNameAndPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("alert: Problem opening talkkonnect.log file Error: ", err)
		log.Fatal("Exiting talkkonnect! ...... bye\n")
	}

	if APIEnabled {
		go func() {
			http.HandleFunc("/", b.httpHandler)
			if err := http.ListenAndServe(":"+APIListenPort, nil); err != nil {
				log.Println("alert: Problem With Starting HTTP API Server Error: ", err)
				log.Fatal("Please Fix Problem or Disable API in XML Config, Exiting talkkonnect! ...... bye\n")
			}
		}()
	}

	b.LEDOffAll()

	if b.Logging == "screen" {
		colog.Register()
		colog.SetOutput(os.Stdout)
	} else {
		wrt := io.MultiWriter(os.Stdout, f)
		colog.SetOutput(wrt)
	}

	err = term.Init()
	if err != nil {
		log.Println("alert: Cannot Initalize Terminal Error: ", err)
		log.Fatal("Exiting talkkonnect! ...... bye\n")
	}

	b.Config.Attach(gumbleutil.AutoBitrate)
	b.Config.Attach(b)
	b.initGPIO()

	talkkonnectBanner()

	if TTSEnabled && TTSTalkkonnectLoaded {
		err := PlayWavLocal(TTSTalkkonnectLoadedFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}
	}

	b.Connect()

	//section to handle timers

	// Heartbeat LED to Show Talkkonnect is Still Alive if enabled in xml config
	if HeartBeatEnabled {
		HeartBeat := time.NewTicker(time.Duration(PeriodmSecs) * time.Millisecond)

		go func() {
			for _ = range HeartBeat.C {
				timer1 := time.NewTimer(time.Duration(LEDOnmSecs) * time.Millisecond)
				timer2 := time.NewTimer(time.Duration(LEDOffmSecs) * time.Millisecond)
				<-timer1.C
				b.LEDOn(b.HeartBeatLED)
				<-timer2.C
				b.LEDOff(b.HeartBeatLED)
			}
		}()
	}

	// Beacon to Make Announcments Every Period of Defined Seconds in XML and Play Announcement File
	if BeaconEnabled {
		BeaconTicker := time.NewTicker(time.Duration(BeaconTimerSecs) * time.Second)

		go func() {
			for _ = range BeaconTicker.C {
				b.PlayIntoStream(BeaconFileNameAndPath, BVolume)
				log.Println("warn: Beacon Enabled and Timed Out Auto Played File ", BeaconFileNameAndPath, " Into Stream")
			}
		}()
	}

	// LCD Backlight Control
	b.BackLightTimer()

keyPressListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				log.Println("--")
				log.Println("ESC Key is Invalid")
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
			case term.KeyCtrlP:
				b.commandKeyCtrlP()
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
			log.Fatal("Exiting talkkonnect! ...... bye\n")
		}
	}
}

func (b *Talkkonnect) CleanUp() {
	// LCD Backlight Control
	b.BackLightTimer()
	t := time.Now()
	log.Println("warn: SIGHUP Termination of Program Requested...shutting down...bye")
	b.Client.Disconnect()
	b.LEDOffAll()
	LcdText = [4]string{"talkkonnect stopped", t.Format("02-01-2006 15:04:05"), "Please Visit", "www.talkkonnect.com"}
	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	c := exec.Command("reset")
	c.Stdout = os.Stdout
	c.Run()
	os.Exit(0)
}

func (b *Talkkonnect) Connect() {

	time.Sleep(2 * time.Second)

	var err error
	b.ConnectAttempts++

	_, err = gumble.DialWithDialer(new(net.Dialer), b.Address, b.Config, &b.TLSConfig)
	if err != nil {
		log.Println("warn: Connection Error ", err, " connecting to ", b.Address, " failed (%s), attempting again in 10 seconds...")
		b.ReConnect()
	} else {

		b.OpenStream()

	}
}

func (b *Talkkonnect) ReConnect() {
	if b.Client != nil {
		log.Println("warn: Attenpting Reconnection With Server")
		b.Client.Disconnect()
	}

	if b.ConnectAttempts < 100 {
		go func() {
			time.Sleep(10 * time.Second)
			b.Connect()
		}()
		return
	} else {
		log.Println("warn: Unable to connect, giving up")
		LcdText = [4]string{"Failed to Connect!", "nil", "nil", "nil"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

		log.Fatal("Exiting talkkonnect! ...... bye\n")
	}
}

func (b *Talkkonnect) OpenStream() {

	if os.Getenv("ALSOFT_LOGLEVEL") == "" {
		os.Setenv("ALSOFT_LOGLEVEL", "0")
	}

	if stream, err := gumbleopenal.New(b.Client, VoiceActivityLEDPin, BackLightPin, BackLightTime, LCDBackLightTimeoutSecs, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin); err != nil {

		log.Println("warn: Stream open error ", err)
		LcdText = [4]string{"Stream Error!", "nil", "nil", "nil"}

		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

		log.Fatal("Exiting talkkonnect! ...... bye\n")
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
	// LCD Backlight Control
	b.BackLightTimer()

	t := time.Now()
	if !(b.IsConnected) {
		return
	}

	b.IsTransmitting = true

	err := volume.Mute(OutputDevice)
	if err != nil {
		log.Println("warn: Unable to Mute ", err)
	} else {
		log.Println("info: Speaker Muted ")
	}

	b.LEDOn(b.TransmitLED)
	LcdText[0] = "Online/TX"
	LcdText[3] = "TX at " + t.Format("15:04:05")

	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

	b.Stream.StartSource()

}

func (b *Talkkonnect) TransmitStop(withBeep bool) {
	// LCD Backlight Control
	b.BackLightTimer()

	if RogerBeepSoundEnabled {
		if !(b.IsConnected) {
			return
		}

		if withBeep {
			err := b.RogerBeep(RogerBeepSoundFilenameAndPath, RogerBeepSoundVolume)
			if err != nil {
				log.Println("alert: Roger Beep Module Returned Error: ", err)
			}
		}

		LcdText[0] = b.Address
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

		b.LEDOff(b.TransmitLED)

		b.IsTransmitting = false

		b.Stream.StopSource()

		err := volume.Unmute(OutputDevice)
		if err != nil {
			log.Println("warn: Unable to Unmute ", err)
		} else {
			log.Println("info: Speaker UnMuted ")
		}
	}
}

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	// LCD Backlight Control
	b.BackLightTimer()

	b.Client = e.Client

	b.ConnectAttempts = 0

	b.IsConnected = true
	b.LEDOn(b.OnlineLED)
	log.Println("info: Connected to ", b.Client.Conn.RemoteAddr(), " on attempt", b.ConnectAttempts)
	if e.WelcomeMessage != nil {
		log.Print(fmt.Sprintf("info: Welcome message: %s\n", esc(*e.WelcomeMessage)))
	}

	LcdText = [4]string{"nil", "nil", "nil", "nil"}
	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	b.ParticipantLEDUpdate(true)

	if b.ChannelName != "" {
		b.ChangeChannel(b.ChannelName)
		prevChannelID = b.Client.Self.Channel.ID
	}

}

func (b *Talkkonnect) OnDisconnect(e *gumble.DisconnectEvent) {
	// LCD Backlight Control
	b.BackLightTimer()

	var reason string
	switch e.Type {
	case gumble.DisconnectError:
		reason = "connection error"

	}

	b.IsConnected = false

	b.LEDOff(b.OnlineLED)
	b.LEDOff(b.ParticipantsLED)
	b.LEDOff(b.TransmitLED)

	if reason == "" {
		log.Println("warn: Connection to ", b.Address, " disconnected, attempting again in 10 seconds...")
	} else {
		log.Println("warn: Connection to ", b.Address, " disconnected ", reason, ", attempting again in 10 seconds...\n")
	}

	b.ReConnect()
}

func (b *Talkkonnect) ChangeChannel(ChannelName string) {
	// LCD Backlight Control
	b.BackLightTimer()

	channel := b.Client.Channels.Find(ChannelName)
	if channel != nil {

		b.Client.Self.Move(channel)
		LcdText[1] = "Joined " + ChannelName
		LcdText[2] = Username

		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
		log.Println("info: Joined Channel Name: ", channel.Name, " ID ", channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		log.Println("warn: Unable to Find Channel Name: ", ChannelName)
		prevChannelID = 0
	}
}

func (b *Talkkonnect) ParticipantLEDUpdate(verbose bool) {

	// LCD Backlight Control
	b.BackLightTimer()

	time.Sleep(100 * time.Millisecond)

	var participantCount = len(b.Client.Self.Channel.Users)

	if participantCount > 1 && participantCount != prevParticipantCount {

		if TTSEnabled && TTSParticipants {
			speech := htgotts.Speech{Folder: "audio", Language: "en"}
			speech.Speak("There Are Currently " + strconv.Itoa(participantCount) + " Users in The Channel " + b.Client.Self.Channel.Name)
		}
		err := PlayWavLocal(EventSoundFilenameAndPath, 100)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

		prevParticipantCount = participantCount
	}

	if participantCount > 1 {

		if verbose {
			log.Println("info: Current Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")
			LcdText[0] = b.Address
			LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
			go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
		}

		b.LEDOn(b.ParticipantsLED)
		b.LEDOn(b.OnlineLED)

	} else {

		if verbose {
			if TTSEnabled && TTSParticipants {
				speech := htgotts.Speech{Folder: "audio", Language: "en"}
				speech.Speak("You are Currently Alone in The Channel " + b.Client.Self.Channel.Name)
			}
			log.Println("info: Channel ", b.Client.Self.Channel.Name, " has no other participants")
			LcdText = [4]string{b.Address, "Alone in " + b.Client.Self.Channel.Name, "", "nil"}
			go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
		}
		b.LEDOff(b.ParticipantsLED)
	}
}

func (b *Talkkonnect) OnTextMessage(e *gumble.TextMessageEvent) {
	// LCD Backlight Control
	b.BackLightTimer()

	log.Println(fmt.Sprintf("alert: Message from %s: %s\n", e.Sender.Name, strings.TrimSpace(cleanstring(e.Message))))
	LcdText[0] = "Message From"
	LcdText[1] = e.Sender.Name
	LcdText[2] = strings.TrimSpace(esc(e.Message))
	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	err := PlayWavLocal(EventSoundFilenameAndPath, 100)
	if err != nil {
		log.Println("Play Wav Local Module Returned Error: ", err)
	}

}

func (b *Talkkonnect) OnUserChange(e *gumble.UserChangeEvent) {
	// LCD Backlight Control
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
		}

		// Implement Seek Dwn Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelDown" && int(b.Client.Self.Channel.ID) > 0 {
			prevChannelID--
			b.ChannelDown()
		}

		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

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
	item := 0
	for _, usr := range b.Client.Users {
		if usr.Channel.ID == b.Client.Self.Channel.ID {
			item++
			log.Println(fmt.Sprintf("info: %d. User %#v is online. [%v]", item, usr.Name, usr.Comment))
		}
	}

}

func (b *Talkkonnect) ListChannels(verbose bool) {

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

	if TTSEnabled && TTSChannelUp {
		err := PlayWavLocal(TTSChannelUpFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelUp"

	b.ListChannels(false)

	// Set Upper Boundary
	if b.Client.Self.Channel.ID == maxchannelid {
		log.Println("info: Can't Increment Channel Maximum Channel Reached")
		LcdText[2] = "Max Chan Reached"
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
		return
	}

	// Implement Seek Up Avoiding any null channels
	if prevChannelID < maxchannelid {

		prevChannelID++

		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {

			channel := b.Client.Channels[i]

			if channel != nil {
				b.Client.Self.Move(channel)
				break
			}
		}
	}
	return
}

func (b *Talkkonnect) ChannelDown() {

	if TTSEnabled && TTSChannelDown {
		err := PlayWavLocal(TTSChannelDownFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	prevButtonPress = "ChannelDown"
	b.ListChannels(false)

	// Set Lower Boundary
	if int(b.Client.Self.Channel.ID) == 0 {
		log.Println("info: Can't Decrement Channel Root Channel Reached")
		LcdText[2] = "Min Chan Reached"
		channel := b.Client.Channels[0]
		b.Client.Self.Move(channel)
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
		return
	}

	// Implement Seek Down Avoiding any null channels
	if int(prevChannelID) > 0 {

		prevChannelID--

		for i := uint32(prevChannelID); uint32(i) < maxchannelid; i-- {
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				break
			}
		}
	}
	return

}

func (b *Talkkonnect) httpHandler(w http.ResponseWriter, r *http.Request) {
	commands, ok := r.URL.Query()["command"]
	if !ok || len(commands[0]) < 1 {
		log.Println("warn: Url Param 'command' is missing")
		return
	}

	command := commands[0]
	log.Println("info: http command " + string(command))
	// LCD Backlight Control
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
	case "commandKeyCtrlP":
		if APIPanicSimulation {
			b.commandKeyCtrlP()
			fmt.Fprintf(w, "API Request Panic Simulation Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Request Panic Simulation Denied\n")
		}
	case "commandKeyCtrlX":
		if APIPrintXmlConfig {
			b.commandKeyCtrlX()
			fmt.Fprintf(w, "API Print XML Config Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Print XML Congfig Denied\n")
		}
	case "commandKeyCtrlE":
		if APIEmailEnabled {
			b.commandKeyCtrlE()
			fmt.Fprintf(w, "API Send Email Processed Succesfully\n")
		} else {
			fmt.Fprintf(w, "API Send Email Congfig Denied\n")
		}
	}
}

func (b *Talkkonnect) commandKeyDel() {
	log.Println("--")
	log.Println("Delete Key Pressed Menu and Session Information Requested")

	if TTSEnabled && TTSDisplayMenu {
		err := PlayWavLocal(TTSDisplayMenuFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	b.talkkonnectMenu()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF1() {
	log.Println("--")
	log.Println("F1 pressed Channel Up (+) Requested")

	b.ChannelUp()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF2() {
	log.Println("--")
	log.Println("F2 pressed Channel Down (-) Requested")

	b.ChannelDown()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF3() {
	log.Println("--")
	log.Println("info: ", TTSMuteUnMuteSpeakerFileNameAndPath)

	origMuted, err := volume.GetMuted(OutputDevice)

	if err != nil {
		log.Println("warn: get muted failed: %+v", err)
	}

	if origMuted {
		err := volume.Unmute(OutputDevice)

		if err != nil {
			log.Println("warn: unmute failed: %+v", err)
		}

		log.Println("F3 pressed Mute/Unmute Speaker Requested Now UnMuted")
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFileNameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("Play Wav Local Module Returned Error: ", err)
			}

		}

		LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	} else {
		if TTSEnabled && TTSMuteUnMuteSpeaker {
			err := PlayWavLocal(TTSMuteUnMuteSpeakerFileNameAndPath, TTSVolumeLevel)
			if err != nil {
				log.Println("Play Wav Local Module Returned Error: ", err)
			}

		}
		err = volume.Mute(OutputDevice)
		if err != nil {
			log.Println("warn: Mute failed: %+v", err)
		}

		log.Println("F3 pressed Mute/Unmute Speaker Requested Now Muted")
		LcdText = [4]string{"nil", "nil", "nil", "Muted"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF4() {
	log.Println("--")
	origVolume, err := volume.GetVolume(OutputDevice)
	if err != nil {
		log.Println("warn: Unable to get current volume: %+v", err)
	}

	log.Println("F4 pressed Volume Level Requested and is at", origVolume, "%")

	if TTSEnabled && TTSCurrentVolumeLevel {
		err := PlayWavLocal(TTSCurrentVolumeLevelFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(origVolume)}
	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
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

		log.Println("F5 pressed Volume UP (+) Now At ", origVolume, "%")

		LcdText = [4]string{"nil", "nil", "nil", "Volume + " + strconv.Itoa(origVolume)}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	} else {
		log.Println("F5 Increase Volume Already at Maximum Possible Volume")
		LcdText = [4]string{"nil", "nil", "nil", "Max Vol"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	}

	if TTSEnabled && TTSDigitalVolumeUp {
		err := PlayWavLocal(TTSDigitalVolumeUpFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
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

		log.Println("F6 pressed Volume Down (-) Now At ", origVolume, "%")

		LcdText = [4]string{"nil", "nil", "nil", "Volume - " + strconv.Itoa(origVolume)}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	} else {
		log.Println("F6 Increase Volume Already at Minimum Possible Volume")
		LcdText = [4]string{"nil", "nil", "nil", "Min Vol"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	}

	if TTSEnabled && TTSDigitalVolumeDown {
		err := PlayWavLocal(TTSDigitalVolumeDownFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF7() {
	log.Println("--")
	log.Println("F7 pressed Channel List Requested")

	if TTSEnabled && TTSListServerChannels {
		err := PlayWavLocal(TTSListServerChannelsFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	b.ListChannels(true)
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF8() {
	log.Println("--")
	log.Println("F8 pressed TX Mode Requested (Start Transmitting)")

	if TTSEnabled && TTSStartTransmitting {
		err := PlayWavLocal(TTSStartTransmittingFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	b.TransmitStart()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF9() {
	log.Println("--")
	log.Println("F9 pressed RX Mode Request (Stop Transmitting)")

	b.TransmitStop(true)

	if TTSEnabled && TTSStopTransmitting {
		err := PlayWavLocal(TTSStopTransmittingFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	log.Println("--")
}

func (b *Talkkonnect) commandKeyF10() {
	log.Println("--")
	log.Println("F10 pressed Online User(s) in Current Channel Requested")

	if TTSEnabled && TTSListOnlineUsers {
		err := PlayWavLocal(TTSListOnlineUsersFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	log.Println(fmt.Sprintf("Channel %#v Has %d Online User(s)", b.Client.Self.Channel.Name, len(b.Client.Self.Channel.Users)))
	b.ListUsers()
	b.ParticipantLEDUpdate(true)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF11() {
	log.Println("--")
	log.Println("F11 pressed Start/Stop Chimes Stream into Current Channel Requested")

	if TTSEnabled && TTSPlayChimes {
		err := PlayWavLocal(TTSPlayChimesFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	if b.IsTransmitting {
		log.Println("--")
		b.TransmitStop(false)
	} else {
		b.IsTransmitting = true
		go b.PlayIntoStream(ChimesSoundFilenameAndPath, ChimesSoundVolume)
	}
	b.IsTransmitting = false
	log.Println("--")
}

func (b *Talkkonnect) commandKeyF12() {
	log.Println("--")
	log.Println("F12 pressed GPS details requested")

	if TTSEnabled && TTSRequestGpsPosition {
		err := PlayWavLocal(TTSRequestGpsPositionFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	err := getGpsPosition(true)
	if err != nil {
		log.Println("warn: GPS Function Returned Error Message", err)
	}
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlC() {
	log.Println("--")
	log.Println("Ctrl-C Terminate Program Requested")

	if TTSEnabled && TTSQuitTalkkonnect {
		err := PlayWavLocal(TTSQuitTalkkonnectFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	b.CleanUp()
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlP() {
	b.BackLightTimer()
	log.Println("--")
	log.Println("Ctrl-P pressed Panic Button Start/Stop Simulation Requested")

	if TTSEnabled && TTSPanicSimulation {
		err := PlayWavLocal(TTSPanicSimulationFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
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
			b.SendMessage(fmt.Sprintf("My Username is %s and Ident is %s", Username, Ident), PRecursive)
		}

		if PSendGpsLocation && GpsEnabled {
			err := getGpsPosition(false)
			if err != nil {
				log.Println("warn: GPS Module Returned Error: ", err)
			} else {
				gpsMessage := "My GPS Coordinates are " + fmt.Sprintf(" Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)) + fmt.Sprintf(" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64))
				b.SendMessage(gpsMessage, PRecursive)
			}
		}

		go b.PlayIntoStream(PFileNameAndPath, PVolume)

		LcdText = [4]string{"nil", "nil", "nil", "Panic Message Sent!"}
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

		if PTxLockEnabled && PTxlockTimeOutSecs > 0 {
			b.TxLockTimer()
		}

	} else {
		log.Println("info: Panic Function Disabled in Config")
	}
	b.IsTransmitting = false
	b.LEDOff(b.TransmitLED)
	log.Println("--")
}

func (b *Talkkonnect) commandKeyCtrlE() {
	log.Println("--")
	log.Println("Ctrl-E Pressed Send Email Requested")

	getGpsPosition(false)

	if TTSEnabled && TTSSendEmail {
		err := PlayWavLocal(TTSSendEmailFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	if EmailEnabled {

		emailMessage := fmt.Sprintf(EmailMessage + "\n")
		emailMessage = emailMessage + fmt.Sprintf("Ident: %s \n", Ident)
		emailMessage = emailMessage + fmt.Sprintf("Mumble Username: %s \n", Username)

		if EmailGpsDateTime {
			emailMessage = emailMessage + fmt.Sprintf("Date "+GPSDate+" UTC Time "+GPSTime+"\n")
		}

		if EmailGpsLatLong {
			emailMessage = emailMessage + fmt.Sprintf("Latitude "+strconv.FormatFloat(GPSLatitude, 'f', 6, 64)+" Longitude "+strconv.FormatFloat(GPSLongitude, 'f', 6, 64)+"\n")
		}

		if EmailGoogleMapsUrl {
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

func (b *Talkkonnect) commandKeyCtrlX() {
	log.Println("--")
	log.Println("Ctrl-X Print XML Config " + ConfigXMLFile)

	if TTSEnabled && TTSPrintXmlConfig {
		err := PlayWavLocal(TTSPrintXmlConfigFileNameAndPath, TTSVolumeLevel)
		if err != nil {
			log.Println("Play Wav Local Module Returned Error: ", err)
		}

	}

	printxmlconfig()
	log.Println("--")
}

func (b *Talkkonnect) SendMessage(textmessage string, PRecursive bool) {
	b.Client.Self.Channel.Send(textmessage, PRecursive)
}

func (b *Talkkonnect) SetComment(comment string) {
	if b.IsConnected {
		// LCD Backlight Control
		b.BackLightTimer()
		b.Client.Self.SetComment(comment)
		t := time.Now()
		LcdText[2] = "Status at " + t.Format("15:04:05")
		time.Sleep(500 * time.Millisecond)
		LcdText[3] = b.Client.Self.Comment
		go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)
	}
}

func (b *Talkkonnect) BackLightTimer() {

	//for the case that backlight timer is set as not enabled or false leave the backlight on all the time
	if LCDBackLightTimerEnabled == false {
		b.LEDOn(b.BackLightLED)
		return
	}

	// reset the same timer every time there is a call to this function to keep the backlight on until the next timeout
	BackLightTime = *BackLightTimePtr
	BackLightTime.Reset(time.Duration(LCDBackLightTimeoutSecs) * time.Second)

	//log.Printf("debug: LCD Backlight Timer Address %v", BackLightTime, " On\n")
	b.LEDOn(b.BackLightLED)

	go func() {
		<-BackLightTime.C
		//log.Printf("debug: LCD Backlight Timer Address %v", BackLightTime, " Off Timed Out After", LCDBackLightTimeoutSecs, " Seconds\n")
		b.LEDOff(b.BackLightLED)
	}()
}

func (b *Talkkonnect) TxLockTimer() {
	// Used for Keep Transmitting for period of defined time after panic simulation
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
