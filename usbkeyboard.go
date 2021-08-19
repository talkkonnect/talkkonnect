package talkkonnect

import (
	"log"
	"strings"

	evdev "github.com/gvalkov/golang-evdev"
)

func (b *Talkkonnect) USBKeyboard() {

	device, err := evdev.Open(USBKeyboardPath)
	if err != nil {
		log.Printf("error: Unable to open USB Keyboard input device: %s\nError: %v It will now Be Disabled\n", USBKeyboardPath, err)
		return
	}

	for {
		events, err := device.Read()
		if err != nil {
			log.Printf("error: Unable to Read Event From USB Keyboard error %v\n", err)
			return
		}

		for _, ev := range events {
			switch ev.Type {
			case evdev.EV_KEY:
				ke := evdev.NewKeyEvent(&ev)

				if ke.State != evdev.KeyDown {
					continue
				}

				if _, ok := USBKeyMap[rune(ke.Scancode)]; ok {
					switch strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command) {
					case "channelup":
						b.cmdChannelUp()
					case "channeldown":
						b.cmdChannelDown()
					case "serverup":
						b.cmdConnNextServer()
					case "serverdown":
						b.cmdConnPreviousServer()
					case "mute":
						b.cmdMuteUnmute("mute")
					case "unmute":
						b.cmdMuteUnmute("unmute")
					case "mute-toggle":
						b.cmdMuteUnmute("toggle")
					case "stream-toggle":
						b.cmdPlayback()
					case "volumeup":
						b.cmdVolumeUp()
					case "volumedown":
						b.cmdVolumeDown()
					case "setcomment":
						CommentMessageOff = USBKeyMap[rune(ke.Scancode)].ParamName
						CommentMessageOn = USBKeyMap[rune(ke.Scancode)].ParamName
					case "transmitstart":
						b.cmdStartTransmitting()
					case "transmitstop":
						b.cmdStopTransmitting()
					case "record":
						b.cmdAudioTrafficRecord()
						b.cmdAudioMicRecord()
					case "voicetargetset":
						b.cmdSendVoiceTargets(USBKeyMap[rune(ke.Scancode)].ParamValue)
					default:
						log.Println("Command Not Defined ", strings.ToLower(USBKeyMap[rune(ke.Scancode)].Command))
					}
				} else {
					if ke.Scancode != uint16(NumlockScanID) {
						log.Println("error: Key Not Mapped ASC ", ke.Scancode)
					}
				}
			}
		}
	}
}
