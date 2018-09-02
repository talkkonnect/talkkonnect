package talkkonnect

import (
//	"fmt"
	"github.com/fatih/color"
	hd44780 "github.com/go-hd44780"
	"github.com/itchyny/volume-go"
	"github.com/kennygrant/sanitize"
	"github.com/suvirkumar/gumble/gumble"
	"github.com/suvirkumar/gumble/gumbleopenal"
	"github.com/suvirkumar/gumble/gumbleutil"
	"net"
	"os"
	"strconv"
//	"strings"
	"time"
)

var (
	LcdText                     = [4]string{"nil", "nil", "nil", "nil"}
	currentChannelID     uint32 = 0
	prevChannelID        uint32 = 0
	prevParticipantCount        = 0
	prevButtonPress             = "none"
)

type channelsList struct {
	chanID   uint32
	chanName string
}

func (b *Talkkonnect) Init() {

	b.LEDOffAll()
	talkkonnectBanner()
	b.Config.Attach(gumbleutil.AutoBitrate)
	b.Config.Attach(b)
	b.initGPIO()
	b.Connect()

}

func (b *Talkkonnect) CleanUp() {
	color.Red(time.Now().Format(time.Stamp) + " SIGHUP   : Termination of Program Requested...shutting down...bye\n")
	b.Client.Disconnect()
	b.LEDOffAll()
	LcdText = [4]string{"nil", "nil", "nil", "nil"}
	go hd44780.LcdDisplay(LcdText)
}

func (b *Talkkonnect) Connect() {
	var err error
	b.ConnectAttempts++

	_, err = gumble.DialWithDialer(new(net.Dialer), b.Address, b.Config, &b.TLSConfig)
	if err != nil {
		color.Red(time.Now().Format(time.Stamp)+" Error    : Connection to %s failed (%s), attempting again in 10 seconds...\n", b.Address, err)
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
		color.Red(time.Now().Format(time.Stamp) + " Error    : Unable to connect, giving up\n")
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
		color.Red(time.Now().Format(time.Stamp)+" Error    : Stream open error (%s)\n", err)
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
		color.Red(time.Now().Format(time.Stamp)+" Error: Unable to Mute %+v \n", err)
	} else {
		color.Yellow(time.Now().Format(time.Stamp) + " Control : Muted\n")
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
		color.Red(time.Now().Format(time.Stamp)+" Error: Unable to Unmute %+v \n", err)
	} else {
		color.Yellow(time.Now().Format(time.Stamp) + " Control : Unmuted\n")
	}

}

func (b *Talkkonnect) OnConnect(e *gumble.ConnectEvent) {
	b.Client = e.Client

	b.ConnectAttempts = 0

	b.IsConnected = true
	b.LEDOn(b.OnlineLED)
	color.Yellow(time.Now().Format(time.Stamp)+" Event   : Connected to %s (%d)\n", b.Client.Conn.RemoteAddr(), b.ConnectAttempts)
	if e.WelcomeMessage != nil {
		color.Green("Welcome message: %+v\n", esc(*e.WelcomeMessage))
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
		color.Red(time.Now().Format(time.Stamp)+" Error: Connection to %s disconnected, attempting again in 10 seconds...\n", b.Address)
	} else {
		color.Red(time.Now().Format(time.Stamp)+" Error: Connection to %s disconnected (%s), attempting again in 10 seconds...\n", b.Address, reason)
	}

	b.ReConnect()
}

