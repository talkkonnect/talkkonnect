# Configurability, Features, and Additional Optional Hardware and Precautions for talKKonnect

## Configuration
* XML Granular configurability covering many flexible uses cases.
* Centralized autoprovisioning for multiple talkkonnect devices from a centralized http provisioning server using XML delivery via http
* Configurable choice of GPIO pins for each function on a variety of commercially available SBC boards

## GPIO and Hardware Support
* GPIO WIth Optional GPIO Expander (up to 16 x 8 = 128 GPIO using The MCP23017 Chip over I2c)
* Rotary Encoder Support for Channel Up/Down, Volume Up/Down, SA818 Radio Module Frequency Change, Voice Target Change
* LCD/OLED Screen (Parallel and I2c Interface) showing relevant real time information such as *server info, current channel, who is currently talking, time, etc.*
* Connecting to low cost USB GPS dongles (for instance “u-blox”) for GPS tracking, Panic Alerts integration with traccar GPS tracking software.
* Seven Segment Support For Showing Channel like CB Rado using MAX7219 Chip with Seven Segment Displays
* Panic button, when pressed, talKKonnect will send an alert message with GPS coordinates, followed by an email indication current location in google maps.

## Remote Control Features
Local or Remote Control via
* GPIO Pins and Buttons
* Locally attached USB keyboard
* SSH, Console terminal
* HTTP API and/or MQTT with Granular Configurable remote control commands, LED Control, Button Control, Relay Control

## Using talkkonnect As a Radio Gateway Interface
* Communications bridge to interface external (otherwise not compatible) radio systems both over the air and over IP networks.
* Interface to portable or base station radios (Beefing portable radios or UART radio boards).
* Tone Based Repeater Opening Function with the ability to specify the tone frequency and duration in configuration.

## Extra Multimedia Features (IP-Speaker)
* Full Duplex Support (No Audio Stuttering on Multiple people talking over each other)
* Sound Files can be tied with events/actions in config (Support both blocking and non-blocking modes)
* Streaming Audio into the channel from locally stored media file or from internet stream by local or API Call
* Announciator Support using Google TTS with Multi Language Support
* Analog Relay Control By Listening Channel for PA Announcements

## Mumble And other Features
* Shout and Whisper Support
* Channel Token Support
* Configurable Voice targeting via USB Numpad keyboard, TTY Keyboard, API, MQTT (Shouting and Whispering)
* Configurable functions such as mute,unmute,channel,txptt available on USB Keypad
* Listening on Multiple Channels Support
* Multiple Server Configurations with channel control, channel scanning and server hopping
* Many Other features as per suggested or requested by the community too many to mention here



You can find the typical circuit diagram in PDF format for raspberry pi 2,3 Series, Zero 2W [here](https://github.com/talkkonnect/talkkonnect/blob/main/circuit-diagram/Schematic_talkkonnect_2023-12-13.pdf)

You can use an external microphone with push buttons (up/down) or rotary encoder for Channel navigation for a mobile transceiver like experience. Using a rotary encoder makes talkkonnect very usable with minimal user interface.

Currently talKKonnect works with 4×20 Hitachi [HD44780](https://www.sparkfun.com/datasheets/LCD/HD44780.pdf) LCD screen in parallel mode.  Other screens like 0.96" and 1.3" [OLED](https://learn.adafruit.com/adafruit-oled-displays-for-raspberry-pi) with I2C interface is also currently supported. Currently for SPI only seven segment displays are supported using MAX7219 chip.

Low cost Class-D audio amplifiers like [PAM8403](https://www.instructables.com/id/PAM8403-6W-STEREO-AMPLIFIER-TUTORIAL/) or similar “D” class amplifiers, are recommended for talKKonnect builds.

A good shileded cable for microphone is recommended to keep the noise picked up to a minimum for mics when using analog soundcards.
Instead of the onboard sound card or USB Sound Card, you can also use a ReSpeaker compatiable HAT, or a ReSpeaker USB Sound Card with built in Amplifier and achieve great audio quality results in a compact form factor.

### LED Statuses
You can connect LED indicators that can be build on the front panel of your build to show the following statuses:
* Connected to a server and is currently online (online LED)
* There are other participants logged into the same channel (participants LED)
* Currently in transmitting mode  (transmit LED)
* Currently receiving an audio stream (someone is talking on the channel) (voiceactivity LED)
* Heart Beat to indicate that talKKonnect is running (heartbeat LED)
* Currently in Voicetarget mode or Speaking in Normal mode for all clients on the channel to hear (voicetarget LED)
* Flashing Attention LED to get the attention of a talkkonnect operator in a noisy environment.

### Software Configurable Features ###
* Colorized Logs are shown on the debugging terminal for events as they happen in real time. Logging with line number, logging to file or screen or both.
* Playing of configurable *alert sounds* as different events happen, such as a different sound when someone "joins" the channel and another sound for someone "leaving" the channel.
* *TTS prompts* to announce different events for those use special use cases where it is required.
* *Roger Beep* sounds that are played at the end of each transmission.
* *Muting* of The speaker when pressing PTT to prevent audio feedback and give a radio communication like experience to simulate simplex mode. Both simplex and duplex   settable in XML config. Duplex mode allows you to keep the speaker open for people to interrupt you while speaking.
* LCD/OLED display can show *channel information, server information, who joined, who is speaking, last transmision received date and time, Rotary encoder whether in volume, channel, voice target or radio channel mode, etc.*
* Thes options can be enabled, disabled and customized in the configuration talkkonnect.xml file.