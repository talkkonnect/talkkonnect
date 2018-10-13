package talkkonnect

import (
	"fmt"
	"github.com/kennygrant/sanitize"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleopenal"
	"github.com/talkkonnect/gumble/gumbleutil"
	"github.com/talkkonnect/volume-go"
	"net"
	"os"
	"os/exec"
	"strconv"
	//	"strings"
	"github.com/comail/colog"
	term "github.com/talkkonnect/termbox-go"
	"io"
	"log"
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
)

type channelsList struct {
	chanID   uint32
	chanName string
}

func reset() {
	term.Sync()
}

func (b *Talkkonnect) Init() {

	err1 := term.Init()
	if err1 != nil {
		panic(err1)
	}

	f, err := os.OpenFile("/var/log/talkkonnect.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	b.LEDOffAll()

	if b.Logging == "screen" {
		colog.Register()
		colog.SetOutput(os.Stdout)
	} else {
		wrt := io.MultiWriter(os.Stdout, f)
		colog.SetOutput(wrt)
	}

	talkkonnectBanner()
	//	b.talkkonnectMenu()
	b.Config.Attach(gumbleutil.AutoBitrate)
	b.Config.Attach(b)
	b.initGPIO()
	b.Connect()

keyPressListenerLoop:
	for {
		switch ev := term.PollEvent(); ev.Type {
		case term.EventKey:
			switch ev.Key {
			case term.KeyEsc:
				log.Println("ESC Key is Invalid")
				reset()
				break keyPressListenerLoop
			case term.KeyDelete:
				log.Println("Delete Key Pressed Menu and Session Information Requested")
				b.talkkonnectMenu()
			case term.KeyF1:
				log.Println("F1 pressed Channel Up (+) Requested")
				b.ChannelUp()
			case term.KeyF2:
				log.Println("F2 pressed Channel Down (-) Requested")
				b.ChannelDown()
			case term.KeyF3:
				origMuted, _ := volume.GetMuted()
				if origMuted {
					_ = volume.Unmute()
					log.Println("F3 pressed Mute/Unmute Speaker Requested Now UnMuted")
					LcdText = [4]string{"nil", "nil", "nil", "UnMuted"}
					go hd44780.LcdDisplay(LcdText)
				} else {
					_ = volume.Mute()
					log.Println("F3 pressed Mute/Unmute Speaker Requested Now Muted")
					LcdText = [4]string{"nil", "nil", "nil", "Muted"}
					go hd44780.LcdDisplay(LcdText)
				}
			case term.KeyF4:
				origVolume, _ = volume.GetVolume()
				log.Println("F4 pressed Volume Level Requested and is at", origVolume, "%")
				LcdText = [4]string{"nil", "nil", "nil", "Volume " + strconv.Itoa(origVolume)}
				go hd44780.LcdDisplay(LcdText)
			case term.KeyF5:
				origVolume, _ = volume.GetVolume()
				if origVolume < 100 {
					err2 := volume.IncreaseVolume(+1)
					if err2 != nil {
						log.Println("F5 Increase Volume Failed! ", err2)
					}

					log.Println("F5 pressed Volume UP (+) Now At ", origVolume, "%")
					LcdText = [4]string{"nil", "nil", "nil", "Volume + " + strconv.Itoa(origVolume)}
					go hd44780.LcdDisplay(LcdText)
				} else {
					log.Println("F5 Increase Volume Already at Maximum Possible Volume")
					LcdText = [4]string{"nil", "nil", "nil", "Max Vol"}
					go hd44780.LcdDisplay(LcdText)
				}
			case term.KeyF6:
				origVolume, _ = volume.GetVolume()
				if origVolume > 0 {
					origVolume--
					err2 := volume.IncreaseVolume(-1)
					if err2 != nil {
						log.Println("F6 Decrease Volume Failed! ", err2)
					}

					log.Println("F6 pressed Volume Down (-) Now At ", origVolume, "%")
					LcdText = [4]string{"nil", "nil", "nil", "Volume - " + strconv.Itoa(origVolume)}
					go hd44780.LcdDisplay(LcdText)
				} else {
					log.Println("F6 Increase Volume Already at Minimum Possible Volume")
					LcdText = [4]string{"nil", "nil", "nil", "Min Vol"}
					go hd44780.LcdDisplay(LcdText)
				}

			case term.KeyF7:
				log.Println("F7 pressed Channel List Requested")
				b.ListChannels(true)
				b.ParticipantLEDUpdate()
			case term.KeyF8:
				log.Println("F8 pressed TX Mode Requested (Start Transmitting)")
				b.TransmitStart()
			case term.KeyF9:
				log.Println("F9 pressed RX Mode Request (Stop Transmitting)")
				b.TransmitStop()
			case term.KeyF10:
				log.Println("F10 Terminate Program Requested")
				b.CleanUp()
			default:
				log.Println("Invalid Keypress ASCII", ev.Ch)
			}
		case term.EventError:
			panic(ev.Err)
		}
	}
}

func (b *Talkkonnect) CleanUp() {
	log.Println("warning: SIGHUP Termination of Program Requested...shutting down...bye\n")
	b.Client.Disconnect()
	b.LEDOffAll()
	LcdText = [4]string{"talkkonnect", "session", "ended", "bye!"}
	go hd44780.LcdDisplay(LcdText)
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
		log.Println("warning: Error ", err, " Connection to ", b.Address, " failed (%s), attempting again in 10 seconds...")
		b.ReConnect()
	} else {

		b.OpenStream()

	}
}

