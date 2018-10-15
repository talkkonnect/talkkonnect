How to build? First you will need some libraries 

First Create a user called mumble
* adduser --disabled-password --disabled-login --gecos "" mumble
* usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio mumble

Then Install required librarys: LibOpenAl, LibAsound 
* on raspbian apt-get install libasound-dev libopenal-dev
* export the GOPATH & GOBIN directorys using env variables:
* export GOPATH=/home/mumble/gocode
* export GOBIN=/home/mumble/bin
* go get github.com/talkkonnect/talkkonnect
* cd $GOPATH/src/github.com/talkkonnect/talkkonnect
* go build -o /home/mumble/bin/talkkonnect cmd/talkkonnect/main.go

Before Running talkkonnect you will need to Create a certificate

This is optional, mainly if you want to register your talkkonnect against a mumble server and apply ACLs.

* su mumble
* cd ~
* openssl genrsa -aes256 -out key.pem
Enter a simple passphrase, its ok, we will remove it shortly...
* openssl req -new -x509 -key key.pem -out cert.pem -days 1095
Enter your passphrase again, and fill out the certificate info as much as you like, its not really that important if you're just hacking around with this.
* openssl rsa -in key.pem -out nopasskey.pem
Enter your password for the last time.
* cat nopasskey.pem cert.pem > mumble.pem

