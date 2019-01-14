# talKKonnect

### A Headless Mumble Client/Gateway for Single Board Computers, PCs or Virtual Environments (Transceiver/Intercom)

---
### What is talKKonnect?

[talKKonnect](http://www.talkkonnect.com) is a headless self contained mumble Push to Talk (PTT) client complete with LCD, Channel and Volume control. 


This project is a fork of [talkiepi](http://projectable.me/) by Daniel Chote which was in turn a fork of [barnard](https://github.com/layeh/barnard) a text based mumble client. 
All clients were made using [golang](https://golang.org/) and based on [gumble](https://github.com/layeh/gumble) library by Tim Cooper.

[talKKonnect](http://www.talkkonnect.com) was developed initially to run on SBCs. The latest version can be scaled to run all the way from SBCs to full fledged servers.
Raspberry Pi 3, Orange Pis, PCs and virtual environments (Oracle VirtualBox, KVM and Proxmox) targets have all been tested and work as expected.

### Why Was talKKonnect created?

I [Suvir Kumar](https://www.linkedin.com/in/suvir-kumar-51a1333b) created talKKonnect for fun. I missed the younger days making homebrew CB radios and talking to all
those amazing people who taught me so much. Living in an apartment in the age of the internet drove me to create talKKonnect.


[talKKonnect](http://www.talkkonnect.com) was originally created to have the form factor of a desktop transceiver. With community feedback we started to push the envelope
to make it more versatile and scalable. 

#### Some of the interesting features are #### 
* Communications bridge to interface external (otherwise not compatable) radio systems both over the air and over IP networks.
* Interface to portable or base radios (Beefing portable radios or UART radio boards). 
* Connecting to low cost USB GPS dongles (for instance “u-blox”) for GPS tracking. 
* Mass scale customization with centralized Configuration using auto-provisioning of a XML config file.
* LCD Screen showing relevent real time information such as *server info, current channel, who is currently talking, etc.*
* Connecting to an [arduino](https://www.arduino.cc/en/Guide/ArduinoDue) daughter board via USB for I/O control when running in the datacenter as a radio gateway 
* local/ssh control via a USB keyboard/terminal and remote control is done over http api.
* panic button, when pressed, talKKonnect will send an alert message with GPS coordinates, followed by an email indication current location in google maps. 


Pictures and more information of my builds can be found on my blog here [www.talKKonnect.com](https://www.talKKonnect.com)

### Hardware Features ###

You can use an external microphone with push buttons (up/down) for Channel navigation for a mobile transceiver like experience. 
Currently talKKonnect works with 4×20 Hitachi [HD44780](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf) LCD screen in parallel mode.  Other screens like [OLED](https://learn.adafruit.com/monochrome-oled-breakouts) and [NEXTION](https://nextion.itead.cc/)
with I2C interfacing support are also in the pipeline.

Low cost audio amplifiers like [PAM8403](https://www.instructables.com/id/PAM8403-6W-STEREO-AMPLIFIER-TUTORIAL/) or similar “D” class amplifiers, are recommended for talKKonnect builds.


#### There are 4 LED indicators that can be build on the front panel to show the following statuses ####
* Connected to a server and is currently online
* There are other participants logged into the same channel
* Currently in transmitting mode 
* Currently receving an audio stream (someone is talking on the channel)



#### The tkio arduino daughter board (USB Interface) ####

The USB arduino due daughter I/O board enables talKKonnect to be used in the datacenter as a Gateway to physically interface at the hardware level between different radio 
networks. (Under Development).

* 4 x relays *(output)*
* 4 x leds *(output)*
* 1 x Buzzer *(output)*
* 1 x DTMF encoder *(output)*
* 4 x opto inputs *(input)*
* 2 x push buttons *(input)*
* 1 x DTMF decoder *(input)*


### Software Features ###

* *Colorized LOGs* are shown on the debugging terminal for events as they happen in real time. 
* Playing of configurable *alert sounds* as different events happen.
* Configurable *TTS prompts* to announce different events for those use special use cases where it is required. 
* *Roger Beep* playing can be enabled on release of the PTT button to indicate end of transmission. 
* *Muting* of The speaker when pressing PTT to prevent audio feedback and give a radio communication like experience. 
* LCD display can show *channel information, server information, who joined, who is speaking, etc.* 
* Configuration is kept in a single *highly granular XML file*, where options can be enabled or disabled.


### Installation Instructions For Raspberry Pi 3 ###


Download the latest version of [Raspbian Stretch Lite](https://www.raspberrypi.org/downloads/raspbian). 
At the time of making this document latest image release date was 2018-11-13 (Kernel Version 4.14). 
Download the ZIP file and extract IMG file to some temporary directory.

Use any USB / SD card imaging software for Windows of your other OS. Some of the many options are:
* [USB Image Tool](https://www.alexpage.de/usb-image-tool)
* [Win32 Disk Imager](https://sourceforge.net/projects/win32diskimager)
* [Rufus](https://rufus.ie) 
* [Etcher](http://www.etcher.io)
* [Linux dd tool](https://elinux.org/RPi_Easy_SD_Card_Setup)


After the imaging, insert the SD card into your Raspberry Pi 3, connect the screen, keyboard and power supply and boot into the OS. 

Log in as user “pi” with password “raspberry” (this is the default username and password for a fresh install of Raspbian)

##### Set the new root password with #####

` sudo passwd root `

Log out of the account pi and log into the root account with your newly set password 

Run raspi-config and expand the file system by chosing “Advanced Options”->”Expand File System”. Reboot.

Next go to “Interfacing Options” in raspi-config and “Enable SSH Server”.
##### Edit the file with your favourite editor. #####

` /etc/ssh/sshd_config`   

##### Change the line #####

` #PermitRootLogin  prohibit-password  to  PermitRootLogin Yes`

##### Restart ssh server with #####

` service ssh restart`

Now you should be able to log in remotely via ssh using the root account and continue the installation.

##### Add user “talkkonnect” #####

` adduser --disabled-password --disabled-login --gecos "" talKKonnect`

##### Add user “talkkonnect” to groups #####

` usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio talkkonnect`

##### Update Raspbian with the command #####

` apt-get update`

##### Install prerequisite programs ##### 
(Note: If building talKKonnect on other than Raspberry Pi board, install mplayer instead of omxplayer) 

` apt-get install golang libopenal-dev libopus-dev libasound2-dev git ffmpeg omxplayer screen `


Decide if you want to run talKKonnect as a local user or root? Up to you. 

##### To build as a local user (Note: you can also build talKKonnect as root, if you prefer). #####

` su talkkonnect `

##### Create code and bin directories #####
````
cd /home/talkkonnect
mkdir /home/talkkonnect/gocode
mkdir /home/talkkonnect/bin
````

##### Export GO paths #####
````
export GOPATH=/home/talkkonnect/gocode
export GOBIN=/home/talkkonnect/bin 
````

##### Get programs and prepare for building talKKonnect #####

````
cd $GOPATH 
go get github.com/talkkonnect/talkkonnect 
cd $GOPATH/src/github.com/talkkonnect/talkkonnect
````

##### Before building the binary, confirm all features which you want enabled, the GPIO pins used and talKKonnect program configuration by editing file: ##### 

` /home/talKKonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml`


##### Build talKKonnect and test connection to your Mumble server. #####

` go build -o /home/talKKonnect/bin/talkkonnect cmd/talkkonnect/main.go `

##### Start  talKKonnect binary #####

````
cd /home/talkkonnect/bin/talkkonnect
./talkkonnect 
````
##### Or create a start script ##### 

````
cd
sudo nano talkkonnect-run
````

##### with contents: #####

````
#!/bin/bash 
killall -vs 9 talkkonnect 
sleep 1 
reset 
sleep 2 
/home/talkkonnect/bin/talkkonnect 
````

##### Make the script executable ##### 

` chmod +x talkkonnect-run ` 


##### You can start talKKonnect automatically on Raspberry Pi start up with “screen” program help. Add this line to /etc/rc.local file. before “exit 0”: #####

` screen -dmS talkkonnect-radio /root/talkkonnect-run & `

##### Then connect to active screen session with command “screen -r”. Exit the screen session with “Ctrl-A-D”. #####

##### talKKonnect welcome screen #####

````
┌────────────────────────────────────────────────────────────────┐
│  _        _ _    _                               _             │
│ | |_ __ _| | | _| | _____  _ __  _ __   ___  ___| |_           │
│ | __/ _` | | |/ / |/ / _ \| '_ \| '_ \ / _ \/ __|  __|         │
│ | || (_| | |   <|   < (_) | | | | | | |  __/ (__| |_           │
│  \__\__,_|_|_|\_\_|\_\___/|_| |_|_| |_|\___|\_ _|\__|          │
├────────────────────────────────────────────────────────────────┤
│A Flexible Headless Mumble Transceiver/Gateway for RPi/PC/VM    │
├────────────────────────────────────────────────────────────────┤
│Created By : Suvir Kumar  <suvir@talKKonnect.com>               │
├────────────────────────────────────────────────────────────────┤
│Version 1.32 Released January 2 2019                            │
│Additional Modifications Released under MPL 2.0 License         │
├────────────────────────────────────────────────────────────────┤
│visit us at www.talKKonnect.com and github.com/talKKonnect      │
└────────────────────────────────────────────────────────────────┘
Press the <Del> key for Menu Options or <Ctrl-c> to Quit talKKonnect
````


### Audio configuration ###


##### USB Sound Cards #####

For your audio input and output to work with talKKonnect, you needs to configure your sound settings. Configire and test your Linux sound system before building talKKonnect. talKKonnect works well with ALSA. There is no need to run it with PulseAudio. Any USB Sound cards supported in Linux, can be used with talKKonnect. Raspberry Pi’s have audio output with BCM2835 chip, but unfortunately no audio input, by the design. This is why we need a USB sound card. Many other types of single board computers come with both audio output and input (Orange Pi). USB Sound cards with CM sound chips like CM108, CM109, CM119, CM6206 chips are affordable and very common.

When connected to a Raspberry Pi, USB sound card can be identified with “lsusb” command. Typical response is something like this:

Bus 001 Device 004: ID 0d8c:000c C-Media Electronics, Inc. Audio Adapter

Audio playback devices can be listed with ”aplay -l” command.

Optional: When external USB Sound card is used, Raspberry Pi BCM2835 internal sound can be blacklisted or preveneted to load. To disable BCM2835 sound:

` nano /boot/config.txt `

##### Add these 2 lines: #####

````
#Disable audio (loads snd_bcm2835) 
dtparam=audio=off 
````

##### Save file and reboot. #####

If the BCM2835 sound is kept enabled, the USB sound card will usually be shown as card 1. When BCM sound is disabled, USB sound will be promoted to card 0.

For talKKonnect to know what audio devices to use (BCM2835 or USB Sound), ALSA audio config file needs to be edited. Edit file /usr/share/alsa/alsa.conf, 

nano /usr/share/alsa/alsa.conf and change 

````
defaults.ctl.card 0
defaults.pcm.card 0
````

from default BCM2835 audio index (0) to the USB Sound index (1)

````
#defaults.ctl.card 0
#defaults.pcm.card 0
defaults.ctl.card 1
defaults.pcm.card 1
````

(This change is not necessary if BCM2835 was disabled. USB sound card will be assigned card index number “0” in that case)

USB sound device can also be set in local profile 

` nano ~/.asoundrc `

For simple USB card cards .asound configuration like this will work:
````
    pcm.!default {
        type asym
        capture.pcm "mic"
        playback.pcm"speaker"
    }
    pcm.mic {
        type plug
        slave {
            pcm"hw:1,0"
        }	
    }
    pcm.speaker {
        type plug
        slave {
            pcm"hw:1,0"
        }
    }
````

When creating .asoundrc. match the sound card index number to the exact number of the device in your system. Run ”aplay -l” or ”amixer” to check on this. You also need to match the names of capture and playback devices in this config file for your particular sound device.

Note: If the sound device was configured in global /usr/share/alsa/alsa.conf coniguration file, there is no need to create a local .asoundrc file.

Microphone or input device needs to be “captured” for talKKonnect to work.   Run alsamixer and find your input device (mic or line in), then select it and press a space key. Red “capture” sign should show under the device in alsamixer.

##### Test that audio output is working by running: #####

` speaker-test `

You should hear white noise.

##### Test that audio input is working by looping recording to audio player: #####

` arecord –f CD | aplay `

You should hear yourself speaking to the microphone. 

Adjust your preferable microphone sensitivity and output gain through “alsamixer” or “amixer”, which requires some trial and error.

For a speaker muting to work when pressing a PTT, you need to enter the exact name of your audio device output in talKKonnect.xml file. This name may be different for different audio devices (e.g. Speaker, Master, Headphone, etc). Check audio output name with “aplay”, “alsamixer” or “amixer” and use that exact device name in the configuration.xml .


#### talKKonnect can be controled from terminal screen with function keys. ####

```
┌────────────────────────────────────────────────────────────────┐
│                 _                                              │
│ _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _           │
│| '_ ` _ \ / _` | | '_ \  | '_ ` _ \ / _ \ '_ \| | | |          │
│| | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |          │
│|_| |_| |_|\__,_|_|_| |_| |_| |_| |_|\___|_| |_|\__,_|          │
├─────────────────────────────┬──────────────────────────────────┤
│ <Del> to Display this Menu  | Ctrl-C to Quit talKKonnect       │
├─────────────────────────────┼──────────────────────────────────┤
│ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)           │
│ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level       │
│ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)    │
│ <F7>  List Server Channels  │ <F8>  Start Transmitting         │
│ <F9>  Stop Transmitting     │ <F10> List Online Users          │
│ <F11> Playback/Stop Chimes  │ <F12> For GPS Position           │
│<Ctrl-P> Start/Stop Panic Sim│<Ctrx-X> Screen Dump XML Config   │
│<Ctrl-E> Send Email          │<Ctrl-N> Connect to Next Server   │
├─────────────────────────────┴──────────────────────────────────┤
│   visit us at www.talKKonnect.com and github.com/talKKonnect   │
└────────────────────────────────────────────────────────────────┘
````


You can also [download](https://talKKonnect.com/wp-content/uploads/2019/01/Readme-13-01-2019.pdf) a PDF version with pictures of this document.
 
Please visit our [blog](www.talkkonnect.com) for our blog or [github](github.com/talkkonnect) for the latest source code and our [facebook](https://www.facebook.com/talkkonnect) page for future updates and information.


## Contributing 
We invite interested individuals to provide feedback and improvements to the project.
Currently we do not have a WIKI so send feedback to <suvir@talkkonnect.com>

## License
[talKKonnect](http://www.talKKonnect.com) is open source and available under the MPL license. 

## Update
<suvir@talkkonnect.com>


Updated 15/January/2019



