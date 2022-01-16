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
if pgrep -x "$SERVICE" >/dev/null
then
    echo "$SERVICE is running"
    systemctl stop talkkonnect
else
    echo "$SERVICE stopped"
fi

if [[ -f "/root/talkkonnect.xml" ]]
then
	echo "removingroot/talkkonnect.xml"
	rm /root/talkkonnect.xml
fi

if [[ -f "/home/talkkonnect/bin/talkkonnect" ]]
then
	echo "removing /home/talkkonnect/bin/talkkonnect binary"
	rm /home/talkkonnect/bin/talkkonnect
fi

if [[ -f "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml" ]]
then
	echo "copying talkkonnect.xml for safe keeping to /root/talkkonnect.xml"
	cp /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml /root/
fi

rm -rf /home/talkkonnect/gocode/src/github.old
rm -rf /home/talkkonnect/gocode/src/google.golang.org
rm -rf /home/talkkonnect/gocode/src/golang.org
cp -R /home/talkkonnect/gocode/src/github.com /home/talkkonnect/gocode/src/github.old
rm -rf  /home/talkkonnect/gocode/src/github.com
rm -rf  /home/talkkonnect/bin/talkkonnect


## Create the necessary directoy structure under /home/talkkonnect/
mkdir -p /home/talkkonnect/gocode
#mkdir -p /home/talkkonnect/bin
mkdir -p /home/talkkonnect/gocode/src
mkdir -p /home/talkkonnect/gocode/src/github.com


## Set up GOENVIRONMENT
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/home/talkkonnect/gocode
export GOBIN=/home/talkkonnect/bin
export GO111MODULE="auto"

## Get the latest source code of talkkonnect from github.com
echo "getting talkkonnect with go get"
cd $GOPATH 
go get -v github.com/talkkonnect/talkkonnect
cp /root/mumble.pem /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/

## Build talkkonnect as binary
cd $GOPATH/src/github.com/talkkonnect/talkkonnect
go build -o /home/talkkonnect/bin/talkkonnect cmd/talkkonnect/main.go

if [[ -f "/home/talkkonnect/gocode/src/github.old/talkkonnect/talkkonnect/talkkonnect.xml" ]]
then
	echo "copying original talkkonnect.xml back to /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml"
	rm /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml
	cp /home/talkkonnect/gocode/src/github.old/talkkonnect/talkkonnect/talkkonnect.xml  /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml
fi


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


