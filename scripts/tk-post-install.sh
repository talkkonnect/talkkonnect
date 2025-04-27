#!/bin/bash

# talkkonnect post installation script to copy a default xml config and propmpt for Mumble server info.
# Start in the "pc" mode for the initial connectivity test, before enabling peripherals like displays, leds, buttons.
# By Zoran D

#Check if talkkonect program code dir exist

if ! [ -z $(ls -A /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect 2>/dev/null) ]; then
echo "Old talkkonnect code is detected in the system."
echo "Working xml config if it exists, will be backed up and default xml config downloaded."

if [ -f /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml 2>/dev/null ]; then
echo "Found an old working talkkonnect.xml config. Backing up to talkkonnect-$timestamp.xml"
timestamp=$(date '+%H-%M-%S-%d-%m-%Y')
mv /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect-$timestamp.xml
fi

cd /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect

#Download new xml config
echo "A default talkkonnect xml configuration will be copied to the program directory."
wget https://raw.githubusercontent.com/talkkonnect/talkkonnect/refs/heads/main/sample-configs/talkkonnect-version2-usb-gpio-example.xml

mv talkkonnect-version2-usb-gpio-example.xml talkkonnect.xml

echo "You will be asked to enable one default Mumble server connection."
echo "If you need to connect to more Mumble servers, please edit your talkkonnect xml configuration manually."

echo "Choose targetboard Raspberry Pi (1) or PC (2)"
echo "(1) Raspberry Pi"
echo "(2) PC"
read targetboard;
case $targetboard in
1) sed -i -e 's/<hardware targetboard.*/<hardware targetboard="rpi">/g' talkkonnect.xml ;;
2) sed -i -e 's/<hardware targetboard.*/<hardware targetboard="pc">/g' talkkonnect.xml ;;
*) echo "This targetboard option is not supported. Please choose a supported option.";;
esac

echo "Choose talkkonnect logging mode (1), (2) or (3)"
echo "(1) Screen"
echo "(2) Screen with line number"
echo "(3) Screen and file with line number"
read tklogging;
case $tklogging in
1) sed -i -e 's/<logging>.*/<logging>screen<\/logging>/g' talkkonnect.xml ;;
2) sed -i -e 's/<logging>.*/<logging>screenandwithlineno<\/logging>/g' talkkonnect.xml ;;
3) sed -i -e 's/<logging>.*/<logging>screenandfilewithlineno<\/logging>/g' talkkonnect.xml ;;
*) echo "This logging option is not supported. Please choose a supported mode." ;;
esac

read -p "Enter Mumble server description (example: talkkonnect-community-server): " mumbleserverdesc
if [[ $mumbleserverdesc == "" ]];
then
sed -i -e 's/<account name=.*/<account name="talkkonnect-community-server" default="true">/g' talkkonnect.xml;
else
desc=$mumbleserverdesc
nospace=${desc// /-}
sed -i -e 's/<account name=.*/<account name="'$nospace'" default="true">/g'  talkkonnect.xml;
fi

read -p "Enter Mumble server name and port (example: mumble.talkkonnect.com:64738): " mumbleserverandport
if [[ $mumbleserverandport == "" ]];
then
sed -i -e 's/<serverandport>.*/<serverandport>mumble.talkkonnect.com:64738<\/serverandport>/g' talkkonnect.xml;
else
sed -i -e 's/<serverandport>.*/<serverandport>'$mumbleserverandport'<\/serverandport>/g' talkkonnect.xml;
fi

read -p "Enter Mumble user name: " mumbleusername
sed -i -e 's/<username>.*/<username>'$mumbleusername'<\/username>/g' talkkonnect.xml;

read -p "Enter Mumble user password: " mumbleuserpassword
sed -i -e 's/<password>.*/<password>'$mumbleuserpassword'<\/password>/g' talkkonnect.xml;

read -p "Enter Mumble channel (none for /): " mumblechannel
sed -i -e 's/<channel>HAM-CB<\/channel>/<channel>'$mumblechannel'<\/channel>/g' talkkonnect.xml;

read -p "User identity: " mumbleuserident
sed -i -e 's/<ident>.*/<ident>'$mumbleuserident'<\/ident>/g' talkkonnect.xml;

echo "talkkonnect xml config is ready! You can now launch your talkkonnect instance!"

#start pulseuadio in the background and tk in foreground.
#screen sudo pulseaudio --system &&
#screen -dmS tk /home/talkkonnect/bin/talkkonnect

# test run once
killall -9 talkkonnect
/home/talkkonnect/bin/talkkonnect

else

echo "talkkonnect code was not found in the system. Nothing to do. Exiting."

fi

exit 0
