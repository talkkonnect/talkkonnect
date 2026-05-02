## Quick Start Readymade talKKonnect Downloadable Images for SD Cards

This page details how to configure talKKonnect for a hardware build using a ready-made image.

For details on how to build and setup talKKonnect on any generic Linux system, visit [Getting Started with talKKonnect](./getting-started.md).

Once you're set up, check out [Configuration and Running](./running-talkkonnect.md) for details on how to configure and run talKKonnect.

----

### Notes on How to Use Raspberry Pi Imager Otherwise the Ready Made images below Will NOT WORK!!
* Click Choose Device (select No Filtering)
* Click Choose OS (Use Custom) choose the talkkonnect img downloaded
* Choose Storage
* Click Next
* Would you like to apply OS customization settings? NO!!!!!
* All existing data on Mass Storage Device will be Erased. Yes

### talkkonnect Version 2 (64 Bit) Quick Download Link for Pre-Made SD Card Image for Use with Raspberry PI tested on Pi 4 series and USB Sound Card ###
* This Image Was Created On 12/Janruary/2024 and Runs Debian Bookworm (Released December 11th 2023) along with talkkonnect version 2.37.01 Released 31/December/2023
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi 4 and USB Sound Card](https://drive.google.com/file/d/1P_2JU4MxTwEF-We1ODEAphg9rKyPiHhT/view?usp=sharing)
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. The GPIOs are mapped to the [circuit diagram](https://github.com/talkkonnect/talkkonnect/blob/main/circuit-diagram/Schematic_talkkonnect_2023-12-13.pdf)
* This image will work with LAN Cabled Ethernet connection out of the box

### talkkonnect Version 2 (64BIT) Quick Download Link for Pre-Made SD Card Image for Use with Raspberry PI tested on Pi 3 Series and USB Sound Card ###
* This Image Was Created On 12/Janruary/2024 and Runs Debian Bookworm (Released December 11th 2023) along with talkkonnect version 2.37.01 Released 31/December/2023
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi 3 and USB Sound Card](https://drive.google.com/file/d/15--tFwJnKSmWstChioWGrom563KE1vmb/view?usp=sharing)
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. The GPIOs are mapped to the [circuit diagram](https://github.com/talkkonnect/talkkonnect/blob/main/circuit-diagram/Schematic_talkkonnect_2023-12-13.pdf)
* This image will work with LAN Cabled Ethernet connection out of the box


### talkkonnect Version 2 (32 BIT) Quick Download Link for Pre-Made SD Card Image for Use with Raspberry PI tested on Pi (2/3 series) and USB Sound Card ###
* This Image Was Created On 12/Janruary/2024 and Runs Debian Bookworm (Released December 11th 2023) along with talkkonnect version 2.37.01 Released 31/December/2023
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi 3 (32 Bit OS) and USB Sound Card](https://drive.google.com/file/d/1iJdRzocS_SR2v_xJKqKWhct5XMWLMomt/view?usp=sharing)
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. The GPIOs are mapped to the [circuit diagram](https://github.com/talkkonnect/talkkonnect/blob/main/circuit-diagram/Schematic_talkkonnect_2023-12-13.pdf)
* This image will work with LAN Cabled Ethernet connection out of the box


### talkkonnect Version 2 (64 BIT) Quick Download Link for Pre-Made SD Card Image for Use with Raspberry Pi Zero2W and USB Sound Card Connected Via USB OTG ###
* This Image Was Created On 26/February/2024 and Runs Debian Bookworm (Released December 11th 2023) along with talkkonnect version 2.40.01 Released Feb 2024
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi Zero 2W and USB Sound Card](https://drive.google.com/file/d/1HL4qqGyLlIxD5EPxDEj_DENFg7OVObyz/view?usp=sharing)
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.
* This image has been configured to work with a USB CM108 Sound Card and GPIO will work out of the box for PTT and LEDs. The GPIOs are mapped to the [circuit diagram](https://github.com/talkkonnect/talkkonnect/blob/main/circuit-diagram/Schematic_talkkonnect_2023-12-13.pdf)
* This image will work with LAN Cabled Ethernet connection out of the box

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Raspberry Pi 3,3A,3B+,Zero2W and RESPEAKER Compatable HAT ###
* This Image Was Created On 12/Janruary/2024 and Runs Debian Bookworm (Released December 11th 2023) along with talkkonnect version 2.37.01 Released 31/December/2023
* The Respeaker Drivers are compiled and audio for receiving and transmitting are working. The LEDs are working and so is the PTT Button on the Respeaker HAT.
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Raspberry Pi 3 Series and Zero2W (32 Bit OS) and Respeaker Hat](https://drive.google.com/file/d/1XPuzVnlYxRnatF4EvBT9r1aME9zSTKHg/view?usp=sharing)
* The XML file is configured to run in rpi mode so GPIO will initalized, this is so that the respeaker will work with output sound on the headphone jack, led strip working and push button microswitch on the hat can be used for transmitting.    
  then you will have to use raspi-config to set the wifi country for the wifi to work.
* For this image out of the box it will connect to a wifi with ssid network and password 1234567890 (if you are lazy you can do this)
* To Connect to your WIFI you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi.

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Orange Pi Zero with external USB sound card ###
* This Image Was Created On 20/Janruary/2024 and Runs Armbian Bookworm along with talkkonnect version 2.37.01 Released 31/December/2023
* You will need a CM-108 or equavilent sound card plugged in before booting for this image to work otherwise it will connect and disconnect to the server in an endless loop. Vention Cards are recommended as cm108 card has a lot of noise.
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Orange Pi Zero External USB Sound Card](https://drive.google.com/file/d/1HL4qqGyLlIxD5EPxDEj_DENFg7OVObyz/view?usp=sharing)
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This Image should work out of the box with a micro-USB OTG Cable to USB Sound Card
* The XML file is configured to run in rpi mode so GPIO will initalized, this means that PTT GPIO key is enabled, Rotary Encoder All Configured.
* For this image please use the raspberry pi imager to set the ssid and password of your WIFI network before writing the SD card or alternatively you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi.

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Orange Pi Zero with internal onboard sound card ###
* This Image Was Created On 20/Janruary/2024 and Runs Armbian Bookworm along with talkkonnect version 2.37.01 Released 31/December/2023
* You will need a to connect a microphone using the same circuit as the orange pi zero microphone hat.
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Orange Pi Zero Onboard Sound](https://drive.google.com/file/d/1rA5P-wpZPDByQHtei0S2qSwaAQ65NmpT/view?usp=drive_link)
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This Image should work out of the box it also has serial console on the usb port for easy access to ssh thorough com port on windows and macOS
* The XML file is configured to run in rpi mode so GPIO will initalized, this means that PTT GPIO key is enabled, Rotary Encoder and OLED Screen All Configured.
* This image will work with LAN Cabled Ethernet connection out of the box then you will have to use raspi-config to set the wifi country for the wifi to work.
* For this image out of the box it will connect to a wifi with ssid network and password 1234567890 (if you are lazy you can do this)
* To Connect to your WIFI you can also create the wpa_supplicant.conf file and put it in the /boot/ folder on windows before inserting the card into your raspberry pi.

### talkkonnect Version 2 Quick Download Link for Pre-Made SD Card Image for Use with Orange Pi Zero 3 H618 Chip with mtech exteral usb mic ###
* This Image Was Created On 02/June/2024 and Runs Debian Bookworm along with talkkonnect version 2.41.01 Released 12/April/2024
* You will need a to connect usb microphone with cm108 chip and using vol Up as media key for PTT.
* [Click Here to Download Pre-Configured SD Card Image for Talkkonnect Version 2 for Orange Pi Zero 3 H618 Chip Mtech USB Microphone](https://drive.google.com/file/d/1rCTiy0HVEl19ie2KDOHi5XymyyTtgThS/view?usp=drive_link)
* This image uses the standard 32 Bit Sampling and will work properly with all mumble clients on windows, android and iphone with good quality sound.
* This Image should work out of the box it also has serial console on the usb port for easy access to ssh thorough com port on windows and macOS
* The XML file is configured to run in rpi mode so GPIO will initalized, this means that PTT GPIO 13 key is enabled.
* This image will work with LAN Cabled Ethernet connection out of the box then you will have to use orangpi-config to set the wifi country for the wifi to work.

### Installation Instructions For Raspberry Pi Boards (from Bash Script) ###
* Install raspberry pi os using instructions from [here](https://www.raspberrypi.com/software/operating-systems/)
* Preferrably with modern boards use the 64 Bit of the Raspberry Pi OS Lite version
* Follow the steps to burn and boot from SD Card
* Log in as root to your device via SSH
* cd to /root directory (if you are logged in as root you should already be in this directory)
* https://raw.githubusercontent.com/talkkonnect/talkkonnect/main/scripts/tkbuild.sh
* chmod +x tkbuild.sh
* run ./tkbuild.sh and wait for golang to install and talkkonnect to download along with all libraries automatically
* You will need to copy and modify the XML Sample from [here](https://github.com/talkkonnect/talkkonnect/blob/main/sample-configs/talkkonnect-version2-usb-gpio-example.xml) and keep in the directory
  /home/talkkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml
* After that you will have to configure alsa as shown in the audio configuration section [below](https://github.com/talkkonnect/talkkonnect#audio-configuration) for talkkonnect to work

### Common Information for the all the Pre-Made Images For Various Hardware Configurations ###
* With all the updates I cannot possibly make all config sample files up to date so as a guideline for raspberry pi boards please always look to the file talkkonnect-version2-usb-gpio-example.xml and talkkonnect-current-sample.xml for the latest tags to copy and implement them in your builds.
* We have for your convinience created a few different images that you can download and burn to your SD card so that you can get up and running quickly with a generic instance of talkkonnect working out of the box. Choose the image based on your hardware and use case. Using one of these images you will not need to follow all the complicated steps of installing and compiling everything from scratch if that seems daunting and overwhelming to you at first.
* This is an easy way to start experimenting with talkkonnect in a matter of minutes. The ability to shorten the time and lessen the barrier of entry will allow you to see if talkkonnect suits your needs.
* The network settings are set as DHCP Client so your device should get an IP Address when by cabled LAN you connect it to your DHCP enabled network.
* After you find the IP Address of your talkkonnect device from the DHCP leases section of your router you can log in over ssh using a tool like putty or equavilent on the standard ssh port 22 using the root user with password talkkonnect.
* NOTICE!! When using these images Talkkonnect will already by started by systemd upon boot and run in a screen instance when you boot this image. There is no reason to manually start talkkonnect. By default with no changes in settings talkkonnect will connect to our community mumble server. If you try start up talkkonnect manually there will be 2 instances of talkkonnect that will clash with each other and you will be connected and disconnected from the server in a endless loop.
* NOTICE!! If you use talkkonnect and notice that the talkkonnect client is connecting to the server and then disconnecting after a few seconds and that this is in an endless loop. Please check that you have plugged in the USB sound card and the sound card configurations match your sound card.
* Since talkkonnect is already running in the background (in a screen) upon boot, you can access the running console of talkkonnect by ssh (as root) into the raspberry pi device and at the command prompt type the command screen -r to see the console of the running talkkonnect. Press the del key to see a menu of the options available to you.
* We request that you to please edit the configuration file of talkkonnect.xml.  This file can be found in the directory /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/ Please change the XML tag <username></username> to a name that describes you so that the members in the community channel can see who you are by name or callsign. If you do not change this setting your name in the channel will be shown as talkkonnect with your mac address to keep the username unique by default. You cannot have more than 1 device per username connected to the server at the same time.
* By default the images of talkkonnect will connect to our community server at mumble.talkkonnect.com port 64738 using any unique username there is no need for a password to connect to the community mumble server
* You can join our channel(s) HAM-CB on the community server and start chatting with us with voice and asking us questions or make suggestions we have a warm and welcomming group of enthusiastic individuals to help you with your questions. This is a good place to hang around and chat with like minded individuals.
* The images are divided into 2 broad categories (the ones that use the respeaker hat, the ones that use onboard sound and the ones that use USB Sound cards)
* The images also are available in both 32 and 64 bit of the underlying operating systems written on to the SD card images.
* For those Respeaker Images (Rpi Zero or RPI 2/3 Images with Respeaker) Out of the box the standard configutation XML file is set to run in GPIO Mode and GPIO will initalized, this means the PTT Button and the LEDS on the 2 Mic Respeaker Hat will work right away. You will need to connect an external speaker to the HAT for these images.
* Feel Free to explore the various example talkkonnect.xml configurations that can be found in the directory /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/sample-configs here you can find various configurations that work with LCD, OLED, LEDS and PUSH Button Switches. The files are named descriptively. See the talkkonnect-version2-usb-gpio-example.xml
  as an example. 