func (b *Talkkonnect) ReConnect() {
	if b.Client != nil {
		b.Client.Disconnect()
	}

	if b.ConnectAttempts < 100 {
		go func() {
			time.Sleep(10 * time.Second)
			b.Connect()
		}()
		return
	} else {
		log.Println("warning: Unable to connect, giving up")
		LcdText = [4]string{"Failed to Connect!", "nil", "nil", "nil"}
		go hd44780.LcdDisplay(LcdText)

		os.Exit(1)
	}
}

func (b *Talkkonnect) OpenStream() {

	if os.Getenv("ALSOFT_LOGLEVEL") == "" {
		os.Setenv("ALSOFT_LOGLEVEL", "0")
	}

	if stream, err := gumbleopenal.New(b.Client); err != nil {

		log.Println("warning: Stream open error ", err)
		LcdText = [4]string{"Stream Error!", "nil", "nil", "nil"}

		go hd44780.LcdDisplay(LcdText)

		os.Exit(1)
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
	if b.IsConnected == false {
		return
	}

	b.IsTransmitting = true

	err := volume.Mute()
	if err != nil {
		log.Println("warning: Unable to Mute ", err)
	} else {
		log.Println("info: Speaker Muted ")
	}

	b.LEDOn(b.TransmitLED)

	LcdText[0] = "Online/TX"
	go hd44780.LcdDisplay(LcdText)

	b.Stream.StartSource()
}

func (b *Talkkonnect) TransmitStop() {
	if b.IsConnected == false {
		return
	}

	go b.RogerBeep()

	LcdText[0] = b.Address
	go hd44780.LcdDisplay(LcdText)

	b.LEDOff(b.TransmitLED)

	b.IsTransmitting = false
	b.Stream.StopSource()

	err := volume.Unmute()
	if err != nil {
		log.Println("warn: Unable to Unmute ", err)
	} else {
		log.Println("info: Speaker UnMuted ")
	}

}

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	b.Client = e.Client

	b.ConnectAttempts = 0

	b.IsConnected = true
	b.LEDOn(b.OnlineLED)
	log.Println("info: Connected to ", b.Client.Conn.RemoteAddr(), " on attempt", b.ConnectAttempts)
	if e.WelcomeMessage != nil {
		log.Print("info: Welcome message: ", esc(*e.WelcomeMessage))
	}

	LcdText = [4]string{"nil", "nil", "nil", "nil"}
	go hd44780.LcdDisplay(LcdText)
	b.ParticipantLEDUpdate()

	if b.ChannelName != "" {
		b.ChangeChannel(b.ChannelName)
		prevChannelID = b.Client.Self.Channel.ID
	}

}

func (b *Talkkonnect) OnDisconnect(e *gumble.DisconnectEvent) {
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
		log.Println("warning: Connection to ", b.Address, " disconnected, attempting again in 10 seconds...")
	} else {
		log.Println("warning: Connection to ", b.Address, " disconnected ", reason, ", attempting again in 10 seconds...\n")
	}

	b.ReConnect()
}

