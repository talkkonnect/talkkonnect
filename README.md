# talKKonnect

### A Headless Mumble Client/Transceiver/Walkie Talkie/Intercom/Gateway for Single Board Computers, PCs or Virtual Environments (IP Radio/IP PTT <push-to-talk>)

---
### If you like and use talkkonnect please let us know, please star talkkonnect/talkkonnect repo on github.com!
	
### talKKonnect Version 2 Release Notice!
* Version 2 breaks backward compatability with version 1 configs!! So make sure you use the correct version of the config file to make it work
* The sample configs can be found in the sample-configs directory (they are marked as version 2 and version 1 respectively)
* These Image Was Created On 09/January/2022 and Runs Bullseye Release along with talkkonnect version 2.09.19 Released 08/January/2022

#### New Features in Version 2
* GPIO expander you can now have up to 16 x 8 = 128 GPIO using The MCP23017 Chip over I2c
* Seven Segment Support For Showing Channel like CB Rado using MAX7219 Chip with Seven Segment Displays 
* Rotary Encoder Support for Channel Up/Down, Volume Up/Down, SA818 Radio Module Frequency Change, Voice Target Change
* Shout and Whisper Support
* Channel Token Support
* Configurable Voice Target and Shortcut Support over USB Keypad
* Full Duplex Support (No More Audio Stuttering on Multiple people talking over each other)
* Sound Files can be tied with events/actions in config (Support both blocking and non-blocking modes)
* Announciator Support using Google TTS with Multi Language Support)
* API/MQTT Granular Configurable Support 
* GPIO Granular Configurable Support 
* Multimedia Playing Support both local and remote by API call

### So What then is talKKonnect, and why should I be Interested?