func (b *Talkkonnect) ChangeChannel(ChannelName string) {
	channel := b.Client.Channels.Find(ChannelName)
	if channel != nil {

		b.Client.Self.Move(channel)
		LcdText[1] = "Joined " + ChannelName
		go hd44780.LcdDisplay(LcdText)
		color.Yellow(time.Now().Format(time.Stamp)+" Event   : Joined Channel Name: %v ID %v \n", channel.Name, channel.ID)
		prevChannelID = b.Client.Self.Channel.ID
	} else {
		color.Red(time.Now().Format(time.Stamp)+" Error: Unable to Find Channel Name: %v \n", ChannelName)
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
		color.Green(time.Now().Format(time.Stamp)+" Info    : Channel '%s' has %d participants\n", b.Client.Self.Channel.Name, participantCount)

		LcdText[0] = b.Address
		LcdText[1] = b.Client.Self.Channel.Name + " (" + strconv.Itoa(participantCount) + " Users)"
		go hd44780.LcdDisplay(LcdText)

		b.LEDOn(b.ParticipantsLED)

	} else {
		color.Green(time.Now().Format(time.Stamp)+" Info    : Channel '%s' has no other participants\n", b.Client.Self.Channel.Name)

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
		color.Yellow(time.Now().Format(time.Stamp)+" Event   : %s Changed Channel to %v\n", cleanstring(e.User.Name), e.User.Channel.Name)
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
			color.Yellow(time.Now().Format(time.Stamp)+" Event   : User %s info=%s Event type=%+v channel=%v\n", cleanstring(e.User.Name), info, e.Type, e.User.Channel.Name)
		} else {
			color.Yellow(time.Now().Format(time.Stamp)+" Event   : User %s Event type=%+v channel=%v\n", cleanstring(e.User.Name), e.Type, e.User.Channel.Name)
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
		if prevButtonPress == "ChannelUp" && int(b.Client.Self.Channel.ID) == len(b.Client.Channels) {
			color.Green(time.Now().Format(time.Stamp) + " Info    : Can't Increment Channel Maximum Channel Reached!\n")
		}

		// Set Lower Boundary
		if prevButtonPress == "ChannelDown" && int(currentChannelID) == 0 {
			color.Green(time.Now().Format(time.Stamp) + " Info    : Can't Decrement Channel Root Channel Reached\n")
		}

		// Implement Seek Up Until Permissions are Sufficient for User to Join Channel whilst avoiding all null channels
		if prevButtonPress == "ChannelUp" && int(b.Client.Self.Channel.ID)+1 < len(b.Client.Channels) {
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
	color.Red(time.Now().Format(time.Stamp)+" Error   : Permission denied: %s\n", info)
}

func (b *Talkkonnect) OnChannelChange(e *gumble.ChannelChangeEvent) {
	color.Yellow(time.Now().Format(time.Stamp) + " Event   : ChangeChannelEvent \n")
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

func esc(str string) string {
	return sanitize.HTML(str)
}

func cleanstring(str string) string {
	return sanitize.Name(str)
}

func (b *Talkkonnect) ListChannels() {
	var records = len(b.Client.Channels)
	for _, ch := range b.Client.Channels {
		color.Magenta(time.Now().Format(time.Stamp)+" Channel : Server Channels ChID=%d ChName=%#v\n", ch.ID, ch.Name)
	}
	color.White(time.Now().Format(time.Stamp)+" Info    : Channel Total Records = %d\n", records)
}

func (b *Talkkonnect) ChannelUp() {
	prevButtonPress = "ChannelUp"

	b.ListChannels()

	// Set Upper Boundary
	if int(b.Client.Self.Channel.ID) == len(b.Client.Channels) {
		color.Green(time.Now().Format(time.Stamp) + " Info    : Can't Increment Channel Maximum Channel Reached!\n")
		LcdText[2] = "Max Chan Reached"
		go hd44780.LcdDisplay(LcdText)
		return
	}

	// Implement Seek Up Avoiding any null channels
	if int(prevChannelID) < len(b.Client.Channels) {

		prevChannelID++

		for i := uint32(prevChannelID); int(i) < len(b.Client.Channels)+1; i++ {
			color.Green(time.Now().Format(time.Stamp)+" Debug   : prevChannelID = %v\n", i)
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

	b.ListChannels()

	// Set Lower Boundary
	if int(b.Client.Self.Channel.ID) == 0 {
		color.Green(time.Now().Format(time.Stamp) + " Info    : Can't Decrement Channel Root Channel Reached\n")
		LcdText[2] = "Min Chan Reached"
		channel := b.Client.Channels[0]
		b.Client.Self.Move(channel)
		go hd44780.LcdDisplay(LcdText)
		return
	}

	// Implement Seek Down Avoiding any null channels
	if int(prevChannelID) > 0 {

		prevChannelID--

		for i := uint32(prevChannelID); int(i) < len(b.Client.Channels); i-- {
			color.Green(time.Now().Format(time.Stamp)+" Debug   : prevChannelID = %v\n", i)
			channel := b.Client.Channels[i]
			if channel != nil {
				b.Client.Self.Move(channel)
				break
			}
		}
	}
	return

}
