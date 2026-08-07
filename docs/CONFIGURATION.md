# talkkonnect.xml Configuration Manual

This is the reference manual for `talkkonnect.xml`, the single configuration file that
drives talKKonnect. Every tag documented here is derived from the parser
(`xmlparser.go`), the sanity checker (`CheckConfigSanity`), and the code that actually
consumes each value.

For installation, audio setup and hardware wiring see
[running-talkkonnect.md](./running-talkkonnect.md). For the HTTP API request
format see [api.md](./api.md).

---

## Table of Contents

* [1. How the config file is loaded](#1-how-the-config-file-is-loaded)
* [2. Conventions you need to know](#2-conventions-you-need-to-know)
* [3. Document skeleton](#3-document-skeleton)
* [4. `<accounts>`](#4-accounts)
* [5. `<global><software>`](#5-globalsoftware)
  * [5.1 `<settings>`](#51-settings)
  * [5.2 `<channelscan>`](#52-channelscan)
  * [5.3 `<remotesshconsole>`](#53-remotesshconsole)
  * [5.4 `<autoprovisioning>`](#54-autoprovisioning)
  * [5.5 `<beacon>`](#55-beacon)
  * [5.6 `<tts>`](#56-tts)
  * [5.7 `<smtp>`](#57-smtp)
  * [5.8 `<sounds>`](#58-sounds)
  * [5.9 `<remotecontrol>`](#59-remotecontrol)
  * [5.10 `<printvariables>`](#510-printvariables)
  * [5.11 `<ttsmessages>`](#511-ttsmessages)
  * [5.12 `<ignoreuser>`](#512-ignoreuser)
  * [5.13 `<memorychannels>`](#513-memorychannels)
  * [5.14 `<presetvoicetargets>`](#514-presetvoicetargets)
  * [5.15 `<multicast>`](#515-multicast)
* [6. `<global><hardware>`](#6-globalhardware)
  * [6.1 Hardware root attributes](#61-hardware-root-attributes)
  * [6.2 `<io>`](#62-io)
  * [6.3 `<heartbeat>`, `<comment>`, `<listening>`](#63-heartbeat-comment-listening)
  * [6.4 `<lcd>`](#64-lcd)
  * [6.5 `<oled>`](#65-oled)
  * [6.6 `<gps>`](#66-gps)
  * [6.7 `<traccar>`](#67-traccar)
  * [6.8 `<panicfunction>`](#68-panicfunction)
  * [6.9 `<usbkeyboard>` and `<keyboard>`](#69-usbkeyboard-and-keyboard)
  * [6.10 `<audiorecordfunction>`](#610-audiorecordfunction)
  * [6.11 `<radio>` (SA818 transceiver module)](#611-radio-sa818-transceiver-module)
  * [6.12 `<analogrelays>`](#612-analogrelays)
* [7. `<global><multimedia>`](#7-globalmultimedia)
* [8. `<global><Radio>` (internet streaming radio)](#8-globalradio-internet-streaming-radio)
* [9. Live reload — what can change at runtime](#9-live-reload--what-can-change-at-runtime)
* [10. Startup validation and what it does to your config](#10-startup-validation-and-what-it-does-to-your-config)
* [11. Known quirks and traps](#11-known-quirks-and-traps)
* [12. Minimal working config](#12-minimal-working-config)

---

## 1. How the config file is loaded

talKKonnect reads exactly one XML file, chosen with the `-config` flag:

```bash
./talkkonnect -config /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml
```

Other relevant flags:

| Flag | Default | Meaning |
| --- | --- | --- |
| `-config` | `/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml` | Full path to the XML config |
| `-serverindex` | `0` | Start on account index *n* (see [`<accounts>`](#4-accounts)) |
| `-daemon` | `false` | Run headless as a daemon |
| `-cpuprofile`, `-memprofile` | — | Write pprof profiles |

Load sequence at startup:

1. The XML is unmarshalled into the config struct.
2. `CheckConfigSanity()` runs. It **mutates your config in memory** — bad values are
   clamped, and misconfigured features are switched off with a `warn:` line. Anything
   logged as `alert:` is fatal and stops talKKonnect.
3. Derived state is built (account list, USB key map, GPIO memory/voice-target maps,
   event-sound preload, multicast and internet-radio setup).

Reloading:

| Method | Effect |
| --- | --- |
| `Ctrl-B` on the terminal | Live reload — re-reads the file and applies the *reloadable* subset (see [§9](#9-live-reload--what-can-change-at-runtime)) |
| `Ctrl-H` on the terminal | Re-run the sanity checker against the in-memory config |
| `http://<host>:<port>/config` | Web editor: saves the posted XML back to disk, then live-reloads it |
| `cfg set <path> <value>` in the bottom CLI | Change one leaf value at runtime; `cfg save` writes it back, `cfg restart` re-execs |
| Restart the process | Only way to apply hardware sections (GPIO, LCD/OLED, GPS, keyboard devices) |

---

## 2. Conventions you need to know

**`enabled` is an attribute, not an element.** Almost every feature block is switched on
with an attribute on the opening tag:

```xml
<beacon enabled="true">…</beacon>
```

A handful of settings use elements instead (`<insecure>true</insecure>`,
`<localplay>true</localplay>`, `<Enabled>true</Enabled>` under `<Radio>`). The tables
below say which is which.

**Booleans** are `true` / `false`. Anything else parses as `false`.

**Unknown tags are silently ignored.** Go's XML decoder does not error on elements it
does not recognise, so a typo in a tag name means the value is simply never read — with
no warning. Compare against the tables here if a setting appears to do nothing.

**Durations.** Several fields are typed as durations but written as bare integers in the
XML, and the unit is applied in code:

| Tag | Written as | Unit applied |
| --- | --- | --- |
| `<streamonstartafter>`, `<txonstartafter>`, `<repeattxdelay>` | integer | seconds |
| `<voiceactivitytimermsecs>` | integer | milliseconds |
| `<predelay value="…">` / `<postdelay value="…">` under `<multimedia>` | integer | seconds |
| `<predelay>` / `<postdelay>` under `<ttsmessages>` | integer | **nanoseconds** — effectively no delay (see [§11](#11-known-quirks-and-traps)) |
| everything named `*msecs` / `*secs` | integer | as the name says |

**Volumes** come in three flavours depending on where they appear:

* `0–100` percent integers (event sounds, beacon, multimedia, internet radio).
* `0.0–1.0` float gains for into-stream playback (`<beaconvolumeintostream>`,
  `<playvolumeintostream>`, `<panicfunction><volume>`).
* `1–200` percent for multicast (values above 100 amplify).

**File paths** must be absolute. Local sound files are validated at startup — a missing
file disables that sound. Values matching `http`, `https` or `rtsp` are accepted as URLs
and not existence-checked.

---

## 3. Document skeleton

```xml
<?xml version="1.0" encoding="UTF-8"?>
<document>
  <accounts>
    <account name="…" default="true">…</account>
  </accounts>
  <global>
    <software>
      <settings/>          <channelscan/>       <remotesshconsole/>
      <autoprovisioning/>  <beacon/>            <tts/>
      <smtp/>              <sounds/>            <remotecontrol/>
      <printvariables/>    <ttsmessages/>       <ignoreuser/>
      <memorychannels/>    <presetvoicetargets/> <multicast/>
    </software>
    <hardware targetboard="pc">
      <ledstripenabled/>   <gpiooffset/>        <voiceactivitytimermsecs/>
      <io/>                <heartbeat/>         <comment/>
      <listening/>         <lcd/>               <oled/>
      <gps/>               <traccar/>           <panicfunction/>
      <usbkeyboard/>       <keyboard/>          <audiorecordfunction/>
      <radio/>             <analogrelays/>
    </hardware>
    <multimedia>…</multimedia>
    <Radio>…</Radio>
  </global>
</document>
```

Note the two different `radio` sections — `<hardware><radio>` is the SA818 RF module,
`<global><Radio>` (capital R) is internet streaming radio. They are unrelated.

---

## 4. `<accounts>`

One `<account>` per Mumble server. Accounts with `default="true"` form the connection
list; talKKonnect starts on the one selected by `-serverindex` or
`<nextserverindex>`, and `connnextserver` / `previousserver` cycle through them.
Accounts with `default="false"` are parsed but never used.

```xml
<account name="talkkonnect-community" default="true">
  <serverandport>mumble.talkkonnect.com:64738</serverandport>
  <username>my-device</username>
  <password/>
  <insecure>true</insecure>
  <register>false</register>
  <certificate>/path/to/mumble.pem</certificate>
  <channel>HAM-CB</channel>
  <ident>MyTalkkonnectDevice1</ident>
  <listentochannels>
    <channel>POC-SPECIAL</channel>
  </listentochannels>
  <tokens enabled="true">
    <token>token1</token>
  </tokens>
  <voicetargets>
    <id value="1">
      <users><user>recorder-daemon</user></users>
    </id>
  </voicetargets>
</account>
```

| Tag | Type | Notes |
| --- | --- | --- |
| `name` (attr) | string | Label shown in logs and on screen. A missing name warns. |
| `default` (attr) | bool | `true` adds this account to the active list. **At least one is required** — zero default accounts is a fatal `alert:`. |
| `<serverandport>` | `host:port` | Required. Mumble's default port is 64738. Missing is fatal. |
| `<username>` | string | Empty means auto-generate `talkkonnect-<MAC>`, or a random suffix if no MAC is readable. |
| `<password>` | string | Server password. Leave empty (`<password/>`) if none. |
| `<insecure>` | bool | `true` skips TLS certificate verification. Needed for self-signed servers. |
| `<register>` | bool | Request registration on the server. |
| `<certificate>` | path | Combined cert+key PEM for certificate auth. If set but missing on disk you get a warning; if set and unreadable at connect time it is fatal. |
| `<channel>` | string | Channel to join after connecting. |
| `<ident>` | string | Device identity string, used in panic messages and status output. |
| `<listentochannels><channel>` | list | Extra channels to monitor without joining. Activated at boot when `<listentochannelsonstart>` is `true`, or on demand via `listeningstart`. |
| `<tokens><token>` | list | Mumble access tokens sent on connect. |
| `<voicetargets>` | block | Preset whisper/shout targets, see below. |

### `<voicetargets>`

Each `<id value="N">` defines Mumble voice target slot *N* (valid range **1–31**;
0 means "normal channel speech"). Multiple `<id>` blocks may share the same `value` —
their members are merged. Activate a slot with the `voicetargetset` command from the
keyboard, HTTP API or MQTT.

```xml
<voicetargets>
  <id value="1">
    <iscurrent>true</iscurrent>
    <users>
      <user>alice</user>
    </users>
    <channels>
      <channel>
        <name>Root/Operations</name>
        <recursive>true</recursive>
        <links>false</links>
        <group></group>
      </channel>
    </channels>
  </id>
</voicetargets>
```

| Tag | Meaning |
| --- | --- |
| `value` (attr) | Target slot number, 1–31 |
| `<iscurrent>` | Mark this slot as the one to select at startup |
| `<users><user>` | Whisper to these usernames |
| `<channels><channel><name>` | Shout into this channel (full path, e.g. `Root/Ops`) |
| `<recursive>` | Include sub-channels |
| `<links>` | Include linked channels |
| `<group>` | Restrict to an ACL group name |

See [voice-targets.md](./voice-targets.md) for the operational side.

---

## 5. `<global><software>`

### 5.1 `<settings>`

Core behaviour. All children are elements, not attributes.

```xml
<settings>
  <singleinstance>true</singleinstance>
  <outputdevice>Speaker</outputdevice>
  <inputdevice>CAPTURE</inputdevice>
  <logfilenameandpath>/var/log/talkkonnect.log</logfilenameandpath>
  <logging>screenwithlineno</logging>
  <loglevel>info</loglevel>
  …
</settings>
```

**Audio devices**

| Tag | Notes |
| --- | --- |
| `<outputdevice>` | ALSA mixer control used for playback, e.g. `Speaker`, `PCM`, `Master`. |
| `<outputdeviceshort>` | Short label for the LCD/OLED. Defaults to `<outputdevice>` if empty. |
| `<outputvolcontroldevice>` | Mixer control for volume up/down. Defaults to `<outputdevice>`. |
| `<outputmutecontroldevice>` | Mixer control for mute/unmute. Defaults to `<outputdevice>`. |
| `<inputdevice>` | ALSA capture control, e.g. `CAPTURE`, `Mic`. |
| `<openalinputdevice>` | OpenAL capture device name for the Mumble mic. Empty = system default. |
| `<openaloutputdevice>` | OpenAL playback device name for Mumble audio. Empty = system default. |
| `<localplaybackdevice>` | ALSA device for local prompts played through `aplay`/`ffplay`, e.g. `plughw:0,0`. Empty = default. |

> Set `<loglevel>debug</loglevel>` and talkkonnect logs all OpenAL devices it can see at
> startup, which is the quickest way to get the exact names.

**Logging**

| Tag | Values | Notes |
| --- | --- | --- |
| `<logging>` | `screen`, `screenwithlineno`, `screenandfile`, `screenandfilewithlineno` | `…withlineno` adds the Go source file:line to each entry. Anything other than `screen` writes to the log file. |
| `<logfilenameandpath>` | path | If empty and logging is not `screen`, talkkonnect picks the first writable of: config directory, CWD, `/var/log/talkkonnect.log`, `/tmp/talkkonnect.log`. The file is opened append-only and created if missing. |
| `<loglevel>` | `trace`, `debug`, `info`, `warning`, `error`, `alert` | Minimum severity printed. Unrecognised or empty falls back to `info`. This is one of the few values applied by live reload. |

**Startup and transmit behaviour**

| Tag | Type | Notes |
| --- | --- | --- |
| `<singleinstance>` | bool | `true` takes a `talkkonnect.lock` lock file; a second instance announces the clash, waits 5 s and exits. |
| `<cancellablestream>` | bool | Allow an in-progress stream playback to be interrupted. |
| `<streamonstart>` | bool | Start the configured stream automatically after connecting. |
| `<streamonstartafter>` | int (seconds) | Delay before auto-starting the stream. |
| `<streamsendmessage>` | bool | Post a channel text message when streaming starts. |
| `<txonstart>` | bool | Begin transmitting automatically after connecting. |
| `<txonstartafter>` | int (seconds) | Delay before auto-TX. |
| `<repeattxtimes>` | int | Number of transmissions for the `repeattxloop` command. `0` disables the loop. |
| `<repeattxdelay>` | int (seconds) | Gap between repeated transmissions. `0` disables the loop. |
| `<simplexwithmute>` | bool | Mute the speaker while transmitting (simplex/half-duplex behaviour — prevents acoustic feedback on a single speaker+mic build). |
| `<txcounter>` | bool | Count and display PTT presses. |
| `<nextserverindex>` | int | Account index to connect to on startup, overriding `-serverindex`. Must be ≤ the number of default accounts, otherwise it is reset to `0` with a warning. |
| `<txlockout>` | bool | Honour TX lock-out (used by the panic function's `txlockenabled`) and play the `txlockout` event sound when TX is refused. |
| `<listentochannelsonstart>` | bool | Activate the account's `<listentochannels>` list at boot. |

### 5.2 `<channelscan>`

Scanning steps through accessible channels looking for traffic.

```xml
<channelscan>
  <dwelltimemsecs>2000</dwelltimemsecs>
  <hangtimemsecs>3000</hangtimemsecs>
  <returntostartchannel>true</returntostartchannel>
  <skipchannels>Root, 42, Lobby</skipchannels>
</channelscan>
```

| Tag | Notes |
| --- | --- |
| `<dwelltimemsecs>` | How long to sit on each channel. Minimum **500** — smaller non-zero values warn and 500 is used. `0` means use the built-in default. |
| `<hangtimemsecs>` | How long to stay after traffic stops. Minimum **500**, same rule. |
| `<returntostartchannel>` | Return to the original channel when scanning stops. |
| `<skipchannels>` | Comma-separated list of channel **names or numeric IDs** to skip. Whitespace around entries is trimmed and matching is case-insensitive. |

### 5.3 `<remotesshconsole>`

Embedded SSH server exposing the talkkonnect console — the practical way to reach a
daemon-mode unit.

```xml
<remotesshconsole enabled="true">
  <username>suvir</username>
  <password>secret</password>
  <idrsafile>/root/.ssh/id_rsa_tk</idrsafile>
  <listen>0.0.0.0:9999</listen>
</remotesshconsole>
```

All four children are mandatory when enabled — if any is empty the whole feature is
disabled with a warning. `<idrsafile>` is the server's host key (an RSA private key;
generate with `ssh-keygen -t rsa -f /root/.ssh/id_rsa_tk`). `<listen>` is `ip:port`;
use `0.0.0.0` to accept from any interface. Access is additionally filtered by
[`<networkacl>`](#networkacl).

### 5.4 `<autoprovisioning>`

Fetch the config from an HTTP server at boot, then reload from the downloaded copy.

```xml
<autoprovisioning enabled="false">
  <tkid>device_id_12345</tkid>
  <url>http://provisioning.example.com/configs/</url>
  <savefilepath>/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect</savefilepath>
  <savefilename>talkkonnect.xml</savefilename>
</autoprovisioning>
```

All four children are mandatory when enabled, otherwise provisioning is disabled with a
warning. `<tkid>` identifies this device to the server. A provisioning failure is fatal.

### 5.5 `<beacon>`

Periodic identification tone or announcement.

```xml
<beacon enabled="false">
  <beacontimersecs>300</beacontimersecs>
  <beaconfileandpath>/…/soundfiles/voiceprompts/Beacon.wav</beaconfileandpath>
  <localplay>false</localplay>
  <localvolume>75</localvolume>
  <gpioenabled>false</gpioenabled>
  <gpioname>voiceactivity</gpioname>
  <playintostream>false</playintostream>
  <beaconvolumeintostream>0.5</beaconvolumeintostream>
</beacon>
```

| Tag | Notes |
| --- | --- |
| `<beacontimersecs>` | Interval between beacons. `0` disables the beacon (with a warning). |
| `<beaconfileandpath>` | Sound file. Required when enabled. |
| `<localplay>` / `<localvolume>` | Play on the local speaker at `0–100` percent. |
| `<playintostream>` / `<beaconvolumeintostream>` | Transmit into the Mumble channel at gain `0.0–1.0`. |
| `<gpioenabled>` / `<gpioname>` | Drive an output pin while the beacon plays; name must be a configured output pin. |

At least one of `<localplay>` or `<playintostream>` must be `true`, otherwise the beacon
is disabled with a warning.

### 5.6 `<tts>`

Pre-recorded voice prompts for UI events. Each `<sound>` maps an internal event name to
a WAV file.

```xml
<tts enabled="true" language="en-US">
  <volumelevel>10</volumelevel>
  <sound action="talkkonnectloaded" file="/…/Loaded.wav" blocking="true" enabled="true"/>
  <sound action="channelup" file="/…/ChannelUp.wav" blocking="false" enabled="false"/>
</tts>
```

| Attribute | Notes |
| --- | --- |
| `enabled` (on `<tts>`) | Master switch. When `false`, no prompt plays. |
| `<volumelevel>` | Volume for **all** prompts, `0–100`. |
| `action` | Event name — must match the list below exactly (case-sensitive). |
| `file` | WAV to play. An empty `file` means nothing audible happens for that event. |
| `blocking` | `true` waits for playback to finish before continuing. Use for shutdown prompts. |
| `enabled` | Per-prompt switch. |

**Valid `action` values** (these are the names the code actually requests):

`channelup`, `channeldown`, `currentrxvolumelevel`, `currenttxvolumelevel`,
`digitalvolumeup`, `digitalvolumedown`, `displaymenu`, `listonlineusers`,
`listserverchannels`, `mutespeaker`, `unmutespeaker`, `nextserver`, `previousserver`,
`panicsimulation`, `pingservers`, `playstream`, `printxmlconfig`, `quittalkkonnect`,
`requestgpsposition`, `sendemail`, `startscanning`, `stopscanning`,
`starttransmitting`, `stoptransmitting`, `talkkonnectloaded`.

Matching is exact, so `muteSpeaker` (mixed case) or `currentvolumelevel` will never
fire — see [§11](#11-known-quirks-and-traps). Extra `<sound>` entries with unrecognised
`action` values are harmless but dead.

The `language` attribute on `<tts>` is parsed and printed but not used; the language for
generated speech comes from `<ttsmessages><ttslanguage>`.

### 5.7 `<smtp>`

Email alerts, used by the `sendemail` command and the panic function.

```xml
<smtp enabled="false">
  <username>email@example.com</username>
  <password>app_password</password>
  <receiver>alert_receiver@example.com</receiver>
  <subject>Talkkonnect Alert</subject>
  <message>An event occurred.</message>
  <gpsdatetime>false</gpsdatetime>
  <gpslatlong>false</gpslatlong>
  <googlemapsurl>false</googlemapsurl>
</smtp>
```

`<username>`, `<password>` and `<receiver>` are mandatory when enabled; missing any of
them disables SMTP with a warning. The three GPS flags append the current fix time,
latitude/longitude, and a clickable Google Maps link to the message body.

### 5.8 `<sounds>`

Three independent groups: channel event sounds, physical input feedback, and the
repeater tone.

#### Event sounds

```xml
<sounds>
  <sound event="rogerbeep" file="/…/rogerbeeps/Waterdrop.wav" volume="50" blocking="false" enabled="true"/>
  …
</sounds>
```

| `event` | Fires when |
| --- | --- |
| `joinedchannel` | A user joins your channel |
| `leftchannel` | A user leaves your channel |
| `message` | A text message arrives |
| `incommingbeep` | Incoming transmission starts |
| `rogerbeep` | Your transmission ends (pre-loaded into PCM at startup for low latency) |
| `txlockout` | TX was refused because of lock-out |
| `stream` | The stream playback sound |

| Attribute | Notes |
| --- | --- |
| `file` | Absolute path, or an `http`/`rtsp` URL. A missing local file disables the entry with a warning. |
| `volume` | `0–100`. **`volume="0"` disables the sound** and raises a warning. |
| `blocking` | Wait for the sound to finish. |
| `enabled` | Per-event switch. |

#### `<input>` — feedback for physical controls

```xml
<input enabled="true">
  <sound event="iotxpttstart" file="/…/YellowJacket.wav" enabled="true"/>
  <sound event="iovolup"     file="/…/RC210#2.wav"      enabled="true"/>
</input>
```

The `enabled` attribute on `<input>` is a master switch. Event names must match exactly
what the code requests; the GPIO events are prefixed `io`, the USB-keyboard events
`usb`:

*GPIO:* `iotxpttstart`, `iotxpttstop`, `txtogglestart`, `iotxtogglestop`,
`iochannelup`, `iochanneldown`, `iovolup`, `iovoldown`, `iopanic`, `iostreamtoggle`,
`iocommenton`, `iocommentoff`, `iorotarycw`, `iorotaryccw`, `iorotarybutton`,
`iotrackingon`, `iotrackingoff`, `iolisteningstart`, `iolisteningstop`,
`iorepeatertone`, `iocnextserver`, `memorychannel`, `changechannel`, `shutdown`

*USB keyboard:* `usbchannelup`, `usbchanneldown`, `usbvolup`, `usbvoldown`,
`usbcurrentrxvol`, `usbcurrenttxvol`, `usbmute`, `usbunmute`, `usbmutetoggle`,
`usbstarttx`, `usbstoptx`, `usbstartlisten`, `usbstopliosten`, `usbstreamtoggle`,
`usbrecord`, `usbsetcomment`, `usbserverup`, `usbpreviousserver`, `usbvoicetarget`,
`usbmqttpubpayloadset`

There is no volume attribute here — these play through `aplay` at the device volume.
Note `usbstopliosten` and `iocnextserver` are spelled that way in the code; use them
verbatim.

#### `<repeatertone>`

```xml
<repeatertone enabled="true">
  <sound event="repeatertone" tonefrequencyhz="1750" volume="50"
         tonedurationsec="1" direction="local" blocking="false" enabled="true"/>
</repeatertone>
```

| Attribute | Notes |
| --- | --- |
| `tonefrequencyhz` | Tone frequency, e.g. `1750` for European repeater access |
| `tonedurationsec` | Duration in seconds (accepts fractions) |
| `direction` | `intostream` transmits the tone into the Mumble channel; **any other value** (e.g. `local`) plays it on the local speaker |
| `volume`, `blocking`, `enabled` | As elsewhere |

The tone WAV is generated on demand with `ffmpeg` into
`soundfiles/repeatertones/sine_<freq>_<duration>.wav` and cached, so `ffmpeg` and
`ffplay` must be installed.

### 5.9 `<remotecontrol>`

Container for the HTTP API, the status endpoint, the network ACL and MQTT.

#### `<http>` — the REST-ish command API

```xml
<http enabled="true" listenport="8080">
  <command action="channelup" funcparamname="" message="Channel Up" enabled="true"/>
  <command action="joinchannel" funcparamname="value" message="Join Channel" enabled="true"/>
</http>
```

Commands are invoked as `http://<host>:<listenport>/?command=<action>` plus any query
parameters. `?command=listapi` lists everything enabled.

A command runs only if **both** conditions hold:

1. `action` is one of the built-in handlers listed below, and
2. a `<command>` element with that `action` exists in this section.

An action not in the built-in list gets a `400`; an action missing from the XML gets a
`404`. `enabled="false"` keeps a command out of the `listapi` listing.

**Built-in actions:**

`displaymenu`, `channelup`, `channeldown`, `mute`, `unmute`, `mute-toggle`,
`currentrxvolume`, `currenttxvolume`, `volumerxup`, `volumerxdown`, `volumetxup`,
`volumetxdown`, `setrxvolume`, `listserverchannels`, `joinchannel`, `whisperuser`,
`whisperclear`, `starttransmitting`, `stoptransmitting`, `listonlineusers`, `playback`,
`gpsposition`, `sendemail`, `previousserver`, `connnextserver`, `clearscreen`,
`pingservers`, `panicsimulation`, `repeattxloop`, `scanchannels`, `thanks`,
`showuptime`, `showversion`, `dumpxmlconfig`, `ttsannouncement`, `announcement`,
`voicetargetset`, `listeningstart`, `listeningstop`, `radiotoggle`, `radionext`,
`radioprev`, `radiovolup`, `radiovoldown`, `multicaston`, `multicastoff`,
`multicasttoggle`, `listapi`.

**Query parameters** (all case-insensitive; the `funcparamname` attribute is
documentation only — the parameter names below are what the server reads):

| Parameter | Used by | Type |
| --- | --- | --- |
| `id` | `voicetargetset` | integer 0–31 |
| `channel` | `joinchannel` | channel name |
| `user` | `whisperuser` | username |
| `volume` | `setrxvolume` | integer |
| `mediaid` | `announcement` | `<multimedia>` profile id |
| `ttsmessage` | `ttsannouncement` | text to speak |
| `ttslocalplay`, `ttsplayintostream` | `ttsannouncement` | `true`/`false` |
| `gpioenabled`, `gpioname` | `ttsannouncement` | `true`/`false`, pin name |
| `predelay`, `postdelay` | `ttsannouncement` | integer |
| `language` | `ttsannouncement` | language code, e.g. `en` |

Example:

```bash
curl "http://192.168.1.50:8080/?command=ttsannouncement&ttsmessage=Muster+at+gate+3&ttslocalplay=true"
curl "http://192.168.1.50:8080/?command=joinchannel&channel=Root/Operations"
```

The listener also serves the XML config editor at `/config`. See
[api.md](./api.md) for full request/response detail.

#### `<uistatus>` — JSON status for external displays

```xml
<uistatus enabled="true" listenip="127.0.0.1" listenport="8080" url="/uistatus"/>
```

| Attribute | Default | Notes |
| --- | --- | --- |
| `enabled` | follows `<http enabled>` **if the `<uistatus>` element is absent entirely** | Set `enabled="false"` when using the built-in LCD/OLED rather than an external framebuffer client |
| `listenip` | `127.0.0.1` | `0.0.0.0` binds all interfaces |
| `listenport` | `<http listenport>`, else `8080` | If it equals the HTTP API port, the path is served by the same listener; otherwise a second listener starts |
| `url` | `/uistatus` | A leading `/` is added if you omit it |

Returns pretty-printed JSON with the current channel, users, and talk state.

#### `<networkacl>`

```xml
<networkacl enabled="true">
  <network cidr="127.0.0.1/32"/>
  <network cidr="192.168.1.0/24"/>
</networkacl>
```

Restricts the HTTP API, `/config`, `/uistatus` **and** the SSH console to these CIDRs.
Invalid CIDRs are skipped with a warning. If enabled with no valid entries the ACL is
inactive and everything is allowed — you get a warning saying so. Non-matching clients
receive `403`.

#### `<mqtt>`

```xml
<mqtt enabled="false">
  <settings enabled="true">
    <mqttsubtopic>talkkonnect/+/control</mqttsubtopic>
    <mqttpubtopic>talkkonnect/status</mqttpubtopic>
    <mqttbroker>tcp://mqtt.example.com:1883</mqttbroker>
    <mqttuser>mqtt_user</mqttuser>
    <mqttpassword>mqtt_pass</mqttpassword>
    <mqttid>talkkonnect_mqtt_client</mqttid>
    <cleansess>true</cleansess>
    <qos>1</qos>
    <num>1</num>
    <payload>default_payload</payload>
    <action>default_action</action>
    <store>/tmp/mqttstore</store>
    <retained>false</retained>
    <attentionblinktimes>3</attentionblinktimes>
    <attentionblinkmsecs>500</attentionblinkmsecs>
    <pubpayload>
      <mqtt item="0" payload="{&quot;status&quot;:&quot;online&quot;}" enabled="true"/>
    </pubpayload>
  </settings>
  <commands>
    <command action="channelup" message="Channel Up" enabled="true"/>
  </commands>
</mqtt>
```

`<mqttsubtopic>`, `<mqttpubtopic>`, `<mqttbroker>`, `<mqttpassword>` and `<mqttid>` are
all mandatory when enabled — an empty one disables MQTT with a warning. (`<mqttuser>`
is not checked.)

| Tag | Notes |
| --- | --- |
| `<mqttsubtopic>` | Topic to subscribe to for commands. `+` and `#` wildcards allowed. |
| `<mqttpubtopic>` | Topic for published status. |
| `<mqttbroker>` | `tcp://host:port` or `ssl://host:port`. |
| `<cleansess>` | Clean-session flag. |
| `<qos>` | `0`, `1` or `2`. |
| `<store>` | Directory for the persistence store when `cleansess` is false. |
| `<retained>` | Publish status with the retained flag. |
| `<attentionblinktimes>` / `<attentionblinkmsecs>` | Blink count and on/off period for the `attention` command. |
| `<pubpayload><mqtt item="N" payload="…">` | Payloads selectable by the `mqttpubpayloadset` keyboard action, addressed by `item`. Payload JSON must be XML-escaped (`&quot;`). |

**Command payload format.** The payload is lower-cased, `:` is stripped, then split on
spaces. Word 0 is the action and must match a `<command action>` entry. Actions taking
arguments:

| Payload | Effect |
| --- | --- |
| `muteunmute mute` / `muteunmute unmute` / `muteunmute toggle` | Speaker mute control |
| `attention on` / `attention off` / `attention blink` | Drive the `attention` output pin |
| `relay <1\|2> on\|off\|pulse` | Intended to drive the `relay1` / `relay2` output pin. **Currently broken** — see [§11](#11-known-quirks-and-traps) |
| `voicetargetset <0-31>` | Select a voice target slot |
| `announcement <profileid>` | Play a `<multimedia>` profile |
| any other action, no arguments | Runs the command |

**MQTT actions** are a *different, smaller* set than the HTTP ones — note `muteunmute`
in place of `mute`/`unmute`/`mute-toggle`, and no `ttsannouncement`, `showversion`,
`joinchannel`, `whisperuser`, `setrxvolume` or `listapi`:

`displaymenu`, `channelup`, `channeldown`, `muteunmute`, `currentrxvolume`,
`currenttxvolume`, `volumerxup`, `volumerxdown`, `volumetxup`, `volumetxdown`,
`listserverchannels`, `starttransmitting`, `stoptransmitting`, `listonlineusers`,
`playback`, `gpsposition`, `sendemail`, `previousserver`, `connnextserver`,
`clearscreen`, `pingservers`, `panicsimulation`, `repeattxloop`, `scanchannels`,
`thanks`, `showuptime`, `dumpxmlconfig`, `announcement`, `voicetargetset`,
`listeningstart`, `listeningstop`, `radiotoggle`, `radionext`, `radioprev`,
`radiovolup`, `radiovoldown`, `multicaston`, `multicastoff`, `multicasttoggle`,
`attention`, `relay`.

### 5.10 `<printvariables>`

Pure diagnostics: each flag controls whether that section is dumped to the log at
startup and on `dumpxmlconfig`. All are `true`/`false` elements.

`printaccount`, `printsystemsettings`, `printremotesshconsole`, `printprovisioning`,
`printbeacon`, `printtts`, `printsmtp`, `printsounds`, `printhttpapi`, `printmqtt`,
`printttsmessages`, `printignoreuser`, `printhardware`, `printgpioexpander`,
`printmax7219`, `printpins`, `printrotary`, `printpulse`, `printvolumebuttonstep`,
`printheartbeat`, `printcomment`, `printlcd`, `printoled`, `printgps`, `printtraccar`,
`printpanic`, `printusbkeyboard`, `printaudiorecord`, `printkeyboardmap`,
`printradiomodule`, `printmultimedia`, `printmemorychannels`,
`printpresetvoicetargets`, `printmulticast`.

`<printlistentochannels>` is a string (`all` in the shipped config) and is currently
parsed but not consulted.

Turning these off is worthwhile on a production unit — `printaccount` dumps account
details (passwords are masked).

### 5.11 `<ttsmessages>`

Generated speech (Google TTS via `htgotts`), as opposed to the pre-recorded prompts in
[`<tts>`](#56-tts). Used by `ttsannouncement`, station announcements and event
messages.

```xml
<ttsmessages enabled="true">
  <ttslanguage>en</ttslanguage>
  <ttslanguagethai>th</ttslanguagethai>
  <ttsmessagefromtag>false</ttsmessagefromtag>
  <ttstone file="/…/announcement-01.wav" volume="10" enabled="true"/>
  <localblocking>true</localblocking>
  <ttssounddirectory>audio</ttssounddirectory>
  <localplay>true</localplay>
  <playintostream>false</playintostream>
  <speakvolumeintostream>50</speakvolumeintostream>
  <playvolumeintostream>0.6</playvolumeintostream>
  <gpio name="tts_active_led" enabled="true"/>
  <predelay value="500" enabled="true"/>
  <postdelay value="500" enabled="true"/>
</ttsmessages>
```

| Tag | Notes |
| --- | --- |
| `<ttslanguage>` | Language code for synthesis (`en`, `de`, `th`, …). This is the value actually used. |
| `<ttslanguagethai>` | Language code used when the text is detected as Thai. |
| `<ttsmessagefromtag>` | Speak incoming Mumble text messages. |
| `<ttstone …>` | Attention tone played before the announcement. `file`, `volume` (0–100), `enabled`. |
| `<localblocking>` | Wait for local playback to finish. |
| `<ttssounddirectory>` | Cache directory for generated MP3s (relative paths resolve against the working directory). Created if missing. |
| `<localplay>` | Play on the local speaker. |
| `<playintostream>` | Transmit into the Mumble channel. |
| `<speakvolumeintostream>` | Volume `0–100` for synthesised speech into the stream. |
| `<playvolumeintostream>` | Gain `0.0–1.0` for file playback into the stream. |
| `<gpio name= enabled=>` | Drive a named output pin during the announcement. |
| `<predelay>` / `<postdelay>` | See the caveat in [§11](#11-known-quirks-and-traps) — these do not currently produce a usable delay. |

Generated speech requires outbound internet access. Cached files are reused, so
repeated announcements work offline.

### 5.12 `<ignoreuser>`

```xml
<ignoreuser enabled="false">
  <ignoreuserregex>(?i)bot|spam</ignoreuserregex>
</ignoreuser>
```

Audio and messages from usernames matching the Go regular expression are ignored. The
regex must be at least 4 characters, otherwise the feature is disabled with a warning.
`(?i)` makes it case-insensitive.

### 5.13 `<memorychannels>`

Binds a GPIO button to a fixed channel — press the button, jump straight to that
channel.

```xml
<memorychannels enabled="true">
  <channel gpioname="memorychannel1" channelname="HAM-CB" enabled="true"/>
  <channel gpioname="memorychannel2" channelname="Root/Favorite2" enabled="true"/>
</memorychannels>
```

`gpioname` **must** be one of `memorychannel1` … `memorychannel4`; any other value
disables the whole section with a warning. A matching input pin with the same `name`
must also exist under [`<pins>`](#pins).

### 5.14 `<presetvoicetargets>`

Same idea for voice targets.

```xml
<presetvoicetargets enabled="true">
  <voicetargetset gpioname="presetvoicetarget1" id="10" enabled="true"/>
</presetvoicetargets>
```

`gpioname` must be one of `presetvoicetarget1` … `presetvoicetarget5`, otherwise the
section is disabled with a warning. `id` is the voice target slot (1–31) defined under
the account's `<voicetargets>`.

### 5.15 `<multicast>`

Sends received channel audio to an RTP multicast group so IP speakers and SIP phones can
play it. See [multicast.md](./multicast.md).

```xml
<multicast enabled="true">
  <group>239.0.1.10</group>
  <port>5004</port>
  <codec>pcmu</codec>
  <ttl>1</ttl>
  <interface>eth0</interface>
  <packetms>20</packetms>
  <l16payloadtype>96</l16payloadtype>
  <volume>100</volume>
  <allchannels>false</allchannels>
  <hangoverms>200</hangoverms>
  <include>
    <user>dispatcher</user>
  </include>
  <exclude>
    <user>noisybox</user>
  </exclude>
</multicast>
```

| Tag | Valid values | Default | Notes |
| --- | --- | --- | --- |
| `<group>` | IPv4 multicast address in `224.0.0.0/4` | — | Required. A non-multicast or IPv6 address disables the section. |
| `<port>` | 1–65535 | `5004` | Should be **even** — RTP convention reserves the odd port above for RTCP. An odd port warns. |
| `<codec>` | `pcmu`, `pcma`, `alaw`, `g711a`, `l16`, `pcm`, `raw` | `pcmu` | Unrecognised values fall back to `pcmu` with a warning. `l16` warns loudly: it is uncompressed PCM on a dynamic payload type and G.711-only hardware will play **silence**. |
| `<ttl>` | 1–255 | `1` | `1` keeps traffic on the local subnet. |
| `<interface>` | interface name | default route | A name not present on the host warns and the default route is used. |
| `<packetms>` | `10`, `20`, `30`, `40`, `60` | `20` | Other values warn and reset to 20. |
| `<l16payloadtype>` | 96–127 | `96` | Only relevant for `l16`. |
| `<volume>` | 1–200 (percent) | `100` | Above 100 amplifies. |
| `<allchannels>` | bool | `false` | `true` multicasts traffic from every channel you monitor, not just the current one. |
| `<hangoverms>` | 0–5000 | `200` | How long a silent talker holds the RTP stream open. |
| `<include><user>` | list | empty | Whitelist. **Empty means every talker in the channel.** |
| `<exclude><user>` | list | empty | Blacklist. Exclude wins over include; a name in both warns. |

`ffmpeg` must be on `PATH` for `<multimedia>` profiles to be multicast.

---

## 6. `<global><hardware>`

### 6.1 Hardware root attributes

```xml
<hardware targetboard="rpi">
  <ledstripenabled>false</ledstripenabled>
  <gpiooffset>0</gpiooffset>
  <voiceactivitytimermsecs>200</voiceactivitytimermsecs>
  …
</hardware>
```

| Tag | Notes |
| --- | --- |
| `targetboard` (attr) | **`rpi`** enables all GPIO, LCD and OLED hardware paths. Any other value (conventionally `pc`) runs software-only: GPIO calls become no-ops and the LCD backlight timer is forced off. Set this to `pc` on a VM or desktop. |
| `<ledstripenabled>` | Enable the addressable LED strip. When on, pins 7–11 clash with SPI and warn. |
| `<gpiooffset>` | Added to every `<pin pinno>` at startup when greater than 0. For boards whose gpiochip base is not 0 (many Orange Pi / mainline-kernel systems). |
| `<voiceactivitytimermsecs>` | How long the voice-activity LED stays lit after audio. Minimum **200**; `0` or anything smaller is raised to 200. |

### 6.2 `<io>`

#### `<gpioexpander>` — MCP23017 I²C port expanders

```xml
<gpioexpander enabled="true">
  <chip id="0" i2cbus="1" mcp23017device="32" enabled="true"/>
</gpioexpander>
```

| Attribute | Notes |
| --- | --- |
| `id` | Chip identifier referenced by a pin's `chipid`. Must be ≤ 8. |
| `i2cbus` | I²C bus number (usually `1` on a Pi). |
| `mcp23017device` | I²C address in **decimal** (`32` = 0x20). |

When any expander is enabled, GPIO pins 2 and 3 warn about clashing with the I²C bus.

#### `<max7219>` — 7-segment / matrix display

```xml
<max7219 enabled="false">
  <max7219cascaded>1</max7219cascaded>
  <spibus>0</spibus>
  <spidevice>0</spidevice>
  <brightness>5</brightness>
</max7219>
```

Note these are elements in the shipped config but are declared as attributes in the
parser (`max7219cascaded`, `spibus`, `spidevice`, `brightness` are all `,attr`), so they
must be written as attributes on `<max7219>` to be read:

```xml
<max7219 enabled="false" max7219cascaded="1" spibus="0" spidevice="0" brightness="5"/>
```

#### `<pins>`

The heart of a hardware build. One `<pin>` per physical connection.

```xml
<pins>
  <pin direction="output" device="led/relay" name="transmit"  pinno="22" type="gpio" chipid="0" inverted="false" enabled="true"/>
  <pin direction="input"  device="pushbutton" name="txptt"    pinno="26" type="gpio" chipid="0" enabled="true"/>
</pins>
```

| Attribute | Valid values | Notes |
| --- | --- | --- |
| `direction` | `input`, `output` | Anything else disables the pin. |
| `device` | inputs: `pushbutton`, `toggleswitch`, `rotaryencoder`<br>outputs: `led/relay`, `lcd` | Mismatched direction/device disables the pin with a warning. |
| `name` | see tables below | Must be a recognised name or the pin is disabled. |
| `pinno` | **1–27** or **513–539** | Anything else disables the pin. BCM numbering. `<gpiooffset>` is added if set. |
| `type` | `gpio`, `mcp23017` | `gpio` = SoC pin. `mcp23017` routes through the expander named by `chipid`. |
| `chipid` | 0–8 | Which `<gpioexpander><chip id>` this pin belongs to. |
| `inverted` | bool | Outputs only: `true` drives the pin low for "on" (active-low relay boards and sinking LED wiring). |
| `enabled` | bool | Per-pin switch. Disabled pins are not validated. |

**Output pin names** (`direction="output"`):

| `name` | Drives |
| --- | --- |
| `voiceactivity` | Lit while receiving audio |
| `participants` | Lit when others are in the channel |
| `transmit` | Lit while transmitting |
| `online` | Lit while connected to the server |
| `attention` | Driven by the MQTT `attention` command |
| `voicetarget` | Lit while a voice target is active |
| `heartbeat` | Blinks per [`<heartbeat>`](#63-heartbeat-comment-listening) |
| `backlight` | LCD backlight — use `device="lcd"` |
| `relay0` | Generic relay. Accepted by the validator, but no code path currently drives it — see the MQTT `relay` note in [§11](#11-known-quirks-and-traps) |

**Input pin names** (`direction="input"`):

| `name` | Action |
| --- | --- |
| `txptt` | Momentary push-to-talk |
| `txtoggle` | Latching transmit toggle |
| `channelup` / `channeldown` | Step through channels |
| `volup` / `voldown` | Volume, stepped by [`<volumebuttonstep>`](#volumebuttonstep) |
| `panic` | Trigger the [panic function](#68-panicfunction) |
| `streamtoggle` | Start/stop stream playback |
| `comment` | Toggle the Mumble comment — use `device="toggleswitch"` |
| `rotarya`, `rotaryb`, `rotarybutton` | Rotary encoder A/B phases and its push button — use `device="rotaryencoder"` |
| `nextserver` | Connect to the next account |
| `repeatertone` | Play the repeater tone |
| `memorychannel1` … `memorychannel4` | Jump to a channel from [`<memorychannels>`](#513-memorychannels) |
| `presetvoicetarget1` … `presetvoicetarget5` | Select a slot from [`<presetvoicetargets>`](#514-presetvoicetargets) |
| `shutdown` | Shut the unit down |

`analogrelay1` and `analogrelay2` are also accepted names (used by
[`<analogrelays>`](#612-analogrelays)).

> **Names handled in code but rejected by the validator.** `listening`, `tracking`,
> `mqtt0`, `mqtt1`, `internetradiotoggle`, `internetradiochannelup`,
> `internetradiochanneldown`, `internetradiovolup` and `internetradiovoldown` have
> working handlers but are **not** in the validator's allow-list, so an enabled pin using
> one is disabled at startup with `Invalid Name`. Reach these functions via the keyboard,
> HTTP API or MQTT instead. Likewise `radiomodule` is not a valid `device` value even
> though `<pins>` examples use it.

Only `type="gpio"` inputs are polled as buttons; MCP23017 pins are configured as inputs
but the button-press paths run on SoC pins.

#### `<rotaryencoder>`

```xml
<rotaryencoder enabled="true">
  <control function="mumblechannel" enabled="true"/>
  <control function="localvolume" enabled="true"/>
</rotaryencoder>
```

Valid `function` values: `mumblechannel`, `localvolume`, `radiochannel`,
`voicetarget`. Enabled controls form a ring — the encoder's push button cycles to the
next function, and turning the knob acts on the current one. Requires `rotarya`,
`rotaryb` and `rotarybutton` pins.

#### `<pulse>`

```xml
<pulse leadingmsecs="1000" pulsemsecs="1000" trailingmsecs="1000"/>
```

Attributes (not elements). Timing for pulsed relay output: delay before the pulse, pulse
width, and delay after.

#### `<volumebuttonstep>`

```xml
<volumebuttonstep>
  <volupstep>1</volupstep>
  <voldownstep>-1</voldownstep>
</volumebuttonstep>
```

Mixer steps per button press. `0` is replaced by `+1` / `-1`. `<voldownstep>` should be
negative.

### 6.3 `<heartbeat>`, `<comment>`, `<listening>`

```xml
<heartbeat enabled="true">
  <heartbeatledpin>heartbeat</heartbeatledpin>
  <periodmsecs>2000</periodmsecs>
  <ledonmsecs>1000</ledonmsecs>
  <ledoffmsecs>1010</ledoffmsecs>
</heartbeat>
```

A "the software is alive" blink. All three timings should be ≥ 100 ms. Requires an
output pin named `heartbeat`.

```xml
<comment>
  <commentbuttonpin>comment</commentbuttonpin>
  <commentmessageoff>Status: Available</commentmessageoff>
  <commentmessageon>Status: Busy</commentmessageon>
</comment>
```

The two messages are the Mumble comment set for each position of the `comment` toggle
switch. `<commentbuttonpin>` is parsed but ignored — the binding comes from the pin
*named* `comment` under `<pins>`.

```xml
<listening enabled="false">
  <listeningbuttonpin>listening</listeningbuttonpin>
</listening>
```

Neither child is read by the code; listen-to-channel is driven by
`<listentochannelsonstart>` and the `listeningstart` / `listeningstop` commands.

### 6.4 `<lcd>`

HD44780 character LCD, either 4-bit parallel or over an I²C backpack. Requires
`targetboard="rpi"`.

```xml
<lcd enabled="true">
  <lcdinterfacetype>i2c</lcdinterfacetype>
  <lcdi2caddress>63</lcdi2caddress>
  <lcdbacklighttimerenabled>true</lcdbacklighttimerenabled>
  <lcdbacklighttimeoutsecs>30</lcdbacklighttimeoutsecs>
  <lcdbacklightpin>6</lcdbacklightpin>
  <lcdrspin>7</lcdrspin>
  <lcdepin>8</lcdepin>
  <lcdd4pin>25</lcdd4pin>
  <lcdd5pin>24</lcdd5pin>
  <lcdd6pin>23</lcdd6pin>
  <lcdd7pin>18</lcdd7pin>
</lcd>
```

| Tag | Notes |
| --- | --- |
| `<lcdinterfacetype>` | `i2c` or `parallel`. Anything else disables the LCD. |
| `<lcdi2caddress>` | **Decimal** I²C address — convert the hex from `i2cdetect` (`0x3f` → `63`). Required and non-zero for `i2c`. |
| `<lcdrspin>` … `<lcdd7pin>` | BCM pin numbers, required and non-zero for `parallel`. A zero disables the LCD. |
| `<lcdbacklighttimerenabled>` | Turn the backlight off after inactivity. Forced off unless `targetboard="rpi"`. Also disabled unless both OLED and LCD are enabled (see [§11](#11-known-quirks-and-traps)). |
| `<lcdbacklighttimeoutsecs>` | Idle seconds before blanking. `0` disables the timer with a warning. |
| `<lcdbacklightpin>` | Backlight control pin. |

### 6.5 `<oled>`

SSD1306-class I²C OLED. Requires `targetboard="rpi"`. If the panel does not respond at
startup, OLED support switches itself off and logs `Cannot Communicate with OLED`.

```xml
<oled enabled="true">
  <oledinterfacetype>i2c</oledinterfacetype>
  <oleddisplayrows>10</oleddisplayrows>
  <oleddisplaycolumns>21</oleddisplaycolumns>
  <oleddefaulti2cbus>1</oleddefaulti2cbus>
  <oleddefaulti2caddress>60</oleddefaulti2caddress>
  <oledscreenwidth>132</oledscreenwidth>
  <oledscreenheight>60</oledscreenheight>
  <oledcommandcolumnaddressing>33</oledcommandcolumnaddressing>
  <oledaddressbasepagestart>176</oledaddressbasepagestart>
  <oledcharlength>5</oledcharlength>
  <oledstartcolumn>1</oledstartcolumn>
</oled>
```

| Tag | Typical | Notes |
| --- | --- | --- |
| `<oledinterfacetype>` | `i2c` | |
| `<oleddisplayrows>` / `<oleddisplaycolumns>` | `10` / `21` | Text grid size. |
| `<oleddefaulti2cbus>` | `1` | |
| `<oleddefaulti2caddress>` | `60` | **Decimal** — `0x3c` from `i2cdetect` is `60`. |
| `<oledscreenwidth>` / `<oledscreenheight>` | `132` / `60` | Panel pixels. |
| `<oledcommandcolumnaddressing>` | `33` | Controller command byte (0x21). |
| `<oledaddressbasepagestart>` | `176` | Controller page-start base (0xB0). |
| `<oledcharlength>` | `5` | Font width in pixels. |
| `<oledstartcolumn>` | `1` | Left offset. |

Enabling the OLED makes GPIO pins 2 and 3 warn as I²C clashes. `<oledttf …>` appears in
some shipped configs but is not parsed.

### 6.6 `<gps>`

NMEA GPS on a serial port.

```xml
<gps enabled="true">
  <port>/dev/ttyAMA0</port>
  <baud>9600</baud>
  <txdata/>
  <even>false</even>
  <odd>false</odd>
  <rs485>false</rs485>
  <rs485highduringsend>false</rs485highduringsend>
  <rs485highaftersend>false</rs485highaftersend>
  <stopbits>1</stopbits>
  <databits>8</databits>
  <chartimeout>100</chartimeout>
  <minread>1</minread>
  <rx>true</rx>
  <gpsinfoverbose>false</gpsinfoverbose>
  <gpsdiagsounds>true</gpsdiagsounds>
  <gpsdisplayshow>true</gpsdisplayshow>
</gps>
```

| Tag | Notes |
| --- | --- |
| `<port>` | Device path. **Must exist at startup** or GPS is disabled with a warning. |
| `<baud>` | One of `2400`, `4800`, `9600`, `14400`, `19200`, `38400`, `57600`, `115200`. Anything else disables GPS. |
| `<even>` / `<odd>` | Parity. Setting **both** disables GPS. Both `false` = no parity. |
| `<stopbits>` / `<databits>` | Must be non-zero, typically `1` and `8`. |
| `<chartimeout>` / `<minread>` | Read timing: inter-character timeout and minimum bytes per read. |
| `<rx>` | Enable receiving. |
| `<txdata>` | Optional initialisation string sent to the module. |
| `<rs485*>` | RS-485 transceiver direction control. |
| `<gpsinfoverbose>` | Log every parsed sentence — noisy, for bring-up only. |
| `<gpsdiagsounds>` | Audible feedback on fix acquired/lost. |
| `<gpsdisplayshow>` | Show position on the LCD/OLED. |

### 6.7 `<traccar>`

Report GPS position to a [Traccar](https://www.traccar.org/) server.

```xml
<traccar enabled="true">
  <track>true</track>
  <clientid>traccar_device_1</clientid>
  <devicescreenenabled>true</devicescreenenabled>
  <traccardiagsounds>true</traccardiagsounds>
  <traccardispayshow>true</traccardispayshow>
  <protocol name="osmand">
    <osmand port="5055" serverurl="http://traccar.example.com"/>
    <t55 port="5001" serverip="traccar.example.com"/>
    <opengts port="5177" serverurl="http://traccar.example.com"/>
  </protocol>
</traccar>
```

| Tag | Notes |
| --- | --- |
| `<track>` | Master switch for sending positions — both this **and** `enabled` must be true. |
| `<clientid>` | Device identifier registered in Traccar. |
| `<devicescreenenabled>` / `<traccardispayshow>` | Show tracking state on the display. Note the typo in `traccardispayshow` — it is spelled that way in the parser. |
| `<traccardiagsounds>` | Audible feedback on report success/failure. |
| `<protocol name>` | `osmand`, `t55` or `opengts` — selects which child block is used. |

Only the block matching `name` is used; the others can stay for reference. `osmand` and
`opengts` take a `serverurl` (with scheme), `t55` takes a bare `serverip`. The default
Traccar ports are 5055 (OsmAnd), 5001 (T55) and 5177 (OpenGTS).

### 6.8 `<panicfunction>`

Emergency alert triggered by the `panic` input pin or the `panicsimulation` command.

```xml
<panicfunction enabled="true">
  <filenameandpath>/…/soundfiles/alerts/alert.wav</filenameandpath>
  <volume>10</volume>
  <blocking>false</blocking>
  <sendident>true</sendident>
  <panicmessage>I Need Help! Now!</panicmessage>
  <panicemail>false</panicemail>
  <eavesdrop>false</eavesdrop>
  <recursivesendmessage>false</recursivesendmessage>
  <sendgpslocation>true</sendgpslocation>
  <txlockenabled>true</txlockenabled>
  <txlocktimeoutsecs>60</txlocktimeoutsecs>
  <lowprofile>false</lowprofile>
</panicfunction>
```

| Tag | Notes |
| --- | --- |
| `<filenameandpath>` | Alert sound transmitted into the channel. |
| `<volume>` | Into-stream gain, `0.0–1.0`. |
| `blocking` | **Attribute on `<panicfunction>`**, not an element, in the parser. |
| `<sendident>` | Include the account's `<ident>` in the message. |
| `<panicmessage>` | Text message sent to the channel. |
| `<panicemail>` | Also send email — requires [`<smtp>`](#57-smtp) enabled. |
| `<eavesdrop>` | Open the microphone so the channel can hear the scene. |
| `<recursivesendmessage>` | Send the text to sub-channels as well. |
| `<sendgpslocation>` | Append the current fix — requires [`<gps>`](#66-gps). |
| `<txlockenabled>` / `<txlocktimeoutsecs>` | Hold TX locked open for this many seconds after the panic. Honoured only when `<txlockout>` is true in `<settings>`. |
| `<lowprofile>` | Silent panic: suppress local sound and screen indication so a bystander cannot tell it fired. |

### 6.9 `<usbkeyboard>` and `<keyboard>`

#### `<usbkeyboard>`

```xml
<usbkeyboard enabled="true">
  <usbkeyboarddevs>
    <usbkeyboarddevpath>/dev/input/by-id/usb-C-Media_…-event-if03</usbkeyboarddevpath>
    <usbkeyboarddevpath>/dev/input/by-id/usb-Another_Device-event-kbd</usbkeyboarddevpath>
  </usbkeyboarddevs>
  <numlockscanid>69</numlockscanid>
</usbkeyboard>
```

Use stable `/dev/input/by-id/...` paths rather than `/dev/input/eventN`, which renumber
across reboots. Multiple devices are supported — many USB PTT handsets and CM-series
sound cards present their buttons as a HID keyboard. A single legacy
`<usbkeyboarddevpath>` directly under `<usbkeyboard>` is still accepted.
`<numlockscanid>` is the scan code treated as Num Lock.

#### `<keyboard>`

Maps key scan codes to actions. Each `<command>` is one action; the nested
`<usbkeyboard>` element binds a USB key to it.

```xml
<keyboard>
  <command action="soundinterfacepttkey" paramname="" paramvalue="" enabled="true">
    <usbkeyboard scanid="115" keylabel="0" enabled="true"/>
  </command>
  <command action="voicetargetset" paramname="voicetargetset" paramvalue="3" enabled="true">
    <usbkeyboard scanid="81" keylabel="3" enabled="true"/>
  </command>
</keyboard>
```

| Attribute | Notes |
| --- | --- |
| `action` | Must be in the list below, otherwise the command is disabled with a warning. |
| `paramname` / `paramvalue` | Argument for actions that need one — e.g. the target slot for `voicetargetset`, the comment text for `setcomment`. |
| `enabled` | Per-command switch. |
| `<usbkeyboard scanid>` | Linux input scan code, **1–255**. `0` or >255 disables the binding with a warning. |
| `<usbkeyboard keylabel>` | Numeric label shown in the on-screen key map. |
| `<usbkeyboard enabled>` | Per-binding switch. |

**Valid `action` values:** `channelup`, `channeldown`, `serverup`, `serverdown`, `mute`,
`unmute`, `mute-toggle`, `stream-toggle`, `volumeup`, `volumedown`, `volumerxup`,
`volumerxdown`, `volumetxup`, `volumetxdown`, `volup`, `voldown`, `setcomment`,
`transmitstart`, `transmitstop`, `pttkey`, `soundinterfacepttkey`, `record`,
`voicetargetset`, `mqttpubpayloadset`, `changechannel`, `listentochannelon`,
`listentochanneloff`, `gpioinput`, `gpiooutput`, `radiotoggle`, `radionext`,
`radioprev`, `radiovolup`, `radiovoldown`.

`<ttykeyboard>` elements appear in shipped configs but are **not parsed** — only
`<usbkeyboard>` bindings take effect. Find scan codes with `evtest` or
`showkey --scancodes`.

### 6.10 `<audiorecordfunction>`

Records channel traffic to disk, optionally indexed in MySQL/MariaDB.

```xml
<audiorecordfunction enabled="true">
  <recordonstart>true</recordonstart>
  <recordmode>traffic</recordmode>
  <recordsavepath>/var/spool/talkkonnect/recordings/</recordsavepath>
  <recordbasename>talkkonnect</recordbasename>
  <maxfilesize>500000</maxfilesize>
  <recordindexlog>recording_index.log</recordindexlog>
  <channelbuffersize>4096</channelbuffersize>
  <writeflushinterval>1000</writeflushinterval>
  <recorddb enabled="true">
    <host>127.0.0.1</host>
    <port>3306</port>
    <user>talkkonnect</user>
    <password>talkkonnect</password>
    <database>Talkkonnect_recordings</database>
  </recorddb>
</audiorecordfunction>
```

| Tag | Notes |
| --- | --- |
| `<recordonstart>` | Begin recording as soon as talkkonnect connects. |
| `<recordmode>` | **`traffic` is the only supported mode** (writes `.mrec` files). Any other non-empty value logs a warning and records nothing. |
| `<recordsavepath>` | Directory for recordings — must exist and be writable. |
| `<recordbasename>` | Filename prefix. |
| `<maxfilesize>` | Bytes per file before rolling over. |
| `<recordindexlog>` | Index file listing recordings with timestamps. |
| `<channelbuffersize>` | In-memory audio buffer in samples. Raise if you see dropped audio under load. |
| `<writeflushinterval>` | Milliseconds between flushes to disk. |
| `<recorddb …>` | Optional MySQL/MariaDB index. `enabled` is an attribute; the database and its schema must already exist. |

Recording is also togglable at runtime with the `record` keyboard action.

### 6.11 `<radio>` (SA818 transceiver module)

Controls an SA818/DRA818 VHF/UHF module over serial — talkkonnect as an RF gateway.

```xml
<radio enabled="true">
  <connectchannelid>1</connectchannelid>
  <sa818 enabled="true">
    <enabled>true</enabled>
    <serial enabled="true">
      <port>/dev/ttyS0</port>
      <baud>9600</baud>
      <stopbits>1</stopbits>
      <databits>8</databits>
    </serial>
    <channels>
      <channel id="1" name="Channel 1" enabled="true">
        <bandwidth>12500</bandwidth>
        <rxfreq>144.000</rxfreq>
        <txfreq>144.000</txfreq>
        <squelch>3</squelch>
        <ctcsstone>0</ctcsstone>
        <dcstone>0</dcstone>
        <predeemph>0</predeemph>
        <highpass>0</highpass>
        <lowpass>0</lowpass>
        <volume>5</volume>
        <txpower>H</txpower>
      </channel>
    </channels>
  </sa818>
</radio>
```

| Tag | Notes |
| --- | --- |
| `<connectchannelid>` | The `id` of the channel programmed at startup. |
| `<sa818 enabled>` (attr) and `<enabled>` (element) | Two separate flags: the attribute gates the section, the element maps to the module's power-down control. |
| `<serial>` | Module UART: `port`, `baud` (usually 9600), `stopbits`, `databits`. |
| `<bandwidth>` | Passed straight through as the *band* field of the module's `AT+DMOSETGROUP` command, which is defined as `0` (12.5 kHz) or `1` (25 kHz). The sample configs' `12500` / `25000` are not values that command accepts. |
| `<rxfreq>` / `<txfreq>` | MHz, sent with 4 decimal places. Equal values = simplex; different = repeater split. |
| `<squelch>` | Squelch level, sent verbatim to `AT+DMOSETGROUP` (`0`–`8` on SA818-class modules). |
| `<ctcsstone>` | CTCSS tone index, `0` = off. |
| `<dcstone>` | DCS code, `0` = off. |
| `<predeemph>`, `<highpass>`, `<lowpass>` | Audio filters, `0` = off, `1` = on (`AT+SETFILTER`). |
| `<volume>` | Module output volume, sent via `AT+DMOSETVOLUME` (`1`–`8`). |
| `<txpower>` | `H` or `L`. Parsed and logged only — no command is sent for it. |

**Transmit only on frequencies you are licensed for.**

### 6.12 `<analogrelays>`

Drives relay pins when traffic appears on specific channels — for PA zones and sirens.

```xml
<analogrelays enabled="true">
  <zones>
    <zone enabled="true" name="Zone1" listenchannel="Root/Alerts">
      <pins>
        <name>analogrelay1</name>
        <name>analogrelay2</name>
      </pins>
    </zone>
  </zones>
</analogrelays>
```

| Attribute / Tag | Notes |
| --- | --- |
| `name` (attr) | Zone label for logs. |
| `listenchannel` (attr) | Channel whose traffic activates this zone. Matched against the talker's channel exactly. |
| `<pins><name>` | Output pin names from [`<pins>`](#pins) to switch on while the channel is active. Use `analogrelay1` / `analogrelay2` — these are the relay names the pin validator accepts. |

Requires `targetboard="rpi"`; on any other target the zones are not created.

---

## 7. `<global><multimedia>`

Named announcement profiles, playable by id from the HTTP API
(`?command=announcement&mediaid=<id>`), MQTT (`announcement <id>`) or a schedule.

```xml
<multimedia>
  <id value="main_announcement" enabled="true">
    <params>
      <announcementtone file="/…/announcement-01.wav" volume="10" blocking="true" enabled="true"/>
      <localplay>true</localplay>
      <gpio name="multimedia_active_led" enabled="true"/>
      <predelay value="1" enabled="true"/>
      <postdelay value="1" enabled="true"/>
      <playintostream>true</playintostream>
      <streamvolume>50</streamvolume>
      <voicetarget id="2">true</voicetarget>
      <multicast>false</multicast>
    </params>
    <schedule intervalsecs="3600" enabled="false"/>
    <media>
      <source name="1st-song" file="/…/announcement-01.wav" volume="10"
              duration="0" offset="0" loop="1" blocking="true" enabled="true"/>
    </media>
  </id>
</multimedia>
```

| Tag | Notes |
| --- | --- |
| `value` (attr) | Profile id used by the API. **Must be non-empty**, or the profile is disabled with a warning. |
| `<announcementtone …>` | Attention tone before the media. A missing local file disables the tone (but not the profile). |
| `<localplay>` | Play on the local speaker. |
| `<playintostream>` / `<streamvolume>` | Transmit into the Mumble channel at `0–100`. |
| `<multicast>` | Also send to the RTP multicast group. Warns if [`<multicast>`](#515-multicast) is not enabled, and needs `ffmpeg` on `PATH`. |
| `<voicetarget id="N">true</voicetarget>` | Route the into-stream announcement at voice target `N` (**1–31**) instead of the plain channel. Omit `id` to follow whichever target is currently active. An out-of-range `id` is reset to `0` with a warning. The legacy empty form `<voicetarget/>` still parses. |
| `<gpio name= enabled=>` | Drive an output pin while the profile plays. |
| `<predelay value= enabled=>` / `<postdelay …>` | Delay in **seconds** before/after playback. |
| `<schedule intervalsecs= enabled=>` | Repeat the profile every *n* seconds. Enabled with `intervalsecs <= 0` is disabled with a warning. |
| `<media><source>` | Playlist entries, played in document order. |

`<source>` attributes: `name` (label), `file` (path or `http`/`https`/`rtsp` URL),
`volume` (`0–100`), `duration` (seconds, `0` = whole file), `offset` (seconds to skip),
`loop` (repeat count, **clamped to a maximum of 3**), `blocking`, `enabled`.

**A profile must have at least one of `<localplay>`, `<playintostream>` or
`<multicast>` set to `true`** — with none of them it is disabled with a warning, since it
would have nowhere to play.

---

## 8. `<global><Radio>` (internet streaming radio)

Plays internet audio streams through `ffmpeg` to ALSA. Note the **capitalised tag
names** in this section — they differ from the rest of the file and are case-sensitive.

```xml
<Radio>
  <Enabled>true</Enabled>
  <AutoResumeDelay>15</AutoResumeDelay>
  <InterruptionMode>pause</InterruptionMode>
  <MasterVolume>50</MasterVolume>
  <DuckVolumePercent>10</DuckVolumePercent>
  <AlsaDevice>default</AlsaDevice>
  <FFmpegPath>/usr/bin/ffmpeg</FFmpegPath>
  <YoutubeMusicPlayback>false</YoutubeMusicPlayback>
  <YtDlpPath>/usr/bin/yt-dlp</YtDlpPath>
  <YtDlpFormat>bestaudio/best</YtDlpFormat>
  <AnnounceStationTTS>false</AnnounceStationTTS>
  <StreamRetrySecs>5</StreamRetrySecs>
  <Stations>
    <Station>
      <Name>News Radio</Name>
      <URL>http://stream.wbez.org/wbez128.mp3</URL>
      <Volume>80</Volume>
      <Backend>http</Backend>
    </Station>
  </Stations>
</Radio>
```

| Tag | Valid values | Default | Notes |
| --- | --- | --- | --- |
| `<Enabled>` | bool | — | Element, not an attribute. |
| `<InterruptionMode>` | `stop`, `pause`, `duck` | `stop` | What happens to the radio when Mumble audio arrives. `duck` lowers it to `<DuckVolumePercent>`. Invalid values warn and fall back to `stop`. |
| `<AutoResumeDelay>` | seconds | `15` | Silence after which the radio resumes. `0` or negative becomes 15. |
| `<MasterVolume>` | 1–100 | `50` | Out-of-range warns and resets to 50. |
| `<DuckVolumePercent>` | 1–100 | `10` | Volume while ducked. |
| `<AlsaDevice>` | ALSA device | `default` | e.g. `plughw:1,0`. |
| `<FFmpegPath>` | path | `/usr/bin/ffmpeg` | Required — this is the playback engine. |
| `<StreamRetrySecs>` | seconds | `5` | Reconnect delay after a stream drop. `0` or negative becomes 5. |
| `<AnnounceStationTTS>` | bool | `false` | Speak the station name on change (uses [`<ttsmessages>`](#511-ttsmessages)). |
| `<YoutubeMusicPlayback>` | bool | `false` | Enable YouTube/YouTube Music URLs via `yt-dlp`. |
| `<YtDlpPath>` | path or command | `/usr/bin/yt-dlp` | If `YoutubeMusicPlayback` is on and this is not found on disk or `PATH`, you get a warning. |
| `<YtDlpFormat>` | yt-dlp format string | `bestaudio/best` | |

Each `<Station>`:

| Tag | Notes |
| --- | --- |
| `<Name>` | Label shown on screen and spoken by `AnnounceStationTTS`. |
| `<URL>` | Stream URL. For YouTube Music use a specific track/watch URL, not the site home page. |
| `<Volume>` | Per-station volume `0–100`. |
| `<Backend>` | Empty or `auto` decides from `YoutubeMusicPlayback` plus URL detection; `youtube` forces `yt-dlp`; `http`, `direct` or `ffmpeg` force plain `ffmpeg -i`. |

Enabling the section with no stations warns and falls back to built-in demo stations.
Control at runtime with `radiotoggle`, `radionext`, `radioprev`, `radiovolup`,
`radiovoldown`.

---

## 9. Live reload — what can change at runtime

`Ctrl-B` (or a save from the `/config` web editor) re-reads the file but applies only an
allow-list of settings. Everything else keeps its startup value until you restart.

**Applied on live reload:**

* `<settings>`: `loglevel`, `cancellablestream`, `streamsendmessage`, `repeattxtimes`,
  `repeattxdelay`, `simplexwithmute`
* `<channelscan>`, `<beacon>`, `<tts>`, `<sounds>` (event sounds are re-preloaded)
* `<remotecontrol>`: `<http enabled>` and its `<command>` list, `<uistatus>`,
  `<networkacl>`, MQTT `<commands>`
* `<printvariables>`, `<ttsmessages>`, `<ignoreuser>`
* `<multicast>` (re-applied to the running sender)
* `<panicfunction>`, `<keyboard>` commands
* `<multimedia>`, `<Radio>` (internet radio is reconfigured)

**Requires a restart:** accounts and server list, all audio device settings, `<logging>`
and `<logfilenameandpath>`, `<singleinstance>`, `<remotesshconsole>`,
`<autoprovisioning>`, `<smtp>`, `<memorychannels>`, `<presetvoicetargets>`, MQTT
`<settings>`, and every `<hardware>` section (GPIO, LCD, OLED, GPS, Traccar, USB
keyboard devices, audio recording, SA818 radio, analog relays).

The bottom-CLI `cfg` commands offer finer-grained runtime edits:

```
cfg keys                                       # list settable paths
cfg list                                       # show current values
cfg set global.software.settings.loglevel debug
cfg set accounts.account.0.username mycall
cfg save                                       # write back to talkkonnect.xml
cfg restart                                    # re-exec talkkonnect
```

Some runtime data (notably the account slices built at startup) only takes full effect
after `cfg restart`.

---

## 10. Startup validation and what it does to your config

`CheckConfigSanity` runs on every load. Its output is worth reading on first boot:

```
info: Starting XML Configuration Sanity and Logical Checks
warn: Config Error [Section GPIO] Enabled GPIO Name listening Pin Number 14 Invalid Name
warn: Non-Critical Errors Found In talkkonnect.xml config file please fix errors …
```

* `warn:` — the value was clamped or the feature was switched off. talkkonnect keeps
  running, possibly without the feature you configured.
* `alert:` — fatal. talkkonnect stops. The two fatal cases are **no default/enabled
  account** and an unusable `<nextserverindex>`.
* `info: Finished XML Configuration Sanity and Logical Checks Without Any
  Alerts/Errors/Warnings` — a clean file.

Summary of automatic corrections:

| Condition | Result |
| --- | --- |
| `<outputdeviceshort>`, `<outputvolcontroldevice>`, `<outputmutecontroldevice>` empty | Copied from `<outputdevice>` |
| `<logfilenameandpath>` empty and logging ≠ `screen` | First writable of config dir, CWD, `/var/log`, `/tmp` |
| `<voiceactivitytimermsecs>` < 200 | Set to 200 |
| `<volupstep>` / `<voldownstep>` = 0 | Set to `+1` / `-1` |
| `<nextserverindex>` > number of default accounts | Reset to 0 (fatal if there are no accounts) |
| Channel scan dwell/hang < 500 ms | 500 ms used |
| Sound file missing (and not an `http`/`rtsp` URL) | That sound disabled |
| Sound `volume="0"` | That sound disabled |
| Media `loop` > 3 | Clamped to 3 |
| Multimedia profile with no output destination | Profile disabled |
| Multicast group / port / codec / packetms / ttl / volume / hangover invalid | Clamped to defaults |
| Beacon, SMTP, SSH console, autoprovisioning, MQTT missing required children | Feature disabled |
| GPIO pin with bad direction / device / name / number / chipid | That pin disabled |
| LCD interface type or required pins invalid | LCD disabled |
| GPS port missing, bad baud, both parities, zero stop/data bits | GPS disabled |
| `<ignoreuserregex>` shorter than 4 chars | Ignore-user disabled |
| Internet radio `InterruptionMode` / `MasterVolume` invalid | Reset to `stop` / `50` |
| Keyboard action not recognised, or scan id 0 / >255 | That binding disabled |
| `<memorychannels>` / `<presetvoicetargets>` with an unrecognised `gpioname` | The whole section disabled |

---

## 11. Known quirks and traps

These are real behaviours of the current code. Knowing them saves hours.

1. **Unknown tags are silently ignored.** Misspell a tag and the value is never read,
   with no warning. `<oledttf>` and `<ttykeyboard>` in the shipped configs fall into
   this category — they parse fine and do nothing.

2. **`<tts>` action names are case-sensitive and must match exactly.** The shipped
   config contains `muteSpeaker`, `unmuteSpeaker` and `currentvolumelevel`; the code asks
   for `mutespeaker`, `unmutespeaker`, `currentrxvolumelevel` and
   `currenttxvolumelevel`. Prompts with the wrong spelling never play. Several other
   entries in the shipped config (`participants`, `leftjoinedchannel`, `message`,
   `listentochannelson/off`, `printxmlconfig`) are not requested by any code path.

3. **`<sounds><input>` event names need the `io`/`usb` prefix.** The shipped config uses
   short names like `txpttstart` and `volup`; the code asks for `iotxpttstart` and
   `iovolup`. Only `txtogglestart` matches without a prefix. Use the list in
   [§5.8](#58-sounds).

4. **`<sound event="alert">`** appears in the shipped config but is not requested
   anywhere. The panic alert sound comes from
   `<panicfunction><filenameandpath>` instead.

5. **Some working GPIO names are rejected by the validator.** `listening`, `tracking`,
   `mqtt0`, `mqtt1` and the five `internetradio*` names have handlers in `gpio.go` but
   are not in the validator's allow-list, so enabling such a pin logs
   `Invalid Name` and disables it. `device="radiomodule"` is likewise not a valid
   device value.

6. **`<listening>` and `<comment><commentbuttonpin>` are ignored.** Button binding is by
   the pin's `name` attribute under `<pins>` (`comment`, `listening`), never by these
   fields.

7. **The LCD backlight timer has an inverted-looking guard.** It is disabled unless
   *both* `<oled enabled>` and `<lcd enabled>` are true, and it is also forced off when
   `targetboard` is not `rpi`.

8. **`<ttsmessages><predelay>` / `<postdelay>` produce no useful delay.** The value is
   treated as nanoseconds there. The `<multimedia>` equivalents are correctly
   interpreted as seconds — use those when you need a real gap.

9. **`<max7219>` children must be attributes.** The parser declares
   `max7219cascaded`, `spibus`, `spidevice` and `brightness` as attributes; writing them
   as child elements (as some sample configs do) leaves them at zero.

10. **HTTP and MQTT action names differ.** MQTT uses `muteunmute` (with a
    `mute`/`unmute`/`toggle` argument) where HTTP uses three separate actions, and MQTT
    has no `ttsannouncement`, `showversion`, `joinchannel`, `whisperuser`, `setrxvolume`
    or `listapi`. Declaring an MQTT command that has no handler passes the "is it
    defined?" check and then fails at dispatch time.

11. **An HTTP command must be declared in the XML *and* be a built-in.** Adding
    `radiovolup` / `radiovoldown` `<command>` entries is needed to reach those handlers —
    the shipped config omits them.

12. **`<tokens enabled="…">` is read from the `<account>` element, not `<tokens>`.** The
    attribute is declared on the account struct, so the flag on `<tokens>` has no effect.
    Tokens themselves are always sent.

13. **`<hardware><radio>` and `<global><Radio>` are different features.** The first is
    the SA818 RF module; the second is internet streaming radio with capitalised tag
    names.

14. **The MQTT `relay` command does not work as written.** It accepts a relay number of
    `1` or `2` and drives a pin named `relay1` / `relay2` — but the pin validator only
    allows `relay0`, so such a pin is disabled at startup. Separately, the handler reads
    the fourth word of a three-word payload, so the argument lookup is out of range.
    Use `<analogrelays>` or a GPIO output driven by another action instead.

---

## 12. Minimal working config

Enough to connect and talk, with nothing hardware-specific:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<document>
  <accounts>
    <account name="my-server" default="true">
      <serverandport>mumble.example.com:64738</serverandport>
      <username>my-device</username>
      <password/>
      <insecure>true</insecure>
      <register>false</register>
      <certificate/>
      <channel>Root</channel>
      <ident>MyDevice</ident>
      <listentochannels/>
      <tokens enabled="false"/>
      <voicetargets/>
    </account>
  </accounts>
  <global>
    <software>
      <settings>
        <singleinstance>true</singleinstance>
        <outputdevice>Speaker</outputdevice>
        <inputdevice>CAPTURE</inputdevice>
        <logfilenameandpath>/var/log/talkkonnect.log</logfilenameandpath>
        <logging>screen</logging>
        <loglevel>info</loglevel>
        <simplexwithmute>true</simplexwithmute>
        <nextserverindex>0</nextserverindex>
      </settings>
      <remotecontrol>
        <http enabled="true" listenport="8080">
          <command action="displaymenu" message="Display Menu" enabled="true"/>
          <command action="channelup" message="Channel Up" enabled="true"/>
          <command action="channeldown" message="Channel Down" enabled="true"/>
          <command action="starttransmitting" message="Start TX" enabled="true"/>
          <command action="stoptransmitting" message="Stop TX" enabled="true"/>
          <command action="listapi" message="List API" enabled="true"/>
        </http>
        <networkacl enabled="true">
          <network cidr="127.0.0.1/32"/>
          <network cidr="192.168.0.0/16"/>
        </networkacl>
        <uistatus enabled="false"/>
      </remotecontrol>
    </software>
    <hardware targetboard="pc">
      <voiceactivitytimermsecs>200</voiceactivitytimermsecs>
      <lcd enabled="false"/>
      <oled enabled="false"/>
      <gps enabled="false"/>
      <usbkeyboard enabled="false"/>
    </hardware>
  </global>
</document>
```

Build up from here one section at a time, watching the sanity-check output after each
change. Complete worked examples for specific hardware live in
[sample-configs/](../sample-configs/).