func (b *Talkkonnect) ChangeChannel(ChannelName string) {
	channel := b.Client.Channels.Find(ChannelName)
	if channel != nil {

		b.Client.Self.Move(channel)
		LcdText[1] = "Joined " + ChannelName
		go hd44780.LcdDisplay(LcdText)
		log.Println("info: Joined Channel Name: ", channel.Name, " ID ", channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		log.Println("warning: Unable to Find Channel Name: ", ChannelName)
		prevChannelID = 0
	}
}

func (b *Talkkonnect) ParticipantLEDUpdate() {
	time.Sleep(100 * time.Millisecond)

	var participantCount = len(b.Client.Self.Channel.Users)

	if participantCount > 1 && participantCount != prevParticipantCount {
		EventDing()
		prevParticipantCount = participantCount
	}

	if participantCount > 1 {

		log.Println("info: Channel ", b.Client.Self.Channel.Name, " has (", participantCount, ") participants")

		LcdText[0] = b.Address
		LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
		go hd44780.LcdDisplay(LcdText)

		b.LEDOn(b.ParticipantsLED)

	} else {

		log.Println("info: Channel ", b.Client.Self.Channel.Name, " has no other participants")

		LcdText = [4]string{b.Address, "Alone in " + b.Client.Self.Channel.Name, "", "nil"}
		go hd44780.LcdDisplay(LcdText)

		b.LEDOff(b.ParticipantsLED)
	}
}

func (b *Talkkonnect) OnTextMessage(e *gumble.TextMessageEvent) {
	//fmt.Printf("Message from %s: %s\n", e.Sender.Name, strings.TrimSpace(cleanstring(e.Message)))
	//LcdText[3]=strings.TrimSpace(esc(e.Message))
	//hd44780.LcdDisplay(LcdText)

	EventDing()

}

func (b *Talkkonnect) OnUserChange(e *gumble.UserChangeEvent) {
	var info string

	switch e.Type {
	case gumble.UserChangeConnected:
		info = "conn"
	case gumble.UserChangeDisconnected:
		info = "disconn"
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
	}

	if info != "chg channel" {
		if info != "" {
			log.Println("info: User ", cleanstring(e.User.Name), " ", info, "Event type=", e.Type, " channel=", e.User.Channel.Name)
		} else {
			log.Println("info: User ", cleanstring(e.User.Name), " Event type=", e.Type, " channel=", e.User.Channel.Name)
		}
		LcdText[2] = cleanstring(e.User.Name) + " " + info //+strconv.Atoi(string(e.Type))
	}

	b.ParticipantLEDUpdate()
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

		go hd44780.LcdDisplay(LcdText)

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

	log.Println("warning: Permission denied  ", info)
}

func (b *Talkkonnect) OnChannelChange(e *gumble.ChannelChangeEvent) {

	log.Println("info: ChangeChannelEvent")
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
	fmt.Println("On Audio Stream Event")
}

func esc(str string) string {
	return sanitize.HTML(str)
}

func cleanstring(str string) string {
	return sanitize.Name(str)
}

func (b *Talkkonnect) ListChannels(verbose bool) {
	var records = len(b.Client.Channels)
	for _, ch := range b.Client.Channels {
		if verbose == true {
			log.Println("info: Server Channels ChID=", ch.ID, " ChName=", ch.Name)
		}
		if ch.ID > maxchannelid {
			maxchannelid = ch.ID
		}
	}
	if verbose == true {
		log.Println("info: Total Channel Records=", records)
		log.Println("info: Max Channel ID=", maxchannelid)
	}
}

func (b *Talkkonnect) ChannelUp() {
	prevButtonPress = "ChannelUp"

	b.ListChannels(false)

	// Set Upper Boundary
	if b.Client.Self.Channel.ID == maxchannelid {
		log.Println("info: Can't Increment Channel Maximum Channel Reached")
		LcdText[2] = "Max Chan Reached"
		go hd44780.LcdDisplay(LcdText)
		return
	}

	// Implement Seek Up Avoiding any null channels
	if prevChannelID < maxchannelid {

		prevChannelID++

		//for i := uint32(prevChannelID); int(i) < len(b.Client.Channels)+1; i++ {
		for i := prevChannelID; uint32(i) < maxchannelid+1; i++ {
			//log.Println("debug: prevChannelID = ", i)

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
	prevButtonPress = "ChannelDown"

	b.ListChannels(false)

	// Set Lower Boundary
	if int(b.Client.Self.Channel.ID) == 0 {
		log.Println("info: Can't Decrement Channel Root Channel Reached")
		LcdText[2] = "Min Chan Reached"
		channel := b.Client.Channels[0]
		b.Client.Self.Move(channel)
		go hd44780.LcdDisplay(LcdText)
		return
	}

	// Implement Seek Down Avoiding any null channels
	if int(prevChannelID) > 0 {

		prevChannelID--

		//for i := uint32(prevChannelID); int(i) < len(b.Client.Channels); i-- {
		for i := uint32(prevChannelID); uint32(i) < maxchannelid; i-- {
			//log.Println("debug: prevChannelID = ", i)
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				break
			}
		}
	}
	return

}
