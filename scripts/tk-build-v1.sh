#!/bin/bash

(

## talKKonnect Raspberry Pi or PC installation script By Zoran D

## Update
SECONDS=0
apt-get update
#apt-get -y dist-upgrade

## Install the dependencies required for talkkonnect
apt-get -y install curl screen pkg-config git gccgo libopenal-dev libopus-dev \
opus-tools pulseaudio pulseaudio-utils libasound2-dev ffmpeg screen pkg-config \
psmisc

## Add talkkonnect user to the system

if id -u "talkkonnect" >/dev/null 2>&1; then
   echo "=> talkkonnect user already exist"
else
   echo "=> talkkonnect user will now be created"

adduser --disabled-password --disabled-login --gecos "" talkkonnect

usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,pulse,pulse-access talkkonnect

#gpio group is for Raspberry Pi OS only. Other systems use dialout group for gpio.
usermod -a -G gpio talkkonnect

fi

#if running talkkonnect as a root, add root to pulse-access group for pulseaudio to work.
usermod -a -G pulse-access root

# Create the necessary directory structure under /home/talkkonnect/
cd /home/talkkonnect/
mkdir -p /home/talkkonnect/gocode
mkdir -p /home/talkkonnect/bin

## Make the log file and install log
[ ! -f /var/log/talkkonnect.log ] && touch /var/log/talkkonnect.log
[ ! -f /var/log/tk-install.log ] && touch /var/log/talkkonnect.log
truncate -s 0 /var/log/talkkonnect.log
truncate -s 0 /var/log/tk-install.log

# Check the latest Golang version
GOLANG_LATEST_STABLE_VERSION=$(curl -s https://go.dev/VERSION?m=text | grep go)

# Check architecture
arch=`dpkg --print-architecture`

#Check if Golang is installed
if ! [ -x "$(command -v go)" ]; then
   echo "=> Golang is not installed. Downloading and installing the latest Golang now." >&2

cd /usr/local

if [[ "$arch" == arm64 ]]; then
   echo "=> 64-bit ARM platform"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz

elif [[ "$arch" == armhf ]] || [[ "$arch" == armel ]]; then
   echo "=> 32-bit ARM platform"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-arm6l.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-arm6l.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-arm6l.tar.gz

elif [[ "$arch" == amd64 ]]; then
   echo "=> 64-bit PC platform"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-amd64.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-amd64.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-amd64.tar.gz

elif [[ "$arch" == i*86 ]]; then
   echo "=> 32-bit PC platform"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-i386.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-i386.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-i386.tar.gz

else
   echo "=> Unknown platform. Don't know how to install Golang."
   echo "=> Fix your Golang first, then try building talkkonnect again."
   echo "=> Exiting talkkonnect installation!";
exit 1

fi

fi

if grep -R "/usr/local/go/bin" ~/.bashrc > /dev/null
then
   echo "=> path to go/bin already defined"
else
   echo export PATH=$PATH:/usr/local/go/bin >> ~/.bashrc
fi

if grep -R "GOPATH=/home/talkkonnect/gocode" ~/.bashrc > /dev/null
then
   echo "=> path to tk go code folder already defined"
else
   echo export GOPATH=/home/talkkonnect/gocode >> ~/.bashrc
fi

if grep -R "GOBIN=/home/talkkonnect/bin" ~/.bashrc > /dev/null
then
   echo "=> path to tk go bin folder already defined"
else
echo export GOBIN=/home/talkkonnect/bin >> ~/.bashrc
fi

if grep -R GO111MODULE="auto" ~/.bashrc > /dev/null
then
   echo "=> GO111MODULE statement defined"
else
   echo export GO111MODULE="auto" >> ~/.bashrc
fi

if grep -R "alias tk='cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/'" ~/.bashrc > /dev/null
then
    echo "=> alias to talkkonnect program folder already defined"
else
   echo "alias tk='cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/'" >> ~/.bashrc
fi

## Set up GOENVIRONMENT
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/talkkonnect/gocode
export GOBIN=/home/talkkonnect/bin
export GO111MODULE="auto"

#Check if talkkonect program code dir exist
if ! [ -z "$(ls -A /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect 2>/dev/null)" ] ; then
   echo "=> Previous version of talkkonnect code exist in the system. It will be overwritten."
else
   echo "=> talkkonnect will be installed now."
fi

## Get the latest source code of talkkonnect from github.com
echo "=> Installing talkkonnect with traditional method avoiding go get. It changed in Golang 1.22."
cd $GOPATH
mkdir -p /home/talkkonnect/gocode/src/github.com/talkkonnect
cd /home/talkkonnect/gocode/src/github.com/talkkonnect
git clone https://github.com/talkkonnect/talkkonnect 2>/dev/null
cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect
go mod init
go mod tidy

#Check if talkkonnect binary file exist
if [ -f /home/talkkonnect/bin/talkkonnect 2>/dev/null ] ; then
   echo "=> Previous version of talkkonnect bin exist. It will be overwritten."
else
   echo "=> talkkonnect bin will build now."
fi

## Build talkkonnect as binary

cd $GOPATH/src/github.com/talkkonnect/talkkonnect
go clean --cache
CC=gccgo go build $1 -o /home/talkkonnect/bin/talkkonnect cmd/talkkonnect/main.go

## Check if talkkonnect building worked and notify user

if [ $? -eq 0 ] && [ -f /home/talkkonnect/bin/talkkonnect 2>/dev/null ]; then

tksize=`du -sh /home/talkkonnect/bin/talkkonnect | cut -c -3`

timestamp=$(date '+%d.%m.%Y at %H:%M:%S')

seconds=$SECONDS

elapsed="$(($seconds / 3600))hrs $((($seconds / 60) % 60))min $(($seconds % 60))sec"

echo "=> Build report from" $timestamp
echo "=> Finished building talkkonnect. It's a success!"
echo "=> Time to build talkkonnect:" $elapsed
echo "=> Your talkkonnect version:" `sed -n 34p /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/version.go \
| cut -c32-38` from `sed -n 35p /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/version.go | cut -c32-42`
echo "=> Your Opus codec version:" `opusenc -V | grep -Po 'libopus\s\K.*' | cut -c1-5`
echo "=> Your Golang version:" `go version | cut -c14-20`
echo "=> talkkonnect binary size:" $tksize
echo "=> talkkonnect binary path: /home/talkkonect/bin/talkkonnect"
echo "=> *** What's left to do before running talkkonnect? ***"
echo "=> 1. Copy a sample xml config from sample-configs to /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect"
echo "=> 2. Rename sample xml config file to talkkonnect.xml"
echo "=> and adjust it for the desired Mumble server settings and program features."
echo "=> or restore your talkkonnect.xml config file from a previous backup."
echo "=> 3. Check your sound system and adjust talkkonnect xml config sound input/output device names"
echo "=> You can then start talkkonnect."
echo "=> Happy t-a-l-K-K-o-n-n-e-c-t-i-n-g!"
echo "=> Keep checking https://github.com/talkkonnect/talkkonnect for the news."
echo "=> ... "

else

timestamp=$(date '+%d.%m.%Y at %H:%M:%S')

echo "=> Failure report from" $timestamp
echo "=> Something went wrong! talkkonnect didn't build. You need to investigate."
echo "=> Check old or current issues at: https://github.com/talkkonnect/talkkonnect/issues"
echo "=> Or report a new issue."
echo "=> ... "

fi

exit 0

) 2>&1 | tee -a /var/log/tk-install.log

