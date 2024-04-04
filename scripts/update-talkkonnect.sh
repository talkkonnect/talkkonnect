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


## Installation BASH Script for talkkonnect on fresh install of raspbian
## Please RUN this Script as root user

SERVICE="talkkonnect"
BACKUPXML=talkkonnect-$(date +"%Y%m%d-%H%M%S").xml
BACKUPCERT=mumble-$(date +"%Y%m%d-%H%M%S").pem

if pgrep -x "$SERVICE" >/dev/null
then
    echo "$SERVICE is running"
    systemctl stop talkkonnect
else
    echo "$SERVICE stopped"
fi

if [[ -f "/home/talkkonnect/bin/talkkonnect" ]]
then
	echo "removing /home/talkkonnect/bin/talkkonnect binary"
	rm /home/talkkonnect/bin/talkkonnect
fi

if [[ -f "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml" ]]
then
	echo "copying talkkonnect.xml for safe keeping to /root/"$BACKUPXML
	cp /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml /root/$BACKUPXML
fi

if [[ -f "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/mumble.pem" ]]
then
	echo "copying mumble.pem for safe keeping to /root/"$BACKUPXML
	cp /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml /root/$BACKUPCERT
fi

rm -rf /home/talkkonnect/gocode/src/github.old
rm -rf /home/talkkonnect/gocode/src/google.golang.org
rm -rf /home/talkkonnect/gocode/src/golang.org
rm -rf  /home/talkkonnect/gocode/src/github.com
rm -rf  /home/talkkonnect/bin/talkkonnect


## Check Latest of GOLANG 64 Bit Version for Raspberry Pi
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


## Create the necessary directoy structure under /home/talkkonnect/
mkdir -p /home/talkkonnect/gocode
mkdir -p /home/talkkonnect/gocode/src
mkdir -p /home/talkkonnect/gocode/src/github.com


## Set up GOENVIRONMENT
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/talkkonnect/gocode
export GOBIN=/home/talkkonnect/bin

## Get the latest source code of talkkonnect from github.com
echo "installing talkkonnect with go installing"
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

echo copying $BACKUPXML
cp /root/$BACKUPCERT /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/mumble.pem
cp /root/$BACKUPXML /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml

if pgrep -x "$SERVICE" >/dev/null
then
    echo "$SERVICE is running I will stop it please start talkkonnect manually"
    systemctl stop talkkonnect
else
    echo "$SERVICE is stopped now restarting talkkonnect"
    systemctl start talkkonnect
fi

## Notify User
echo "=> Finished Updating TalKKonnect"
echo "=> Updated talkkonnect binary is in /home/talkkonect/bin"
echo "copied old talkkonnect.xml file and replaced in /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/"
echo "Happy talkkonnecting!!"

exit
