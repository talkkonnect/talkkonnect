#!/bin/bash

## talkkonnect headless mumble client/gateway with lcd screen and channel control
## Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
##
## This Source Code Form is subject to the terms of the Mozilla Public
## License, v. 2.0. If a copy of the MPL was not distributed with this
## file, You can obtain one at http://mozilla.org/MPL/2.0/.
##
## Software distributed under the License is distributed on an "AS IS" basis,
## WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
## for the specific language governing rights and limitations under the
## License.
##
## The Initial Developer of the Original Code is
## Suvir Kumar <suvir@talkkonnect.com>
## Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
##
## Contributor(s):
##
## Suvir Kumar <suvir@talkkonnect.com>
##
## My Blog is at www.talkkonnect.com
## The source code is hosted at github.com/talkkonnect


## Installation BASH Script for talkkonnect version 2 on fresh install of raspbian bullseye
## Please RUN this Script as root user

## If this script is run after a fresh install of raspbian you man want to update the 2 lines below

apt-get update
apt-get -y dist-upgrade
apt-get install git -y

## Add talkkonnect user to the system
adduser --disabled-password --disabled-login --gecos "" talkkonnect
usermod -a -G cdrom,audio,video,plugdev,users,dialout,dip,input,gpio talkkonnect

## Install the dependencies required for talkkonnect
apt-get -y install libopenal-dev libopus-dev libasound2-dev git ffmpeg mplayer screen pkg-config

## Create the necessary directory structure under /home/talkkonnect/
cd /home/talkkonnect/
mkdir -p /home/talkkonnect/gocode
mkdir -p /home/talkkonnect/bin

## Create the log file
touch /var/log/talkkonnect.log

# Check Latest of GOLANG 64 Bit Version for Raspberry Pi
GOLANG_LATEST_STABLE_VERSION=$(curl -s https://go.dev/VERSION?m=text | grep go)
cputype=`lscpu | grep Architecture | cut -d ":" -f 2 | sed 's/ //g'`
bitsize=`getconf LONG_BIT`

cd /usr/local

if [ $bitsize == '32' ]
then
echo "32 bit processor"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-armv6l.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-armv6l.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-armv6l.tar.gz
else
echo "64 bit processor"
wget -nc https://go.dev/dl/$GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz $GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz
tar -zxvf /usr/local/$GOLANG_LATEST_STABLE_VERSION.linux-arm64.tar.gz
fi

echo export PATH=$PATH:/usr/local/go/bin >>  ~/.bashrc
echo export GOPATH=/home/talkkonnect/gocode >>  ~/.bashrc
echo export GOBIN=/home/talkkonnect/bin >>  ~/.bashrc
echo export GO111MODULE="auto" >>  ~/.bashrc
echo "alias tk='cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/'" >>  ~/.bashrc


## Set up GOENVIRONMENT
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/talkkonnect/gocode
export GOBIN=/home/talkkonnect/bin
export GO111MODULE="auto"

## Get the latest source code of talkkonnect from github.com
echo "installing talkkonnect with traditional method avoiding go get cause its changed in golang 1.22 "
cd $GOPATH
mkdir -p /home/talkkonnect/gocode/src/github.com/talkkonnect
cd /home/talkkonnect/gocode/src/github.com/talkkonnect
git clone https://github.com/talkkonnect/talkkonnect
cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect
go mod init
go mod tidy

## Build talkkonnect as binary
cd $GOPATH/src/github.com/talkkonnect/talkkonnect
go build -o /home/talkkonnect/bin/talkkonnect cmd/talkkonnect/main.go

## Notify User
echo "=> Finished building TalKKonnect"
echo "=> talkkonnect binary is in /home/talkkonect/bin"
echo "=> Now enter Mumble server connectivity details"
echo "talkkonnect.xml from $GOPATH/src/github.com/talkkonnect/talkkonnect"
echo "and configure talkkonnect features. Happy talkkonnecting!!"

exit


