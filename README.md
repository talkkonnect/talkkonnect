# talKKonnect

### A Headless Mumble Client/Transceiver/Walkie Talkie/Intercom/Gateway for Single Board Computers, PCs or Virtual Environments (IP Radio/IP PTT <push-to-talk>)

---
### What is talKKonnect?

[talKKonnect](http://www.talkkonnect.com) is a headless self contained mumble Push to Talk (PTT) client complete with LCD, Channel and Volume control. 

This project is a fork of [talkiepi](http://projectable.me/) by Daniel Chote which was in turn a fork of [barnard](https://github.com/layeh/barnard) a text based mumble client. 
talKKonnect was developed using [golang](https://golang.org/) and based on [gumble](https://github.com/layeh/gumble) library by Tim Cooper.
Libraries are however heavily vendored (modified from original). You will need to get most of the vendored libraries from this repo.

[talKKonnect](http://www.talkkonnect.com) was developed initially to run on SBCs. The latest version can be scaled to run all the way from ARM SBCs to full fledged X86 servers.
Raspberry Pi 3, Orange Pis, PCs and virtual environments (Oracle VirtualBox, KVM and Proxmox) targets have all been tested and work as expected.

### Why Was talKKonnect created?

I [Suvir Kumar](https://www.linkedin.com/in/suvir-kumar-51a1333b) created talKKonnect for fun. I missed the younger days making homebrew CB radios and talking to all
those amazing people who taught me so much. Living in an apartment in the age of the internet with the itch to innovate drove me to create talKKonnect. I am in no way a professional programmer but have tried to make talKKonnect production quality. 

[talKKonnect](http://www.talkkonnect.com) was originally created to have the form factor and functionality of a desktop transceiver. With community feedback we started to push the envelope to make it more versatile and scalable. 

#### Some of the interesting features are #### 
* Communications bridge to interface external (otherwise not compatible) radio systems both over the air and over IP networks.
* Interface to portable or base radios (Beefing portable radios or UART radio boards). 
* Connecting to low cost USB GPS dongles (for instance “u-blox”) for GPS tracking. 
* Mass scale customization with centralized Configuration using auto-provisioning of a XML config file.
* LCD/OLED Screen showing relevant real time information such as *server info, current channel, who is currently talking, etc.*
* Local/ssh control via a USB keyboard/terminal and remote control is done over http api.
* Panic button, when pressed, talKKonnect will send an alert message with GPS coordinates, followed by an email indication current location in google maps. 
* Future Plans of Connecting to an [arduino](https://www.arduino.cc/en/Guide/ArduinoDue) daughter board via USB for I/O control when running in the datacenter as a radio gateway 
* MQTT support for remote control for commands, LED Control, Button Control, Relay Control
* Repeater Opening Function with the ability to specify the tone frequency and duration.
* Other features as per suggested or requested by the community

Pictures and more information of my builds can be found on my blog here [www.talkkonnect.com](https://www.talkkonnect.com)

### Hardware Features ###

You can use an external microphone with push buttons (up/down) for Channel navigation for a mobile transceiver like experience. 
Currently talKKonnect works with 4×20 Hitachi [HD44780](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf) LCD screen in parallel mode.  Other screens like 0.96" and 1.3" [OLED](https://learn.adafruit.com/adafruit-oled-displays-for-raspberry-pi)
with I2C interface is also currently supported. 

Low cost audio amplifiers like [PAM8403](https://www.instructables.com/id/PAM8403-6W-STEREO-AMPLIFIER-TUTORIAL/) or similar “D” class amplifiers, are recommended for talKKonnect builds.

A good shileded cable for microphone is recommended to keep the noise picked up to a minimum.

#### There are 4 LED indicators that can be build on the front panel to show the following statuses ####
* Connected to a server and is currently online
* There are other participants logged into the same channel
* Currently in transmitting mode 
* Currently receiving an audio stream (someone is talking on the channel)
* Heart Beat to indicate that talKKonnect is running


### Software Features ###

* *Colorized LOGs* are shown on the debugging terminal for events as they happen in real time. Logging with line number, logging to file or screen or both. 
* Playing of configurable *alert sounds* as different events happen.
* Configurable *TTS prompts* to announce different events for those use special use cases where it is required. 
* *Roger Beep* playing can be enabled on release of the PTT button to indicate end of transmission. 
* *Muting* of The speaker when pressing PTT to prevent audio feedback and give a radio communication like experience. Both simplex and duplex settable in XML config. 
* LCD/OLED display can show *channel information, server information, who joined, who is speaking, etc.* 
* Configuration is kept in a single *highly granular XML file*, where options can be enabled or disabled.

### Quick Download Link for SD Card Image for Use with Pi 3/4 and USB Sound Card (For the impatient) ###
* [talkkonnect sd card img file](https://drive.google.com/file/d/1hbMFtKvlEYX-akqf976aVjHP4TcYFXgL/view?usp=sharing)
* For this pre-made image you can log in as root using the password talkkonnect
* This image will not be current but will be convinient for you so that you don't have to install everything from scratch
* After you intall the image you can copy the tk-update.sh in the scripts folder to your /root home and run it to update to the lastest version
* This image has been configured to work with a external USB sound card out of the box and the on board sound card for RPI is disabled
* The XML file is configured to run in PC mode so no GPIO will initalized, to run using GPIO you can change the mode to rpi mode.    

### Installation Instructions For Raspberry Pi 3/4 ###


Download the latest version of [Raspberry Pi OS Lite](https://www.raspberrypi.org/software/operating-systems/#raspberry-pi-os-32-bit). 
At the time of making/updating this document latest image release date was 02/12/2020 (Kernel Version 5.4). 
Download the ZIP file and extract IMG file to some temporary directory.

Use any USB / SD card imaging software for Windows or your other OS. Some of the many options are:
* [Raspberry Pi Imager](https://www.raspberrypi.org/software/)
* [USB Image Tool](https://www.alexpage.de/usb-image-tool)
* [Win32 Disk Imager](https://sourceforge.net/projects/win32diskimager)
* [Rufus](https://rufus.ie) 
* [balenaEtcher](https://www.balena.io/etcher/)
* [Linux dd tool](https://elinux.org/RPi_Easy_SD_Card_Setup)


After the imaging, insert the SD card into your Raspberry Pi, connect the screen, keyboard and power supply and boot into the OS. 

Log in as user “pi” with password “raspberry” (this is the default username and password for a fresh install of Raspbian)

##### Set the new root password with #####

` sudo passwd root `

Log out of the account pi and log into the root account with your newly set password 

Run raspi-config and expand the file system by choosing “Advanced Options”->”Expand File System”. Reboot.

Next go to “Interfacing Options” in raspi-config and “Enable SSH Server”.
##### Edit the file with your favourite editor. #####

` /etc/ssh/sshd_config`   

##### Change the line #####

` #PermitRootLogin  prohibit-password  to  PermitRootLogin yes`

##### Restart ssh server with #####

` service ssh restart`

##### Alternative Way to Enable SSH #####
With windows you can browse to your SD card and place the blank file ssh in the root folder.

Now you should be able to log in remotely via ssh using the root account and continue the installation.

##### Add user “talkkonnect” #####

` adduser --disabled-password --disabled-login --gecos "" talkkonnect`

##### Add user “talkkonnect” to groups #####

` usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio talkkonnect`

##### Update Raspbian with the command #####

` apt update`

##### Install prerequisite programs ##### 
(Note: If building talkkonnect on other than Raspberry Pi board, install mplayer instead of omxplayer) 

` apt install libopenal-dev libopus-dev libasound2-dev git ffmpeg omxplayer screen `

##### Install prerequisite programs ##### 

To get the newer versions of golang used for this project I suggest installing a precompiled binary of golang. If you use apt-get to install golang at this moment you will get an older incompatible version of golang.

To install GO as required for this project on the raspberry pi. First with your browser look on the website https://golang.org/dl/ on your browser and choose the latest version for the 
arm archecture. At the time of this writing the version is go1.15.6.linux-armv6l.tar.gz.

Please Note that if you use apt-get to install golang instead of follow the recommended instructions in this blog you will get the following error when compiling 
BackLightTime.Reset undefined (type * time.Ticker has no field or method Reset) 

As root user Get the link and use wget to download the binary to your talkkonnect

` cd /usr/local `

` wget https://golang.org/dl/go1.15.6.linux-armv6l.tar.gz `

` tar -zxvf go1.15.6.linux-armv6l.tar.gz `

` nano ~/.bashrc `

` export PATH=$PATH:/usr/local/go/bin `

` export GOPATH=/home/talkkonnect/gocode `

` export GOBIN=/home/talkkonnect/bin `

` export GO111MODULE="auto" `

` alias tk='cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/' `

Then log out and log in as root again and check if go in installed properly

` go version `

You should see the version that you just installed if all is ok you can continue to the next step

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
go get -v github.com/talkkonnect/talkkonnect 
cd $GOPATH/src/github.com/talkkonnect/talkkonnect
````

##### Before building the binary, confirm all features which you want enabled, the GPIO pins used and talKKonnect program configuration by editing file: ##### 

` /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml`

##### Build talKKonnect and test connection to your Mumble server. #####

` go build -o /home/talkkonnect/bin/talkkonnect cmd/talkkonnect/main.go `

##### Start  talKKonnect binary #####

````
cd /home/talkkonnect/bin
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
 │Created By : Suvir Kumar  <suvir@talkkonnect.com>               │
 ├────────────────────────────────────────────────────────────────┤
 │Press the <Del> key for Menu or <Ctrl-c> to Quit talkkonnect    │
 │Additional Modifications Released under MPL 2.0 License         │
 │visit us at www.talkkonnect.com and github.com/talkkonnect      │
 └────────────────────────────────────────────────────────────────┘
 Talkkonnect Version 1.53.01 Released December 24 2020
````


### Audio configuration ###


##### USB Sound Cards #####

For your audio input and output to work with talKKonnect, you needs to configure your sound settings. Configure and test your Linux sound system before building talKKonnect. talKKonnect works well with ALSA. There is no need to run it with PulseAudio. Any USB Sound cards supported in Linux, can be used with talKKonnect. Raspberry Pi’s have audio output with BCM2835 chip, but unfortunately no audio input, by the design. This is why we need a USB sound card. Many other types of single board computers come with both audio output and input (Orange Pi). USB Sound cards with CM sound chips like CM108, CM109, CM119, CM6206 chips are affordable and very common.

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

Note: If the sound device was configured in global /usr/share/alsa/alsa.conf configuration file, there is no need to create a local .asoundrc file.

Microphone or input device needs to be “captured” for talKKonnect to work.   Run alsamixer and find your input device (mic or line in), then select it and press a space key. Red “capture” sign should show under the device in alsamixer.

##### Test that audio output is working by running: #####

` speaker-test `

You should hear white noise.

##### Test that audio input is working by looping recording to audio player: #####

` arecord –f CD | aplay `

You should hear yourself speaking to the microphone. 

Adjust your preferable microphone sensitivity and output gain through “alsamixer” or “amixer”, which requires some trial and error.

For a speaker muting to work when pressing a PTT, you need to enter the exact name of your audio device output in talKKonnect.xml file. This name may be different for different audio devices (e.g. Speaker, Master, Headphone, etc). Check audio output name with “aplay”, “alsamixer” or “amixer” and use that exact device name in the configuration.xml .


#### talKKonnect can be controlled from terminal screen with function keys. ####

```
┌──────────────────────────────────────────────────────────────┐
│ _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _         │
│| '_ ` _ \ / _` | | '_ \  | '_ ` _ \ / _ \ '_ \| | | |        │
│| | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |        │
│|_| |_| |_|\__,_|_|_| |_| |_| |_| |_|\___|_| |_|\__,_|        │
├─────────────────────────────┬────────────────────────────────┤
│ <Del> to Display this Menu  | Ctrl-C to Quit talkkonnect     │
├─────────────────────────────┼────────────────────────────────┤
│ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)         │
│ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level     │
│ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)  │
│ <F7>  List Server Channels  │ <F8>  Start Transmitting       │
│ <F9>  Stop Transmitting     │ <F10> List Online Users        │
│ <F11> Playback/Stop Chimes  │ <F12> For GPS Position         │
├─────────────────────────────┼────────────────────────────────┤
│<Ctrl-D> Debug Stacktrace    │                                │
│<Ctrl-E> Send Email          │<Ctrl-N> Conn Next Server       │
│<Ctrl-F> Conn Previous Server│<Ctrl-P> Panic Simulation       │
│<Ctrl-Q> Reserved            │<Ctrl-S> Scan Channels          │
│<Ctrl-V> Display Version     │<Ctrl-T> Thanks/Acknowledgements│
├─────────────────────────────┼────────────────────────────────┤
│<Ctrl-L> Clear Screen        │<Ctrl-O> Ping Servers           │
│<Ctrl-R> Repeat TX Loop Test │<Ctrl-X> Dump XML Config        │
├─────────────────────────────┼────────────────────────────────┤
│<Ctrl-I> Traffic Record      │<Ctrl-J> Mic Record             │
│<Ctrl-K> Traffic & Mic Record│<Ctrl-U> Show Uptime            │
├─────────────────────────────┼────────────────────────────────┤
│  visit us at www.talkkonnect.com and github.com/talkkonnect  │
└──────────────────────────────────────────────────────────────┘
````


### Explanation of talkkonnect.xml configuration files sections and tags 
[youtube-video](https://www.youtube.com/watch?v=-Dy96FXw0gA&ab_channel=SuvirKumar) is a video made for explaining the xml tags.

#### The Accounts Section
* The account section can have multiple accounts, talkkonnect will look for the first account with the xml tag default = "true" and attempt to connect to that server 
* When talkkonnected is connected to a server you can cycle through accounts in which enabled = "true" by pressing CTRL-N, talkkonnect will connect to the next enabled server in the list
* Talkkonnect will not attempt to connect to a server that has the account tag set default = "false" 
* The tag account name is just used to identify the server for logging purposes 
* The serverandport tag is for the server FQDN or IP address followed by  ":" (colon) and the port of mumble is running on for that particlar server.
* The username tag is used for identifying yourself on the mumble server and for authentication 
* The password tag is used if the mumble server requires password authentication 
* The insecure tag should be set as true if the server you are connecting to does not require a certificate 
* The certificate tag should contain the full path to your previously generated certificate which is usually a file with the extension of pem  
* The channel tag should only be populated want to connect to a specific channel other than the root channel on startup

### The Global Section of talkkonnect.xml (Software & Hardware)

#### Software Section

##### Settings Section
* The outputdevice tag should be set as the default audio output device that represents your audio output device when you run alsamixer. Examples are Speaker or Headphone etc. (Please note that the device name should be set exactly as shown in alsamixer. 
* The logfilenameandpath tag should contain the full path to a writable file that is created prior to running talkkonenct for logging purposes  
* Should you not require logging to screen set the logging tag to screen. Any other value will result logs to be shown on the screen and in the log file (note that if logging is not set to screen the logs will no longer be colorized)
* The daemonize tag is not currently supported. To run at startup and in the background you can configure in /etc/rc.local talkkonnect to run in a screen session.
* Cancellable Stream is used so that if you are streaming some audio via talKKonnect another user in the channel can stop your streaming by pressing PTT.
* Simplexwithmute is used to set simplex mode (mute speaker when transmitting) or full duplex mode (not mute speaker with transmitting)
* Nextserver index should be set to 0 as default, this is used to inform talKKonnect which server to connect to the next tim talKKonnect runs

##### Autoprovisioning Section
* Autoprovisioning is provided so that you can remotely provision a talkkonnect machine via http protocol from a web server 
* The autoprovisioning tag when set to true or false turns on and off the autoprovisioning function respectively
* The tkid tag is used to set the autoprovisioning filename (xxxx.xml) that talkkonnect will request from the autoprovisioing web server 
* The URL tag is used to define the url of the autoprovisioning webserver that hosts the configuration XML file 
* The savefileandpath tag are used to define the name and where the http fetched xml file will be stored locally. This is usually /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml

##### Beacon Section
* The beacon function was created to emulate a radio repeater beacon that will play certain wav files at defined periods to notify all users on a particular channel that the repeater is online nad functioning 
* The beacontimersecs is the interval time in seconds between the repleated messages 
* The beaconfileandpath is the tag which defines the file and path to a wav file that to be played at regular intervals 
* The volume tag can be set from 0.1 to 1 in intervals of 0.1 for setting up the volume the file playback into stream will be played

##### The TTS Section
* This section was created for users that want an audible response to events that happen (Users without LCD Screen) 
* You can disable the whole section TTS functionality by the tag tts enabled = false 
* You can choose to enable only certain events you are interested in by setting tag tts enabled = true and selecting the tag you want for your particular use case

##### The SMTP Section
* Talkkonnect currently can only connect to gmail's SMTP for sending emails 
* Define your gmail username and password along with the receiver of the email message in their respective tags 
* Define the subject and fixed message body of the email in their respective tags 
* Should you want to send the GPS timestamp in the email set the gpsdatetime tag to true (You have to have a USB GPS Dongle Connected and Configured for this to work) 
* Should you want to send by email your current GPS position in LAT and LONG coordinates you can enable this tag 
* If you want to include the url with your pinned location on google maps enable the googlemapurl tag

##### The Sounds Section
* Each sound item can be enabled/disabled and the corresponding playback volume can be also be set individually
* The filenameandpath tag should contain the the full path and filename of the WAV file you wish to play for each event 
* The event tag is used to play an audible alert when there are changes of other users statuses 
* The alert tag is used to play an WAV file into the stream to the receiving party upon a user generated panic request
* The rogerbeep tag is used to define the WAV file to play at the end of every transmission 
* The tag name chimes can be misleading. This function is very powerful and can be used to define a file or network stream that will be played into the mumble channel upon pressing the F11 key. Very useful for debugging.

##### The TXTIMEOUT section
* The txtimeout tag is used to limit the length of a single transmission in seconds. This tag is useful when used as a repeater between RF and mumble.

##### The API Section
* API section enables the user to granually control which remote control functions are available over http within the network 
* The tag apilisten port defines the port that talkkonnect should listen and respond to remote control http requests 
* To use httpapi you can use your browser to go to the url http://{talkkonnectip}/?command=F1 (Replace {talkkonnectip} with the IP address of your talkkonnect)
* HTTPAPI commands supported are F1  Channel Up (+), F2  Channel Down (-), F3  Mute/Unmute Speaker, F4  Current Volume Level, F5  Digital Volume Up (+), F6  Digital Volume Down (-), 
F7  List Server Channels, F8  Start Transmitting, F9  Stop Transmitting, F10 List Online Users, F11 Playback/Stop Chimes, F12 For GPS Position, Ctrl-E Send Email, Ctrl-L Clear Screen, 
Ctrl-M Ping Servers, Ctrl-N Connect Next Server, Ctrl-P Panic Simulation, Ctrl-S Scan Channels, Ctrl-X Dump XML Config

##### The PrintVariables Section
* This function is useful for debugging the values read from each section of the config xml file. You can control which section is shown. This command is tied to the CTRL-X key

#### Hardware Section
* The tag targetboard has 2 option (1) pc and (2)rpi. pc mode is used when talkkonnect is running on a pc or server that does not have GPIOs and is not interfaced to buttons and a LCD screen. 
* To run on raspberry pi or other compatible single board computers set the targetboard to rpi this will enable the GPIO outputs/inputs.

##### The Lights Section (OUTPUT)
* This section is used to define how the raspberry pi hardware (GPIO) is connected to the LED indicators 
* The voiceactivitypin tag defines the GPIO pin that will go to Logic HIGH and light up with there is someone transmitting on the mumble channel 
* The participantsledpin tag defines the GPIO pin that will go to Logic HIGH and light up when there are other users logged into the same mumble channel as you 
* The transmitledpin tag defines the GPIO pin that will go to Logic HIGH when you are transmitting on talkkonnect 
* The onlineledpin tag defines the GPIO pin that will go to Logic HIGH when you are authenticated and connected to a mumble server

##### The Heartbeat Section (OUTPUT)
* The heartbeat tag defines the GPIO pin that will toggle as per the defined values to show that talkkonnect is alive and operational 
* Note that this heartbeat can uses the same GPIO PIN and voiceactivitypin so that one LED can have dual function
* Note Disable heartbeat or do not use the same pin as voiceactivity LED if you connect talKKonnect to a transceiver

##### The Buttons Section (INPUT)
* This section defines the raspberry GPIO pins that are connected to push buttons that are pulled to ground by keypress and float upon release
* The txbuttonpin tag is connected to the PTT push button 
* The txtogglepin tag is connected to the PTT toggle button (Press and Release to Change State from RX to TX and vice versa) 
* The upbuttonpin tag is connected to the channel up button
* The downbuttonpin tag is connected to the channel down button 
* The panic button tag is connected to a button that will set the talkkonnect into panic mode (request for help)

##### The Comment Section
* This function allows the user to set 2 possible messages like for example away messages depending on the state of a toggle switch 
* When another party using talkkonenct presses F10 they can see the username along with the defined message (depending on the position of the switch on/off) in square brackets 
* The commentbuttonpin tag defines the GPIO pin that the toggle switch is connected to

##### The LCD Section (For HD44780 20x4 LCD SCREEN)
* At this moment talkkonnect supports the easily available 4 lines 20 characters HD44780 LCD Module. 
* To disable this screen option you can set enabled = "false"
* Parallel and i2c interfacing to the HD44780 LCD Module are both supported and can be configured in this section 
* Valid interfacetype tag are either parallel or i2c 
* The i2c address can be obtained from running the i2cdetect -y 1 command. Convert the address displayed in HEX to Decimal and fill into the lcdi2caddress tag 
* The backlight function and time is also available to turn off the LCD's backlight in case of inactivity on the channel for the defined timeout period in seconds 
* The rs, e, d4, d5, d6, d7 pins are the GPIO pins that connect to the HD44780 display in parallel mode 
* NOTE! You cannot use the pins 2,3 on raspberry pi for anything else other than I2C mode if you want to connect an I2C display

##### The OLED Section (For 0.96 and 1.3 Inch I2C Interface OLED SCREEN)
* At this moment talkkonnect also supports the easily available 0.96 and 1.3 Inch I2C OLED Screen. 
* To disable this screen option you can set enabled = "false"
* i2c interfacing is the only option that should be specified now spi has not been developed
* The i2c address can be obtained from running the i2cdetect -y 1 command. Convert the address displayed in HEX to Decimal and fill into the lcdi2caddress tag and mostly the i2c bus is 1. 
* There is no backlight function for oled screens yet 
* Your will have to specify the rows and columns your screen supports (for my screen i used 8 rows and 21 columns)
* The OLED display is display width and height for my screen was 130 by 64
* Another important settings is the oledstartcolumn setting for 0.96 screens set to 0 and for 1.3 inch screens set to 1. This will clear any garbage you see on the edge of the screen.
* NOTE! You cannot use the pins 2,3 on raspberry pi for anything else other than I2C mode if you want to connect an I2C display

##### The GPS Section
* Talkkonnect supports a ublox 6 USB module to provide GPS tracking on Panic mode activation  
* Set the enabled tag to false if you do not have a USB dongle connected 
* Define the port which the GPS is detected as in linux usually /dev/ttyACM0 
* Define all other serial port settings such as serial baud, even/odd/none parity, also stop and databits.

##### The PanicFunction Section
* The panic function can be enabled or disabled and is used to request for help 
* Filenameandpath tag is used to define the WAV file that will be played into a stream if the panic button is pressed 
* The volume tag defines the playback volume of the wav file into the stream 
* The sendident will send the contents of the ident tag defined in the account section. This is used in case you want for example your Name or alternate ID sent in the panic message. 
* The panicmessage tag defines the text message that will be sent to the parent channel and all child channels if recursivemessage is set as true when the panic button is pressed 
* The sendgpslocation tag enables the sending of the gps coordinates of the talkkonnect requesting help as a text message 
* The txlock enabled tag will lock up talkkonnect in transmit mode for the defined txlocktimeoutsecs after the button is pressed so the requester can talk without having to press ptt button


## Contributing 
We invite interested individuals to provide feedback and improvements to the project. Currently we do not have a WIKI so send feedback to <suvir@talkkonnect.com> or open and Issue in github
you can also check my blog  [www.talkkonnect.com](https://www.talkkonnect.com) for updates on the project

Please visit our [blog](www.talkkonnect.com) for our blog or [github](github.com/talkkonnect) for the latest source code and our [facebook](https://www.facebook.com/talkkonnect) page for future updates and information. 

## License 
[talKKonnect](http://www.talkkonnect.com) is open source and available under the MPL V2.00 license.

<suvir@talkkonnect.com> Updated 29/12/2020 talkkonnect version 1.53.01 is the latest release as of this writing.

