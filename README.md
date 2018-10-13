A Headless Mumble Client Based on Raspberry Pi (Intercom)

talKKonnect is a headless self contained mumble Push to Talk (PTT) client with a mobile transceiver form factor complete with LCD, Channel and Volume control.
This project is a fork of talkiepi by Daniel Chote. You can find Daniel’s page here http://projectable.me/
Developed to be run on Rasperry Pi 3 b+


Hardware Improvements

Use External Mic and push buttons on the device for Channel Up/Down navigation
Has a 4×20 LCD Screen
Has a built in Amplifier using the TDA2030 Chip
Has a volume adjustment knob
Self contained unit with all power supplies built in
Female RJ45 connector at the back for network connectivity
4 LEDs that show status such as online, other participants in channel, Transmit Mode and an LED that flashes as there is voice activity

Software Improvements

Colorized LOGS on debugging terminal for events as they happen
Can play an alert sound as events happen
A Roger Beep like effect that lets the remote party that you have released the PTT button
Mutes Speaker when pressing PTT to prevent feedback and give a radio like experience
Channel seek
LCD with useful information for showing status, channel joined, who is speaking, etc.

Installation Instructions

1. Go to get the latest version of RASPBIAN LITE
at the time of making this document version downloaded was June 2018 Release Date 2018-06-27 Kernel Version 4.14
2. Download the ZIP file and extract it to get an img file
3. Use a software on windows called RUFUS to write the image to a SD Card (Becareful don't choose the wrong drive)
4. After Done Insert the SD card into your Raspberry Pi 3 b+ connect the screen, keyboard and power supply and boot into the OS.
5. Log in as user pi password raspberry (this is the default username and password for a fresh install of Raspbian)
6. Do a sudo passwd root to set the root password 
7. Log out of the account pi and log into the root account with your newly set password.
8. run raspi-config and expand the file system 
9. Next go to interfacing options and enable ssh server
10. Edit the file /etc/ssh/sshd_config Change the line #PermitRootLogin prohibit-password to PermitRootLogin Yes and reboot the raspberry pi
11. Now you should be able to log in remotely via ssh using the root account and continue the installation
12. adduser --disabled-password --disabled-login --gecos "" mumble
13. usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio mumble
14. apt-get install golang libopenal-dev libopus-dev libasound2-dev git ffmpeg omxplayer screen
15. su mumble
16. cd /home/mumble
17. mkdir /home/mumble/gocode
18. mkdir /home/mumble/bin
19. export GOPATH=/home/mumble/gocode
20. export GOBIN=/home/mumble/bin
21. cd $GOPATH
22. go get github.com/talkkonnect/gumble
23. go get github.com/talkkonnect/talkkonnect
24. cd $GOPATH/src/github.com/suvirkumar/talkkonnect
25. go build -o /home/mumble/bin/talkkconnect cmd/talkkonnect/main.go

More information can be found at www.talkkonnect.com

<suvir@talkkonnect.com>
