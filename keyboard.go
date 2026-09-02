/*
 * Shared keyboard command handling for USB and TTY evdev input devices.
 */

package talkkonnect

import (
	"log"
	"strconv"
	"strings"

	evdev "github.com/gvalkov/golang-evdev"
)

func (b *Talkkonnect) evdevKeyboardListener(devicePath string, deviceLabel string, keyMap map[rune]KBStruct, numlockScanID rune) {
	device, err := evdev.Open(devicePath)
	if err != nil {
		log.Printf("error: Unable to open %s input device %s: %v It will now be disabled\n", deviceLabel, devicePath, err)
		return
	}

	log.Printf("info: %s listener started on %s\n", deviceLabel, devicePath)

	var keyPrevStateDown bool

	for {
		events, err := device.Read()
		if err != nil {
			log.Printf("error: Unable to read event from %s %s: %v\n", deviceLabel, devicePath, err)
			return
		}

		for _, ev := range events {
			if ev.Type != evdev.EV_KEY {
				continue
			}

			ke := evdev.NewKeyEvent(&ev)
			kb, mapped := keyMap[rune(ke.Scancode)]

			if ke.State == evdev.KeyDown {
				keyPrevStateDown = true
				if mapped && strings.EqualFold(kb.Command, "soundinterfacepttkey") {
					b.TransmitStart()
				}
			}

			if ke.State == evdev.KeyHold {
				keyPrevStateDown = false
				if mapped {
					b.executeKeyboardRepeatCommand(kb)
				} else if ke.Scancode != uint16(numlockScanID) {
					log.Printf("error: %s key not mapped scanid %v\n", deviceLabel, ke.Scancode)
				}
				continue
			}

			if ke.State == evdev.KeyUp {
				if mapped && strings.EqualFold(kb.Command, "pttkey") && b.IsTransmitting {
					b.TransmitStop(false)
				}
			}

			if keyPrevStateDown && ke.State == evdev.KeyUp {
				keyPrevStateDown = false
				if mapped {
					b.executeKeyboardCommand(kb)
				} else if ke.Scancode != uint16(numlockScanID) {
					log.Printf("error: %s key not mapped scanid %v\n", deviceLabel, ke.Scancode)
				}
			}
		}
	}
}

func (b *Talkkonnect) executeKeyboardRepeatCommand(kb KBStruct) {
	switch strings.ToLower(kb.Command) {
	case "channelup":
		playIOMedia("usbchannelup")
		b.cmdChannelUp()
	case "channeldown":
		playIOMedia("usbchanneldown")
		b.cmdChannelDown()
	case "volumeup":
		playIOMedia("usbvolup")
		b.cmdVolumeRXUp()
	case "volumedown":
		playIOMedia("usbvoldown")
		b.cmdVolumeRXDown()
	case "volumetxup":
		playIOMedia("usbvolup")
		b.cmdVolumeTXUp()
	case "volumetxdown":
		playIOMedia("usbvoldown")
		b.cmdVolumeTXDown()
	case "radiovolup":
		b.cmdInternetRadioVolUp()
	case "radiovoldown":
		b.cmdInternetRadioVolDown()
	case "pttkey":
		if !b.IsTransmitting {
			b.TransmitStart()
		}
	}
}

func (b *Talkkonnect) executeKeyboardCommand(kb KBStruct) {
	switch strings.ToLower(kb.Command) {
	case "channelup":
		playIOMedia("usbchannelup")
		b.cmdChannelUp()
	case "channeldown":
		playIOMedia("usbchanneldown")
		b.cmdChannelDown()
	case "serverup":
		playIOMedia("usbserverup")
		b.cmdConnNextServer()
	case "serverdown":
		playIOMedia("usbpreviousserver")
		b.cmdConnPreviousServer()
	case "mute":
		playIOMedia("usbmute")
		b.cmdMuteUnmute("mute")
	case "unmute":
		b.cmdMuteUnmute("unmute")
		playIOMedia("usbunmute")
	case "mute-toggle":
		playIOMedia("usbmutetoggle")
		b.cmdMuteUnmute("toggle")
		playIOMedia("usbmutetoggle")
	case "stream-toggle":
		playIOMedia("usbstreamtoggle")
		b.cmdPlayback()
	case "currentrxvolume":
		playIOMedia("usbcurrentrxvol")
		b.cmdCurrentRXVolume()
	case "volumerxup", "volumeup", "volup":
		playIOMedia("usbvolup")
		b.cmdVolumeRXUp()
	case "volumerxdown", "volumedown", "voldown":
		playIOMedia("usbvoldown")
		b.cmdVolumeRXDown()
	case "currenttxvolume":
		playIOMedia("usbcurrenttxvol")
		b.cmdCurrentTXVolume()
	case "volumetxup":
		playIOMedia("usbvolup")
		b.cmdVolumeTXUp()
	case "volumetxdown":
		playIOMedia("usbvoldown")
		b.cmdVolumeTXDown()
	case "setcomment":
		if kb.ParamName == "setcomment" {
			log.Println("info: Set Commment ", kb.ParamValue)
			playIOMedia("usbsetcomment")
			b.Client.Self.SetComment(kb.ParamValue)
		}
	case "transmitstart":
		playIOMedia("usbstarttx")
		b.cmdStartTransmitting()
	case "transmitstop":
		playIOMedia("usbstoptx")
		b.cmdStopTransmitting()
	case "record":
		playIOMedia("usbrecord")
		if Config.Global.Hardware.AudioRecordFunction.Enabled && Config.Global.Hardware.AudioRecordFunction.RecordMode == "traffic" {
			go StartOpusTrafficRecording(b)
		} else {
			log.Println("warn: Traffic Recording Not Enabled")
		}
	case "voicetargetset":
		voicetarget, err := strconv.Atoi(kb.ParamValue)
		if err != nil {
			log.Println("error: Target is Non-Numeric Value")
		} else {
			playIOMedia("usbvoicetarget")
			b.cmdSendVoiceTargets(uint32(voicetarget))
		}
	case "mqttpubpayloadset":
		if kb.ParamName == "payloadvalue" {
			playIOMedia("usbmqttpubpayloadset")
			MQTTPublish(kb.ParamValue)
		}
	case "changechannel":
		if kb.ParamName == "channelname" {
			playIOMedia("changechannel")
			b.ChangeChannel(kb.ParamValue)
		}
	case "memorychannel":
		playIOMedia("memorychannel")
		b.cmdMemoryChannel(kb.ParamValue)
	case "repeatertoneplay":
		playIOMedia("iorepeatertone")
		b.cmdPlayRepeaterTone()
	case "listentochannelon":
		playIOMedia("usbstartlisten")
		b.listeningToChannels("start")
	case "listentochanneloff":
		playIOMedia("usbstopliosten")
		b.listeningToChannels("stop")
	case "soundinterfacepttkey":
		b.TransmitStop(false)
	case "gpioinput":
		GPIOInputPinControl(kb.ParamName, kb.ParamValue)
	case "gpiooutput":
		GPIOOutputPinControl(kb.ParamName, kb.ParamValue)
	case "radiotoggle":
		b.cmdInternetRadioToggle()
	case "radionext":
		b.cmdInternetRadioNext()
	case "radioprev":
		b.cmdInternetRadioPrev()
	default:
		log.Println("error: Command Not Defined ", strings.ToLower(kb.Command))
	}
}
