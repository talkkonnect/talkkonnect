# Configuring and Running talKKonnect

##### Sample XML files can be found from sample-configs folder: #####

https://github.com/talkkonnect/talkkonnect/tree/main/sample-configs

See the file talkkonnect-version2-usb-gpio-example.xml and talkkonnect.tkv1pcb as an example with all the latest XLS tags.

##### Start  talKKonnect binary #####

````
cd /home/talkkonnect/bin
./talkkonnect 
````

##### You can start talKKonnect automatically on Raspberry Pi start up with “screen” program help. Add this line to /etc/rc.local file. before “exit 0”: #####

` screen -dmS talkkonnect-radio /home/talkkonnect/bin/talkkonnect & `

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
````

##### I2C OLED Screen Installation #####
For those of you who wish to use a 0.96 or 1.3 inch OLED screen follow the instructions below (logged in as root)

[enabling i2c](https://www.raspberrypi-spy.co.uk/2014/11/enabling-the-i2c-interface-on-the-raspberry-pi/) read and Follow Step 1 - Enable I2C Interface.

For detecting the address of your screen install the tool below

` apt-get install -y i2c-tools `

Then using i2cdetect to detect your screen following the instructions on the same page under the section Testing Hardware (Optional)

Once you get the address note that it will be in HEX you will have to convert this address to decimal to put in the talkkonnect.xml file
under the xml tag  <oleddefaulti2caddress>60</oleddefaulti2caddress>

In the example above I got the address 3c from i2c tools and converted that to decimal value 60. For My Setup the i2c-detect tool gave
me the hex address of 3c so converted to decimal it is the value 60.


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

## The talkkonnect.xml configuration File tags and their meaning

### The Accounts Section
* The account section can have multiple accounts, talkkonnect will look for the first account with the xml tag default = "true" and attempt to connect to that server
* When talkkonnected is connected to a server you can cycle through accounts in which enabled = "true" by pressing CTRL-N, talkkonnect will connect to the next enabled server in the list
* Talkkonnect will not attempt to connect to a server that has the account tag set default = "false"
* The tag account name is just used to identify the server for logging purposes
* The serverandport tag is for the server FQDN or IP address followed by a (colon) and the port of mumble is running on for that particlar server. Our community mumble server setting is currently mumble.talkkonnect.com:64738
* The username tag is used for identifying yourself on the mumble server and for authentication
* The password tag is used if the mumble server requires password authentication
* The insecure tag should be set as true if the server you are connecting to does not require a certificate
* The certificate tag should contain the full path to your previously generated certificate which is usually a file with the extension of pem
* The channel tag should only be populated want to connect to a specific channel other than the root channel on startup
* The ident tag is used to identify the name of the person using talkkonnect for help request sent when using the panic function
* The listentochannel has xml child tags of  channel in which you can setup channels you want to listen to using the channellisten feature of mumble
* The tokens list for each account for autorization to token protected channels
* The voicetargets IDs and their corresponding users and channels

### The Global Section of talkkonnect.xml (Software & Hardware)

### Software Settings Section

* The outputdevice tag should be set as the default audio output device that represents your audio output device when you run alsamixer. Examples are Speaker or Headphone etc. (If you are not sure whats going on set all 4 xml tags starting with outputdevices to what is the default output sound device on alsa.)
* The logfilenameandpath tag should contain the full path to a writable file that is created prior to running talkkonenct for logging purposes
* Should you not require logging to screen set the logging tag to screen. Any other value will result logs to be shown on the screen and in the log file (note that if logging is not set to screen the logs will no longer be colorized) The options available for logging are screen, screenwithlineno, screenandfilewithlineno
* Cancellable Stream is used so that if you are streaming some audio via talKKonnect another user in the channel can stop your streaming by pressing PTT.
* Simplexwithmute is used to set simplex mode (mute speaker when transmitting) or full duplex mode (not mute speaker with transmitting)
* The for the loglevel tags there are the following options (trace, debug, info, warning, error, alert) default to info
* The tag cancellablestream determines whether whilst you are streaming audio to a channel if you allow your stream to be stopped by an incomming transmission.
* The streamonstart tag lets talkkonnect know that you want to start streaming after talkkonnect starts
* The streamafterstart tag lets talkkonnect know how many seconds to wait until talkkonnect starts streaming after connecting to the server
* The streamsendmessage tag lets talkkonnect know whether to send a message to the talkkonnect channel users that streaming is starting
* The txonstart tag lets talkkonnect know if you want talkkonnect to start trasmitting after startup (input from microphone)
* The txafterstart tag lets talkkonnect know how many seconds to wait until talkkonnect starts transmitting after connecting to the server
* The repeattxtimes and repeattxdelay are the looped amount of times you want talkkonnect to start transmitting with delay (used for testing)
* The simplexwithmute function when set to true will mute talkkonnect speaker when transmitting to prevent feedback and behave like a radio transceiver
* The txcounter tag is enabled just for debugging to count the amount of times the user pressed ptt since startup (used for debugging/testing)
* The nextserver index should be set to 0 as default, this is used to inform talKKonnect which server to connect to in the xml config if there are more than
  one server marked true. The count starts at zero from the first server with the default tag marked as true.
* The txlockout feature when set to true prevents talkkonnect from transmitting if there is someone speaking on the channel
* The listentochannelonstart when set to true will enable talkkonnect to listen to channels defined in the listentochannel child tag channel on start as set
  in the accounts section of the configuration

#### Autoprovisioning Section
Autoprovisioning is provided so that you can remotely provision a talkkonnect machine via http protocol from a web server
* The autoprovisioning tag when set to true or false turns on and off the autoprovisioning function respectively
* The tkid tag is used to set the autoprovisioning filename (xxxx.xml) that talkkonnect will request from the autoprovisioing web server
* The URL tag is used to define the url of the autoprovisioning webserver that hosts the configuration XML file
* The savefileandpath tag are used to define the name and where the http fetched xml file will be stored locally. This is usually /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml

#### Beacon Section
The beacon function was created to emulate a radio repeater beacon that will play certain wav files at defined periods to notify all users on a particular channel that the repeater is online and functioning. You can now play the wav file into the mumble stream or you can also choose to play the announcement on the Speaker (RF Side if used as a repeater controller)
* The beacontimersecs is the interval time in seconds between the repleated messages
* The beaconfileandpath is the tag which defines the file and path to a wav file that to be played at regular intervals
* The localplay tag is set to true will play the sound on to the speaker (or RF side)
* The localvolume tag can be set from 1-100 for setting up the volume the file playback into speaker
* The GPIO enabled features can be used to key up a transmitter when the message is on localplay (used for repeaters)
* The gpioname is the name of the gpio pin defined that will go high/low when the message is played
* The playintostream tag enables you to play the wav file into direction of the mumble stream
* The beaconvolumeintostream can be set from 1-100 for setting up the volume the file playback into mumble stream

#### The TTS Section
This section was created for users that want an audible response to events that happen (Usually Users without LCD Screen)
* You can disable the whole section TTS functionality by the tag tts enabled = false
* You can choose to enable only certain events you are interested in by setting tag tts enabled = true and selecting the tag you want for your particular use case
* You can also set the blocking tag to determine the nature of the tts playback
* The volumelevel can be set from 1-100
* the language default is en

#### The SMTP Section
Talkkonnect currently can only connect to gmail's SMTP for sending emails
* Define your gmail username and password along with the receiver of the email message in their respective tags
* Define the subject and fixed message body of the email in their respective tags
* Should you want to send the GPS timestamp in the email set the gpsdatetime tag to true (You have to have a USB GPS Dongle Connected and Configured for this to work)
* Should you want to send by email your current GPS position in LAT and LONG coordinates you can enable this tag
* If you want to include the url with your pinned location on google maps enable the googlemapurl tag

#### The Sounds Section
Basically in talkkonnect there are 2 types of sound events one is the system event type and the other is the user command input type.
The system event type sounds are generated either by mumble or talkkonnect when events occur in the server or channel. The user command input
type even sound is generated when the talkkonnect user commands talkkonnect to do something either directly or via remote control.
* Each sound item can be enabled/disabled and the corresponding playback volume can be also be set individually
* Each event such as when a person joins a channel, leaves a channel or sends a message into the channel can be configured seperately.
* The filenameandpath tag should contain the the full path and filename of the WAV file you wish to play for each event
* The event tag is used to play an audible alert when there are changes of other users statuses
* The alert tag is used to play an WAV file into the stream to the receiving party upon a user generated panic request
* The rogerbeep tag is used to define the WAV file to play at the end of every transmission
* The tag name stream, This function is very powerful and can be used to define a local file or network stream that will be played into the mumble channel upon pressing the F11 key. Very useful for debugging.
* The blocking tag defines if the user wants talkkonnect to wait for playing the file before continuing or just play the file and contiune to the next event

#### The Repeater Tone Section
The repeater tone is to define talkkonnect to play a sine wave tone into the speaker or radio input for a pre-defined duration in seconds
of pre-defined frequency possibly to open a repeater. This feature was requested by european hams. I thought that if there was a sine generator

#### The TXTIMEOUT section
The txtimeout tag is used to limit the length of a single transmission in seconds. This tag is useful when used as a repeater between RF and mumble.
Also used to prevent the stuck key and talkkonect stuck in transmit mode.

#### The API Section (Remote Control over HTTP)
API section enables the user to granually control which remote control functions are available over http within the local network. The tag apilisten port defines the port that talkkonnect should listen and respond to remote control http requests

````
Listing API available http://{your-talkkonnect-ipaddress}:8080/?command=listapi
Channel Up           	http://{your-talkkonnect-ipaddress}:8080/?command=channelup
Channel Down 		http://{your-talkkonnect-ipaddress}:8080/?command=channeldown
Mute/UnMute Toggle   	http://{your-talkkonnect-ipaddress}:8080/?command=mute-toggle
Mute Speaker         	http://{your-talkkonnect-ipaddress}:8080/?command=mute
Unmute Speaker       	http://{your-talkkonnect-ipaddress}:8080/?command=unmute 
Volume UP            	http://{your-talkkonnect-ipaddress}:8080/?command=volumeup
Volume Down          	http://{your-talkkonnect-ipaddress}:8080/?command=volumedown
Start Transmitting   	http://{your-talkkonnect-ipaddress}:8080/?command=starttransmitting
Stop Transmitting    	http://{your-talkkonnect-ipaddress}:8080/?command=stoptransmitting
Play/Stop Stream     	http://{your-talkkonnect-ipaddress}:8080/?command=stream-toggle
Request GPS Position 	http://{your-talkkonnect-ipaddress}:8080/?command=gpsposition
Send Email           	http://{your-talkkonnect-ipaddress}:8080/?command=sendemail
Previous Server      	http://{your-talkkonnect-ipaddress}:8080/?command=connpreviousserver
Next Server          	http://{your-talkkonnect-ipaddress}:8080/?command=connnextserver
TTS Annoncement      	http://{your-talkkonnect-ipaddress}:8080/?command=ttsannouncement
Voice Target 0       	http://{your-talkkonnect-ipaddress}:8080/?command=setvoicetarget&id=0
Voice Target 1       	http://{your-talkkonnect-ipaddress}:8080/?command=setvoicetarget&id=1
ChannelListener Start	http://{your-talkkonnect-ipaddress}:8080/?command=listeningstart
ChannelListener Stop 	http://{your-talkkonnect-ipaddress}:8080/?command=listeningstop
````

#### The PrintVariables Section
* This function is useful for debugging the values read from each section of the config xml file. You can control which section is shown. This command is tied to the CTRL-X key

#### The MQTT Section
* Talkkonnect can be remotely controlled by an public or local MQTT Server
* This eliminates the problem of controlling those talkkonnect devices that are in NATTED networks all over the internet
* You can subscribe to the mqtt server topic of your choice
* With MQTT you can remote control talkkonnect as well as Relays to control external devices

Example Below is the talkkonnect.xml with working MQTT Configuration, Let's For Example Say we want to increase the microphone volume of cameraman1
First Set Up talkkonect.xml as below make sure to change your mqtt server details for the parameters mqttbroker, mqttpassword, mqttuser,mqttid.

````
        <mqtt enabled="true">
          <settings>
          <mqttsubtopic>event/cameraman1/</mqttsubtopic>
          <mqttpubtopic>response/cameraman1</mqttpubtopic>
          <mqttbroker>tcp://mqtt.yourserver.com:1883</mqttbroker>
          <mqttpassword>yourpassword</mqttpassword>
          <mqttuser>camera1</mqttuser>
          <mqttid>camera</mqttid>
          <cleansess>false</cleansess>
          <qos>0</qos>
          <num>1</num>
          <payload/>
          <action>sub</action>
          <store/>
          <retained/>
            <attentionblinktimes>20</attentionblinktimes>
            <attentionblinkmsecs>300</attentionblinkmsecs>
            <pubpayload>
              <mqtt item="0" payload="channelup" enabled="true"/>
              <mqtt item="1" payload="channeldown" enabled="true"/>
            </pubpayload>
          </settings>
          <commands>
            <command action="displaymenu" message="Display Menu" enabled="true"/>
            <command action="channelup" message="Channel Up" enabled="true"/>
            <command action="channeldown" message="Channel Down" enabled="true"/>
            <command action="muteunmute" message="Mute-Toggle" enabled="true"/>
            <command action="currentrxvolume" message="Current Volume" enabled="true"/>
            <command action="volumerxup" message="Volume Up" enabled="true"/>
            <command action="volumerxdown" message="Volume Down" enabled="true"/>
            <command action="currenttxvolume" message="Current Volume" enabled="true"/>
            <command action="volumetxup" message="Volume Up" enabled="true"/>
            <command action="volumetxdown" message="Volume Down" enabled="true"/>
            <command action="listserverchannels" message="List Channels" enabled="true"/>
            <command action="starttransmitting" message="Start Transmitting" enabled="true"/>
            <command action="stoptransmitting" message="Stop Transmitting" enabled="true"/>
            <command action="listonlineusers" message="List Users" enabled="true"/>
            <command action="playback" message="Playback" enabled="true"/>
            <command action="gpsposition" message="GPS Position" enabled="true"/>
            <command action="sendemail" message="Send Email" enabled="true"/>
            <command action="previousserver" message="Previous Server" enabled="true"/>
            <command action="connnextserver" message="Next Server" enabled="true"/>
            <command action="clearscreen" message="Clear Screen" enabled="true"/>
            <command action="pingservers" message="Ping Servers" enabled="true"/>
            <command action="panicsimulation" message="Panic Simulation" enabled="true"/>
            <command action="repeattxloop" message="Repeat TX Loop" enabled="true"/>
            <command action="scanchannels" message="Scan Channels" enabled="true"/>
            <command action="thanks" message="Thanks" enabled="true"/>
            <command action="showuptime" message="Show UpTime" enabled="true"/>
            <command action="dumpxmlconfig" message="Dump XML Config" enabled="true"/>
            <command action="ttsannouncement" message="TTS Announcement" enabled="true"/>
            <command action="voicetargetset" message="Set Voice Target" enabled="true"/>
            <command action="attention" message="Attention LED" enabled="true"/>
            <command action="relay" message="RelayControl" enabled="true"/>
          </commands>
        </mqtt>

````
If we connect to the mqtt server and publish to the topic event/cameraman1 and send the message(payload) as volumetxup we will notic that the volume of the cameraman1 is now louder
when he/she speaks.

Below are Valid Commands for MQTT
* channelup
* channeldown
* mute-toggle
* mute
* unmute
* currentrxvolume
* volumerxup
* volumerxdown
* currenttxvolume
* volumetxup
* volumetxdown
* starttransmitting
* stoptransmitting
* stream-toggle
* playback
* gpsposition
* connpreviousserver
* connnextserver
* panicsimulation
* scanchannels
* attentionled:on
* attentionled:off
* attentionled:blink
* relay1:on
* relay1:off
* relay1:pulse
* voicetargetset
* listeningstart
* listeningstop

For Example on the topic thailand/bangkok/company/talkkonnect/attentionled:on will turn on the LED to get the attentionled
of a user.

Another Example on the topic thailand/bangkok/company/talkkonnect/relay1:pulse will simulate a push button for example to
open the door for a an access control system

For the above example to work you will have to specify the gpio pin in the <lights> section of the xml file
<attentionledpin></attentionledpin>
<relay1pin></relay1pin>

### TTSMessages Section
* This function is useful for when you want to talkkonnect to act as annunciator. This function uses google translate
  to generate speech from text.
* The ttsmessages enabled tag disables or enables this functionality in talkkonnect
* ttslanguage can be set to en for english for example (see google translate for other language codes)
* ttsmessagesfromflag if this is set to true the message will say "message from" in the accouncement, however for announcement purposes
  we recommend that you set it to false if you dont want the word "message from" spoken in the announcement
* ttstone is the tag that needs an absolute path to a wav file to play before the announcement
* blocking is set so that sounds are played in order waiting for each sound to finish playing
* ttssounddirectory is set to audio this is where the wav files from the tts engine will be created
* localplay is set to true if you want to play the file locally out of the talkkonnect speaker or local amplifiers
* playintostream is set to true if you want to play the announcement into the stream for all users in the channel to hear the announcement
* speakvolumeintostream will set the volume from 0 to 100 for the playback of the tts sound into the stream for others on the same channel to hear
* playvolumeintostream  will set the volume from 0 to 100 for the playback of the wav file sound into the stream for others on the same channel to hear
* gpio name=transmit if enabled the GPIO will be set to high while there is an announcement playing for the purposes of driving a relay to possibly
  turn on an external device like an audio amplifier or flashing light for calling attention when the announcement is playing.
* the predelay and postdelay can be set in seconds to provide a slight delay before and after the playing the announcement

### Ignore User Section
* This function is used in case you want not to receive the audio of any user(s) for example you are transmitting on another
  device in the same room and do not want to hear yourself
* Notice the ignoreuserregex accepts a regular expression of the user to ignore and not play the audio for that user upon reception

## Hardware Section
* The tag targetboard has 2 option (1) pc and (2)rpi. pc mode is used when talkkonnect is running on a pc or server that does not have GPIOs and is not interfaced to buttons and a LCD screen.
* To run on raspberry pi or other compatible single board computers set the targetboard to rpi this will enable the GPIO outputs/inputs.
* If you are using the respeaker hat with a ledstrip

````
     <ledstripenabled>true</ledstripenabled>
     <voiceactivitytimermsecs>200</voiceactivitytimermsecs>
````

### The IO Section
* The IO Section has been completely rewritten to make it more flexible and easier to add functionality to talkkonnect in the future.
* The gpioexpander enabled = true indicates that for there is a MCP23017D chip connected to the i2c bus address as specified in the i2cbus tag with
  the address configuration specifed in the mcpdevice this number can be from 0 to 7 supporting a total of 8 chips on the same i2c bus. So far the
  expander chip is used for gpio in the output direction. The enabled or disabled tag per chip id enables or disables that particular chip.
* There is also now support for max7219 8 x 7-Segment Driver connected to the SPI bus for the purposes of showing the channel id. Here you will need to
  set the cascaded to 1 and the spibus address aldo the spidevice and brightness from 0 to 7. The enabled tag enables and disables this display.
* The pins section with the pins tag maps the direction of the gpio pin which can be input or output. The device can be led or relay or lcd or radiomodule
  or pushbutton or toggleswitch. The name tag maps to pre-defined functions in talkkonnect, the pinno tag maps to the gpio number and if external expander
  chip is not used set the chipid to 0. The inverted signal inverts the logic voltage of the output signal from active high to active low and vice versa.
  For normal cases the inverted would be set to false. As usual the enabled tag disables or enables each gpio line.
* The rotary encoder function can now be enabled and disabled, to control the mumble channel up down, local volume, or radio channel of the sa818 rf module,
  depending on what functionality you want from the rotary encoder you can set the order and enable and disable them. You can toggle between rotary encoder
  functions by pressing the rotary encoder button. If the rotary encoder is responding too quickly or too slowly adjust the leadingmsecs pulsemsecs and
  trailing msecs tags.
* To use the rotary encoder you will also have to define in the io section pins for rotarya,rotaryb and button, since they are inputs the chipid has to
  be 0 and the device has to be rotaryecoder.
* The memory channel features enable you to define a GPIO name to a favourite channel. You can have up to 4 buttons for memory channels.

#### The Volume Button Step Section
* This setting allows you to define the increase or decrease step of volume adjustment. This setting will depend on your preference and sound card used.

#### The Heartbeat Section (OUTPUT)
* The heartbeat tag defines the GPIO pin that will toggle as per the defined values to show that talkkonnect is alive and operational
* Note that this heartbeat can uses the same GPIO PIN and voiceactivitypin so that one LED can have dual function
* Note Disable heartbeat or do not use the same pin as voiceactivity LED if you connect talKKonnect to a transceiver

#### The Comment Section
* This function allows the user to set 2 possible messages like for example away messages depending on the state of a toggle switch
* When another party using talkkonenct presses F10 they can see the username along with the defined message (depending on the position of the switch on/off) in square brackets
* The commentbuttonpin tag defines the GPIO pin that the toggle switch is connected to

#### The LCD Section (For HD44780 20x4 LCD SCREEN)
* At this moment talkkonnect supports the easily available 4 lines 20 characters HD44780 LCD Module.
* To disable this screen option you can set enabled = "false"
* Parallel and i2c interfacing to the HD44780 LCD Module are both supported and can be configured in this section
* Valid interfacetype tag are either parallel or i2c
* The i2c address can be obtained from running the i2cdetect -y 1 command. Convert the address displayed in HEX to Decimal and fill into the lcdi2caddress tag
* The backlight function and time is also available to turn off the LCD's backlight in case of inactivity on the channel for the defined timeout period in seconds
* The rs, e, d4, d5, d6, d7 pins are the GPIO pins that connect to the HD44780 display in parallel mode
* NOTE! You cannot use the pins 2,3 on raspberry pi for anything else other than I2C mode if you want to connect an I2C display

#### The OLED Section (For 0.96 and 1.3 Inch I2C Interface OLED SCREEN)
* At this moment talkkonnect also supports the easily available 0.96 and 1.3 Inch I2C OLED Screen.
* To disable this screen option you can set enabled = "false"
* i2c interfacing is the only option that should be specified now spi has not been developed
* The i2c address can be obtained from running the i2cdetect -y 1 command. Convert the address displayed in HEX to Decimal and fill into the lcdi2caddress tag and mostly the i2c bus is 1.
* There is no backlight function for oled screens yet
* Your will have to specify the rows and columns your screen supports (for my screen i used 8 rows and 21 columns)
* The OLED display is display width and height for my screen was 130 by 64
* Another important settings is the oledstartcolumn setting for 0.96 screens set to 0 and for 1.3 inch screens set to 1. This will clear any garbage you see on the edge of the screen.
* NOTE! You cannot use the pins 2,3 on raspberry pi for anything else other than I2C mode if you want to connect an I2C display

#### The GPS Section
* Talkkonnect supports a ublox 6 USB module to provide GPS tracking on Panic mode activation
* Set the enabled tag to false if you do not have a USB dongle connected
* Define the port which the GPS is detected as in linux usually /dev/ttyACM0
* Define all other serial port settings such as serial baud, even/odd/none parity, also stop and databits.

#### The Traccar Section
* You can disable and enable GPS tracking in traccar platform by the enabled tag
* if you set track to true data will be sent to traccar depeding on the protocol name you define and protocol connection details you configured.

#### The PanicFunction Section
* The panic function can be enabled or disabled and is used to request for help
* Filenameandpath tag is used to define the WAV file that will be played into a stream if the panic button is pressed
* The volume tag defines the playback volume of the wav file into the stream
* The sendident will send the contents of the ident tag defined in the account section. This is used in case you want for example your Name or alternate ID sent in the panic message.
* The panicmessage tag defines the text message that will be sent to the parent channel and all child channels if recursivemessage is set as true when the panic button is pressed
* The sendgpslocation tag enables the sending of the gps coordinates of the talkkonnect requesting help as a text message
* The txlock enabled tag will lock up talkkonnect in transmit mode for the defined txlocktimeoutsecs after the button is pressed so the requester can talk without having to press ptt button

#### The USB Keyboard Section
* You can enable a wired/wireless USB numpad here for voice targeting and other direct commands to a headless talkkonnect

#### The KeyboardCommands Section
* With talkkonnect you have a choice of using a USB keyboard/USB Numpad to control talkkonnect for example press '1' for channelup and '2' for channeldown
* Using the usb keyboard you do not have to be logged in to the terminal its just plug and play
* You can define the command associated with each tty key or each key on your USB Numpad
* TTY Means the keyboard that you use when you ssh into the box or use talkkonnect in a terminal
* in the ttykeyboard you can set the scan id for each key in the terminal session
* in the usbkeyboard you can set the scan id for each key in the physical keyboard
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
* voicetargetset
* mqttpubpayloadset
* repeatertoneplay
* listentochannelstart
* listentochannelstop

#### Radio Section
* The radio section is to support a SA818 type RF Transciver connected to talkkonnect via serial so that talkkonnect
  can control the RF parameters of the module (this is a use case for repeater)
* the enabled tag can tell talkkonnect whether there is a SA818 device connected to talkkonnect
* you can set the serial parameters to the serial port, baud, stopbits, databits in the serial tags
* for the purposes for channel control you have t odevice channel ID for example 01,02,03,04 and so on
* you will need to set the
* bandwidth to 0 or 1
* rxfrequency Mhz with 3 4 decimals
* txfrequency Mhz with 3 4 decimals
* squelch 0 - 8
* ctsstone
* dcstone
* predeemph
* highpass
* lowpass
* volume 0 - 8
   