# Talkkonnect
A Headless Mumble Client Based on Raspberry Pi (Intercom)

talKKonnect is a headless self-contained mumble Push to Talk (PTT) client with a mobile transceiver form factor complete with LCD, Channel and Volume control.
This project is a fork of talkiepi by Daniel Chote. You can find Daniel’s page [here](http://projectable.me/)
This project was developed to be run on a Raspberry Pi 3 B+.

## Instructions
### Raspberry Pi

1. Get the latest version of [Raspbian Lite](https://www.raspberrypi.org/downloads/raspbian/) (latest version is currently 2018-06-27 Kernel Version 4.14)
1. Follow the instructions [here](https://www.raspberrypi.org/documentation/installation/installing-images/README.md) to install the raspbian image onto an SD card.
1. Insert the SD card into your Raspberry Pi 3 B+, connect the screen, keyboard, power supply, and boot into the OS.
1. Log in with the default credentials (user: `pi` password: `raspberry`)
1. Set the root password with: 
    ```bash
    sudo passwd root
    ```
1. Log out of the account pi and log into the root account with your newly set password.
1. Run `raspi-config` and expand the file system 
1. Enable SSH server under 'Interfacing Options' ([More Info](https://www.raspberrypi.org/documentation/remote-access/ssh/README.md)
1. Edit the file `/etc/ssh/sshd_config` and change the following line to allow root login: `PermitRootLogin prohibit-password` to `PermitRootLogin Yes`
1. Reboot the Raspberry Pi: `sudo reboot`
1. Now you should be able to log in remotely via ssh using the root account and continue the installation
1. Add a new user mumble:
    ```bash
    adduser --disabled-password --disabled-login --gecos "" mumble
    ```
1. Add the new user to the appropriate groups:
    ```
    usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio mumble
    ```

### Installation
1. Install required packages:
    ```bash
    apt-get install golang libopenal-dev libopus-dev libasound2-dev git ffmpeg omxplayer screen
    ```
1. Create directories:
    ```bash
    su mumble
    cd /home/mumble
    mkdir /home/mumble/gocode
    mkdir /home/mumble/bin
    ```
1. Set environment variables:
    ```bash
    export GOPATH=/home/mumble/gocode
    export GOBIN=/home/mumble/bin
    ```
1. Install dependencies:
    ```bash
    cd $GOPATH
    go get github.com/talkkonnect/gumble
    go get github.com/talkkonnect/talkkonnect
    ```
1. Compile!
    ```
    cd $GOPATH/src/github.com/talkkonnect/talkkonnect
    go build -o /home/mumble/bin/talkkconnect cmd/talkkonnect/main.go
    ```

### Optional

If you want to register your talkkonnect against a mumble server and apply ACLs, before running talkkonnect you will need to create a certificate:

```bash
su mumble
cd ~
openssl genrsa -aes256 -out key.pem
#Enter a simple passphrase, its ok, we will remove it shortly...
openssl req -new -x509 -key key.pem -out cert.pem -days 1095
#Enter your passphrase again, and fill out the certificate info as much as you like, its not really that important if you're just hacking around with this.
openssl rsa -in key.pem -out nopasskey.pem
#Enter your password for the last time.
cat nopasskey.pem cert.pem > mumble.pem
```
## Improvements
### Hardware Improvements

* Use External Mic and push buttons on the device for Channel Up/Down navigation
* Has a 4×20 LCD Screen
* Has a built-in Amplifier using the TDA2030 Chip
* Has a volume adjustment knob
* Self-contained unit with all power supplies built-in
* Female RJ45 connector at the back for network connectivity
* 4 LEDs that show status such as online, other participants in the channel, transmit mode and voice activity

### Software Improvements

* Colorized logs to debug events as they happen
* Can play an alert sound as events happen
* A Roger Beep-like effect that lets the remote party that you have released the PTT button
* Mutes Speaker when pressing PTT to prevent feedback and give a radio-like experience
* Channel seek
* LCD with useful information for showing status, channel joined, who is speaking, etc.

More information can be found at [www.talkkonnect.com](www.talkkonnect.com)

<suvir@talkkonnect.com>