[talKKonnect](http://www.talkkonnect.com) is a headless self contained mumble Push to Talk (PTT) client complete with LCD, Channel and Volume control.

The Potential Uses of talKKonnect
* Mobile Radio Transceiver Desktop unit for communication between workgroups stationary or mobile without distance limiting the quality of communications
* Device to Bridge Between the World of PTT Communications over IP/Internet with the world of RF/Radio
* Open Source Replacement for Camera Crew Communication for Live Production Events
* Dispatch Communicatins between Dispatcher and Mobile Responders
* Ad-Hoc Group Communications where talkkonnect is used as the base station and Android Phones/IPhones or other rugged Android Devices used in the field.
* Use for IP Based Public Announcements (Recorded and Live) with targeting to specific devices or groups
* A Text to Speech Alert Announcement by API/MQTT to either play locally or to remote clients
* IP Intercom/Door Intercom or Intercom between remote places (Germans Love this for some unknown reason)
* A Toy for our big adults like amateur radio enthusiasts (like me and many of you). 
* A toy for your kids (so that they can feel how it was like to be a kid in the 80s with a CB radio Now With 7-Segment and Rotary Encoder Channel Changing)
* A customized version of your particular PTT Communication unique usecase as this project is an open souce platform whereby people can build on quickly


This project is a fork of [talkiepi](http://projectable.me/) by Daniel Chote which was in turn a fork of [barnard](https://github.com/layeh/barnard) a text based mumble client. 
talKKonnect was developed using [golang](https://golang.org/) and based on [gumble](https://github.com/layeh/gumble) library by Tim Cooper.
Most Libraries are however heavily vendored (modified from original). You will need to get the vendored libraries from this repo.

[talKKonnect](http://www.talkkonnect.com) was developed initially to run on SBCs. The latest version can be scaled to run all the way from ARM SBCs to full fledged X86 servers.
To compile on X86 archectures you would need to revert back to Tim Cooper's version of GOOPUS (Opus).
Raspberry Pi 2B,3B,3A+,3B+,4B Orange Pis, PCs and virtual environments (Oracle VirtualBox, KVM and Proxmox) targets have all been tested and work as expected. 
Rasperry Pi Zero W and Pi Zero WH (Version 1) will work with a "watered down" version of talkkonnect that uses a lower sampling rate so as not to use up all of the little CPU power provided by the Zero. However the newly released Raspberry Pi Zero Version 2 W is a perfect candidate for talkkonnect, both small and compact.

### Why Was talKKonnect created?

I [Suvir Kumar](https://www.linkedin.com/in/suvir-kumar-51a1333b) created talKKonnect for fun. I missed the younger days making homebrew CB, HAM radios and talking to all
those amazing people who taught me so much. 

Living in an apartment in the age of the internet with the itch to innovate drove me to create talKKonnect. I did it to learn programming so in no way am I a professional programmer, however talkkonnect is very stable and production ready.

I have tried to make the talKKonnect source code readable and stable to the best of my ability. Time permitting I will continue to work and learn from all those people who give feedback 
and show interest in using talkkonnect. 

[talKKonnect](http://www.talkkonnect.com) was originally created to have the form factor and functionality of a desktop transceiver. With community feedback we started to push the envelope to make it more versatile and scalable. 

#### Some of the interesting features are #### 
* XML Granular configurability for many uses cases.
* Multiple Server Configurations with channel control, channel scanning and server hopping
* Streaming Audio into the channel from locally stored media file or from internet stream
* Autoprovisioning for configuring multiple talkkonnects from a centralized http provisioning server 
* The User has a configurable choice of GPIO pins to use for each function on different boards 
* Communications bridge to interface external (otherwise not compatible) radio systems both over the air and over IP networks.
* Interface to portable or base station radios (Beefing portable radios or UART radio boards). 
* Connecting to low cost USB GPS dongles (for instance “u-blox”) for GPS tracking, Panic Alerts integration with traccar GPS tracking software. 
* LCD/OLED Screen (Parallel and I2c Interface) showing relevant real time information such as *server info, current channel, who is currently talking, etc.*
* Local or Remote Control via a USB keyboard/terminal or SSH terminal,  remote control can also be achieved over http api and/or MQTT.
* Panic button, when pressed, talKKonnect will send an alert message with GPS coordinates, followed by an email indication current location in google maps. 
* API/MQTT support for remote control for commands, LED Control, Button Control, Relay Control
* Tone Based Repeater Opening Function with the ability to specify the tone frequency and duration in configuration.
* Configurable Voice targeting via USB Numpad keyboard, TTT Keyboard, API, MQTT (Shouting and Whispering)
* Many Other features as per suggested or requested by the community

Pictures and more information of my builds can be found on my blog here [www.talkkonnect.com](https://www.talkkonnect.com)

### Hardware Features ###

You can use an external microphone with push buttons (up/down) for Channel navigation for a mobile transceiver like experience. 
Currently talKKonnect works with 4×20 Hitachi [HD44780](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf) LCD screen in parallel mode.  Other screens like 0.96" and 1.3" [OLED](https://learn.adafruit.com/adafruit-oled-displays-for-raspberry-pi)
with I2C interface is also currently supported. Currently SPI interfaced screens are not yet supported.

Low cost Class-D audio amplifiers like [PAM8403](https://www.instructables.com/id/PAM8403-6W-STEREO-AMPLIFIER-TUTORIAL/) or similar “D” class amplifiers, are recommended for talKKonnect builds.

A good shileded cable for microphone is recommended to keep the noise picked up to a minimum. I am currently experimenting with mems microphones for better audio.

Instead of the onboard sound card or USB Sound Card, you can also use a ReSpeaker compatiable HAT, or a ReSpeaker USB Sound Card with built in Amplifier and achieve great audio quality results in a compact form factor.
	
#### You can connect up to 4 LED indicators that can be build on the front panel of your build to show the following statuses ####
* Connected to a server and is currently online
* There are other participants logged into the same channel
* Currently in transmitting mode 
* Currently receiving an audio stream (someone is talking on the channel)
* Heart Beat to indicate that talKKonnect is running
* Currently in Voicetarget mode or Speaking in Normal mode for all clients on the channel to hear


### Software Features ###

* *Colorized LOGs* are shown on the debugging terminal for events as they happen in real time. Logging with line number, logging to file or screen or both. 
* Playing of configurable *alert sounds* as different events happen, such as a different sound when someone "joins" the channel and another sound for someone "leaving" the channel.
* Configurable *TTS prompts* to announce different events for those use special use cases where it is required.  
* Cusomizable *Roger Beep* sounds that are played at the end of each transmission. 
* *Muting* of The speaker when pressing PTT to prevent audio feedback and give a radio communication like experience to simulate simplex mode. Both simplex and duplex   settable in XML config. Duplex mode allows you to keep the speaker open for people to interrupt you while speaking. 
* LCD/OLED display can show *channel information, server information, who joined, who is speaking, last transmision received date and time, etc.* 
* Configuration is kept in a single *highly granular XML file*, where options can be enabled, disabled and customized.

### Common Information for the all the Pre-Made Images For Various Hardware Configurations ###
* We have for your convinience created a few different images that you can download and burn to your SD card so that you can get up and running quickly with a generic instance of talkkonnect working out of the box. Choose the image based on your hardware and use case. Using one of these images you will not need to follow all the complicated steps of installing and compiling everything from scratch if that seems daunting and overwhelming to you at first. 
* This is an easy way to start experimenting with talkkonnect in a matter of minutes. The ability to shorten the time and lessen the barrier of entry will allow you to see if talkkonnect suits your needs.
* The network settings are set as DHCP Client so your device should get an IP Address when you connect it to your DHCP enabled network.
* After you find the IP Address of your talkkonnect device from the DHCP leases section of your router you can log in over ssh using a tool like putty or equavilent on the standard ssh port 22 using the root user with password talkkonnect. The pi user is also accessable using the password raspberry.
* NOTICE!! When using these images Talkkonnect will already by started by systemd upon boot and run in a screen instance when you boot this image. There is no reason to manually start talkkonnect. By default with no changes in settings talkkonnect will connect to our community server. If you try start up talkkonnect by hand there will be 2 instances of talkkonnect that will clash with each other and you will be connected and disconnected from the server in a endless loop.
* Since talkkonnect is already running in the background (in a screen) upon boot, you can access the running console of talkkonnect by ssh (as root) into the raspberry pi device and at the command prompt type the command screen -r to see the console of the running talkkonnect. Press the <del> key to see a menu of the options available to you.
* We request that you to please edit the configuration file of talkkonnect.xml.  This file can be found in the directory /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/ Please change the XML tag <username>talkkonnect</username> to a name that describes you so that the members in the community channel can
  see who you are by name or callsign. If you do not change this setting your name in the channel will be shown as talkkonnect with some random numbers and letters so   as the keep the username unique by default. You cannot have more than 1 device per username connected to the server at the same time.
* By default the images of talkkonnect will connect to our community server at mumble.talkkonnect.com port 64738 using any unique username and the password talkkonnect
* You can join our channel and start chatting with us with voice and asking us questions or make suggestions we have a warm and welcomming group of enthusiastic individuals to help you with your questions. This is a good place to hang around and chat with like minded individuals.
* The images are divided into 2 broad categories (the ones that use the respeaker hat and the ones that do not)
* For those Non-Respeaker Images (Usb Sound Card or MEMS Microphone Images) Out of the box the standard configutation XML file is set to run in PC mode so no GPIO will initalized. 
* For those Respeaker Images (Rpi Zero or RPI 2/3 Images with Respeaker) Out of the box the standard configutation XML file is set to run in GPIO Mode and GPIO will initalized,
  this means the PTT Button and the LEDS on the 2 Mic Respeaker Hat will work right away. You will need to connect an external speaker to the HAT for these images.
* Feel Free to explore the various example talkkonnect.xml configurations that can be found in the directory /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs here you can find various configurations that work with LCD, OLED, LEDS and PUSH Button Switches. The files are named descriptively.
* To update to the lastest version release of talkkonnect in the image you can cd to the root directory and issue wget https://raw.githubusercontent.com/talkkonnect/talkkonnect/main/scripts/update-talkkonnect.sh then chmod +x ./update-talkkonnect.sh after that you can use the command /root/update-talkkonnect.sh to update to the latest version.	

## talKKonnect Version 2 Images	

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry PI tested on Pi (4 series) and USB Sound Card ###
* This Image Was Created On 17/April/2022 and Runs Bullseye Release along with talkkonnect version 2.13.06 Released 11/April/2022
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi 4 and USB Sound Card](https://drive.google.com/file/d/1s7Nh1sU2UcdT6xiXto8X8-cCnGJnlt8j/view?usp=sharing) 
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.	
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. 
* This image will work with LAN Cabled Ethernet connection out of the box
	
### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry PI tested on Pi (2/3 series) and USB Sound Card ###
* This Image Was Created On 09/January/2022 and Runs Bullseye Release along with talkkonnect version 2.09.19 Released 08/January/2022
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi and USB Sound Card](https://drive.google.com/file/d/128P2f7esB1cvx7Ma9NniduUM-4atvxqU/view?usp=sharing) 
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop.	
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. 
* This Image should work out of the box it also has serial console on the usb port for easy access to ssh thorough com port on windows and macOS
* This image will work with LAN Cabled Ethernet connection out of the box
* Since this image was created on Raspberry 3B+ Board if you use Raspberry 4 please note that you will have to do the following apt install firmware-brcm80211
  then you will have to use raspi-config to set the wifi country for the wifi to work and not sure if anything else is broken on RBP V4.
	
### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry 2/3 and RESPEAKER Compatable HAT ###
* This Image Was Created On 09/January/2022 and Runs Bullseye Release along with talkkonnect version 2.09.19 Released 08/January/2022	
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry 2/3 Respeaker Hat](https://drive.google.com/file/d/1c8f5EuKzgayuESFF3m35aPjC_26-4Gq9/view?usp=sharing) 
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This image has been configured to work with a Respeaker HAT out of the box so I2S, I2C and all required modules are installed and running. 
* The XML file is configured to run in rpi mode so GPIO will initalized, this is so that the respeaker will work with output sound on the headphone jack, led strip working and push button microswitch on the hat can be used for transmitting.    
* Since this image was created on Raspberry 3B+ Board if you use Raspberry 4 please note that you will have to do the following apt install firmware-brcm80211
  then you will have to use raspi-config to set the wifi country for the wifi to work.	
* For this image out of the box it will connect to a wifi with ssid network and password 1234567890 (if you are lazy you can do this)
* To Connect to your WIFI you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi.

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry Pi Zero 2W and RESPEAKER Compatable HAT ###
* This Image Was Created On 09/January/2022 and Runs Bullseye Release along with talkkonnect version 2.09.19 Released 08/January/2022	
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Zero 2W Respeaker Hat](https://drive.google.com/file/d/12jbEhgvDCkisCvOB92yGwW1JLXzFaBEO/view?usp=sharing) 
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This image has been configured to work with a Respeaker HAT out of the box so I2S, I2C and all required modules are installed and running. 
* The XML file is configured to run in rpi mode so GPIO will initalized, this is so that the respeaker will work with output sound on the headphone jack, led strip working and push button 
  microswitch on the hat can be used for transmitting.    
* This Image should work out of the box it also has serial console on the usb port for easy access to ssh thorough com port on windows and macOS
* For this image out of the box it will connect to a wifi with ssid network and password 1234567890 (if you are lazy you can do this)
* You can also use the serial console to log in as root and change the /etc/wpa_supplicant/wpa_supplicant.conf file to change it to your network for this you will need a micro usb port cable and plug it into your windows machine and access it over a tool like putty over the serial port 
* Alternativly you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi.
	
### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry pi 2/3  and IM69D130 Mems Microphone ### 
* This image is the initial release created a while ago and is not yet updated.
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry 3/4/Pi with IM69D130 Mems Microphone](https://drive.google.com/file/d/1s7Qjtj8XAfQmdr766WBvYZq4vjaQ7CZ3/view?usp=sharing) 
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This image has a custom kernel and used Instructions was found at https://github.com/Infineon/GetStarted_IM69D130_With_RaspberryPi
* This image has been configured to work with a IM69D130 Mems Microphone and the onboard raspberry pi sound card (3.5mm Jack) out of the box.
* The XML file is configured to run in rpi mode so GPIO will initalized, this is so that the Pin 11 XML tag value 17 when shorted to ground will act as the PTT button. 
* This mems microphone will enable you to have a small build with excellent sound quality whilst using the internal provided sound card in the raspberry pi.
* For the wiring of the microphone to Raspberry Pi See This [inmp411 wiring diagram](https://makersportal.com/shop/i2s-mems-microphone-for-raspberry-pi-inmp441)
* Since this image was created on Raspberry 3B+ Board if you use Raspberry 4 please note that you will have to do the following apt install firmware-brcm80211
  then you will have to use raspi-config to set the wifi country for the wifi to work.	
* For this image out of the box it will connect to a wifi with ssid network and password 1234567890 (if you are lazy you can do this)
* Alternativly you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi
	
### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image 32GB for Use with Orange Pi Zero ###
* This Image Was Created On 05/April/2022 and Runs Bullseye Release along with talkkonnect Version 2.13.01 Released Mar 6 2022
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Orange Pi Zero Onboard Sound](https://drive.google.com/file/d/19IZAi6zRIzWeS8i4eo8djSPOkNSF6C3v/view?usp=sharing) 
* You will need a to connect a microphone using the same circuit as the orange pi zero microphone hat.	
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This Image should work out of the box it also has serial console on the usb port for easy access to ssh thorough com port on windows and macOS
* This image will work with LAN Cabled Ethernet connection out of the box

### Installation Instructions For Raspberry Pi Boards (from Bash Script) ###
* Log in as root to your device via SSH
* cd to /root directory (if you are logged in as root you should already be in this directory)
* wget 	https://raw.githubusercontent.com/talkkonnect/talkkonnect/main/scripts/tkbuild.sh
* you will have downloaded the shell script tkbuild.sh
* chmod +x tkbuild.sh to make it executable
* run ./tkbuild.sh and wait for golang to install and talkkonnect to download along with all libraries automatically
* you may need to modify this script a little bit as versions of golang change	
	
### Installation Instructions For Raspberry Pi Boards (from Source code) ###

You have the choice of using a 32 bit or 64 bit os, the example below is for 32 bits.
	
Download the latest version of [Raspberry Pi OS Lite](https://downloads.raspberrypi.org/raspios_lite_armhf/images/raspios_lite_armhf-2022-04-07/2022-04-04-raspios-bullseye-armhf-lite.img.xz). 
	
At the time of making/updating this document latest image release date was 04/04/2022 (Kernel Version 5.15). 
Download the 297MB file and extract IMG file to some temporary directory.

It is recommended that you use the raspberry Pi Imager for Windows or any USB / SD card imaging software for Windows or your other OS. 
The best current option for windows is :
* [Raspberry Pi Imager](https://www.raspberrypi.org/software/)

After downloading a standard image and using the imaging tool, insert the SD card into your Raspberry Pi, connect the screen, keyboard and power supply and boot into the OS. 

As of the new release the pi user is removed please us a user you specified in the raspberry pi imager instead of the pi user.	

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

` apt install libopenal-dev libopus-dev libasound2-dev git ffmpeg screen `

##### Install prerequisite programs ##### 

To get the newer versions of golang used for this project I suggest installing a precompiled binary of golang. If you use apt-get to install golang at this moment you will get an older incompatible version of golang.

To install GO as required for this project on the raspberry pi. First with your browser look on the website https://golang.org/dl/ on your browser and choose the latest version for the 
arm archecture. At the time of this writing the version can be found at https://go.dev/dl/go1.18.linux-armv6l.tar.gz

Please Note that if you use apt-get to install golang instead of follow the recommended instructions in this blog you may get some errors like the following error when compiling 
BackLightTime.Reset undefined (type * time.Ticker has no field or method Reset) . This is just an example of how changes in the language break old code that was perhaps not written properly. For best results stick to the latest version of golang.

As root user Get the link and use wget to download the binary to your talkkonnect

` cd /usr/local `

` wget https://golang.org/dl/go1.18.linux-armv6l.tar.gz `

` tar -zxvf go1.18.linux-armv6l.tar.gz `

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
│Blog at www.talkkonnect.com, source at github.com/talkkonnect   │
└────────────────────────────────────────────────────────────────┘
Software Ver 2.13.06 Rel Apr 11 2022 (Latest Release)
	
````

##### I2C OLED Screen Installation #####
For those of you who wish to use a 0.96 or 1.3 inch OLED screen follow the instructions below (logged in as root)

[enabling i2c](https://www.raspberrypi-spy.co.uk/2014/11/enabling-the-i2c-interface-on-the-raspberry-pi/) read and Follow Step 1 - Enable I2C Interface.

For detecting the address of your screen install the tool below

` apt-get install -y i2c-tools `

Then using i2cdetect to detect your screen following the instructions on the same page under the section Testing Hardware (Optional)

Once you get the address note that it will be in HEX you will have to convert this address to decimal to put in the talkkonnect.xml file
under the xml tag  <oleddefaulti2caddress>60</oleddefaulti2caddress>

In the example above I got the address 3c from i2c tools and converted that to decimal value 60. 


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

USB sound device can also be set in local profile (this step is not necessary if you have used the global configuration above)

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
│     _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _     │
│    | '_ ` _ \ / _` | | '_ \  | '_ ` _ \ / _ \ '_ \| | | |    │
│    | | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |    │
│    |_| |_| |_|\__,_|_|_| |_| |_| |_| |_|\___|_| |_|\__,_|    │
├─────────────────────────────┬────────────────────────────────┤
│ <Del> to Display this Menu  | <Ctrl-C> to Quit talkkonnect   │
├─────────────────────────────┼────────────────────────────────┤
│ <F1>  Channel Up (+)        │ <F2>  Channel Down (-)         │
│ <F3>  Mute/Unmute Speaker   │ <F4>  Current Volume Level     │
│ <F5>  Digital Volume Up (+) │ <F6>  Digital Volume Down (-)  │
│ <F7>  List Server Channels  │ <F8>  Start Transmitting       │
│ <F9>  Stop Transmitting     │ <F10> List Online Users        │
│ <F11> Playback/Stop Stream  │ <F12> For GPS Position         │
├─────────────────────────────┼────────────────────────────────┤
│<Ctrl-B> Reload XML Config   │ <Ctrl-C> Stop Talkkonnect      │
│<Ctrl-D> Debug Stacktrace    │ <Ctrl-E> Send Email            │
├─────────────────────────────┼────────────────────────────────┤
│<Ctrl-F> Conn Previous Server│<Ctrl-G> Send Repeater Tone     │
│<Ctrl-H> XML Config Checker  │<Ctrl-I> Traffic Record         │
│<Ctrl-J> Mic Record          │<Ctrl-K> Traffic & Mic Record   │
│<Ctrl-L> Clear Screen        │<Ctrl-M> Radio Channel (+)      │
│<Ctrl-N> Next Server         │<Ctrl-O> Ping Servers           │
│<Ctrl-P> Panic Simulation    │<Ctrl-R> Repeat TX Loop Test    │
│<Ctrl-S> Scan Channels       │<Ctrl-T> Thanks/Acknowledgements│
│<Ctrl-U> Show Uptime         │<Ctrl-V> Display Version        │
│<Ctrl-X> Dump XML Config     │                                │
├─────────────────────────────┼────────────────────────────────┤
│  Visit us at www.talkkonnect.com and github.com/talkkonnect  │
│  Thanks to Global Coders Co., Ltd. for their sponsorship     │
└──────────────────────────────────────────────────────────────┘
````
### Explanation of the history and reasons for creating talkkonnect 
[youtube-video](https://youtu.be/nLmHM48SqFs)

#### The Accounts Section
* The account section can have multiple accounts, talkkonnect will look for the first account with the xml tag default = "true" and attempt to connect to that server 
* When talkkonnected is connected to a server you can cycle through accounts in which enabled = "true" by pressing CTRL-N, talkkonnect will connect to the next enabled server in the list
* Talkkonnect will not attempt to connect to a server that has the account tag set default = "false" 
* The tag account name is just used to identify the server for logging purposes 
* The serverandport tag is for the server FQDN or IP address followed by  " (colon) and the port of mumble is running on for that particlar server.
* The username tag is used for identifying yourself on the mumble server and for authentication 
* The password tag is used if the mumble server requires password authentication 
* The insecure tag should be set as true if the server you are connecting to does not require a certificate 
* The certificate tag should contain the full path to your previously generated certificate which is usually a file with the extension of pem  
* The channel tag should only be populated want to connect to a specific channel other than the root channel on startup
* The tokens list for each account for autorization to token protected channels
* The voicetargets IDs and their corresponding users and channels
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
* Each event such as when a person joins a channel, leaves a channel or sends a message into the channel can be configured seperately.
* The filenameandpath tag should contain the the full path and filename of the WAV file you wish to play for each event 
* The event tag is used to play an audible alert when there are changes of other users statuses 
* The alert tag is used to play an WAV file into the stream to the receiving party upon a user generated panic request
* The rogerbeep tag is used to define the WAV file to play at the end of every transmission 
* The tag name stream, This function is very powerful and can be used to define a local file or network stream that will be played into the mumble channel upon pressing the F11 key. Very useful for debugging.

##### The TXTIMEOUT section
* The txtimeout tag is used to limit the length of a single transmission in seconds. This tag is useful when used as a repeater between RF and mumble.

##### The API Section
* API section enables the user to granually control which remote control functions are available over http within the local network 
* The tag apilisten port defines the port that talkkonnect should listen and respond to remote control http requests
* HTTP API for Channel Up           http://{your-talkkonnect-ipaddress}:8080/?command=channelup
* HTTP API for Channel Down         http://{your-talkkonnect-ipaddress}:8080/?command=channeldown
* HTTP API to  Mute/UnMute Toggle   http://{your-talkkonnect-ipaddress}:8080/?command=mute-toggle
* HTTP API to  Mute Speaker         http://{your-talkkonnect-ipaddress}:8080/?command=mute
* HTTP API to  Unmute Speaker       http://{your-talkkonnect-ipaddress}:8080/?command=unmute 
* HTTP API for Volume UP            http://{your-talkkonnect-ipaddress}:8080/?command=volumeup
* HTTP API for Volume Down          http://{your-talkkonnect-ipaddress}:8080/?command=volumedown
* HTTP API to  Start Transmitting   http://{your-talkkonnect-ipaddress}:8080/?command=starttransmitting
* HTTP API to  Stop Transmitting    http://{your-talkkonnect-ipaddress}:8080/?command=stoptransmitting
* HTTP API to  Play/Stop Stream     http://{your-talkkonnect-ipaddress}:8080/?command=stream-toggle
* HTTP API to  Request GPS Position http://{your-talkkonnect-ipaddress}:8080/?command=gpsposition
* HTTP API to  Send Email           http://{your-talkkonnect-ipaddress}:8080/?command=sendemail
* HTTP API for Previous Server      http://{your-talkkonnect-ipaddress}:8080/?command=connpreviousserver
* HTTP API for Next Server          http://{your-talkkonnect-ipaddress}:8080/?command=connnextserver
* HTTP API for Request Panic Sim    http://{your-talkkonnect-ipaddress}:8080/?command=panicsimulation
* HTTP API to  Play Repeater Tone   http://{your-talkkonnect-ipaddress}:8080/?command=playrepeatertone
* HTTP API to  Set Voice Target 0   http://{your-talkkonnect-ipaddress}:8080/?command=setvoicetarget&id=0
* HTTP API to  Set Voice Target 1   http://{your-talkkonnect-ipaddress}:8080/?command=setvoicetarget&id=1

##### The PrintVariables Section
* This function is useful for debugging the values read from each section of the config xml file. You can control which section is shown. This command is tied to the CTRL-X key

##### The MQTT Section
* Talkkonnect can be remotely controlled by an public or local MQTT Server
* This eliminates the problem of controlling those talkkonnect devices that are in NATTED networks all over the internet
* You can subscribe to the mqtt server topic of your choice
* With MQTT you can remote control talkkonnect as well as Relays to control external devices 

Below are Valid Commands for MQTT
* channelup
* channeldown
* mute-toggle
* mute
* unmute
* volumeup
* volumedown
* starttransmitting
* stoptransmitting
* stream-toggle
* gpsposition
* sendemail
* connpreviousserver
* connnextserver
* panicsimulation
* repeattxLoop
* scanchannels
* attentionled:on
* attentionled:off
* attentionled:blink
* relay1:on
* relay1:off
* relay1:pulse
* playerepeatertone

For Example on the topic thailand/bangkok/company/talkkonnect/attentionled:on will turn on the LED to get the attentionled
of a user. 

Another Example on the topic thailand/bangkok/company/talkkonnect/relay1:pulse will simulate a push button for example to
open the door for a an access control system

For the above example to work you will have to specify the gpio pin in the <lights> section of the xml file
<attentionledpin></attentionledpin>
<relay1pin></relay1pin>

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

##### The USB Keyboard Section
* You can enable a wired/wireless USB numpad here for voice targeting and other direct commands to a headless talkkonnect

##### The KeyboardCommands Section
* You can define the command associated with each tty key or each key on your USB Numpad here Example 
* TTY Means the keyboard that you use when you ssh into the box or use talkkonnect in a terminal
* USB means using an external Wired or Wireless USB Numeric Key Pad directly on the Raspberry Pi USB port to control talkkonnect in headless mode
* Below are the commands you can map to your TTY Keyboard or the USB Keyboard 
* channelup
* channeldown
* serverup
* serverdown
* mute
* unmute
* mute-toggle
* stream-toggle
* volumeup
* volumedown
* setcomment
* transmitstart
* transmitstop
* record
* setvoicetarget
     

## Questions & Contributing 
We invite interested individuals to provide feedback and improvements to the project. 
To speak to us you can connect with a standard mumble client (android/iphone/wnidows/linux) to our community server to have a chat or ask questions 
at mumble.talkkonnect.com port 64738 you can use any username with the password talkkonnect

Currently we do not have a WIKI so send feedback to <suvir@talkkonnect.com> or open and Issue in github
you can also check my blog  [www.talkkonnect.com](https://www.talkkonnect.com) for updates on the project
	
Please visit our [blog](www.talkkonnect.com) for our blog or [github](github.com/talkkonnect) for the latest source code and our [facebook](https://www.facebook.com/talkkonnect) page for future updates and information. 

Thank you all for your kind feedback sent along with some pictures and use cases for talkkonnect.

## License 
[talKKonnect](http://www.talkkonnect.com) is open source and available under the MPL V2.00 license.

<suvir@talkkonnect.com> Updated 11/04/2022 talkkonnect version 2/13/06 is the latest release as of this writing.
