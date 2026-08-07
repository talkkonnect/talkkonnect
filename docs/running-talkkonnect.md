# Configuring and Running talKKonnect

This page covers getting a built talKKonnect running: starting it, driving it from the terminal,
setting up audio and the optional LCD/OLED screens, and then a walkthrough of every section of
`talkkonnect.xml`.

For building from source see [Getting Started](./getting-started.md). For the exhaustive per-tag
reference — valid values, defaults, and exactly what the sanity checker does to each field — see the
[talkkonnect.xml Configuration Manual](./CONFIGURATION.md), which this page links into section by
section. For the HTTP API request/response format see [api.md](./api.md).

---

## Contents

* [Sample configurations](#sample-configurations)
* [Starting talKKonnect](#starting-talkkonnect)
* [The welcome banner](#the-welcome-banner)
* [Driving talKKonnect from the terminal](#driving-talkkonnect-from-the-terminal)
* [Audio configuration](#audio-configuration)
* [LCD and OLED screen installation](#lcd-and-oled-screen-installation)
* [The talkkonnect.xml configuration File tags and their meaning](#the-talkkonnectxml-configuration-file-tags-and-their-meaning)

---

## Sample configurations

Working configurations live in [sample-configs/](https://github.com/talkkonnect/talkkonnect/tree/main/sample-configs).
Start from the one closest to your build and cut it down rather than writing a file from scratch:

| File | For |
| --- | --- |
| `talkkonnect-current-sample-config.xml` | The most current file, with every tag present — the one to diff against |
| `talkkonnect-v4-pc-x86.xml` | PC / VM / server, no GPIO (`targetboard="pc"`) |
| `talkkonnect-v4-rpi4.xml` | Raspberry Pi 4 with GPIO, LCD/OLED |
| `talkkonnect-version2-usb-gpio-example.xml` | USB sound card plus GPIO buttons and LEDs |
| `talkkonnect-version2-respeaker.xml` | ReSpeaker HAT with LED strip |
| `talkkonnect-version2-opi-internal-soundcard.xml` | Orange Pi with the internal sound card |
| `talkkonnect-mcp23017.xml` | MCP23017 I²C GPIO expander |
| `talkkonnect.tkv1pcb` | The talKKonnect v1 PCB pinout |

There is also a [minimal working config](#minimal-working-config) at the end of this page — enough to
connect and talk, with nothing hardware-specific.

---

## Starting talKKonnect

`make build` puts the binary in `dist/talkkonnect`; `make install` puts it in
`/usr/local/bin/talkkonnect`. Many existing installs (and the shipped systemd unit) use
`/home/talkkonnect/bin/talkkonnect`.

```bash
./talkkonnect -config /home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml
```

Command line flags:

| Flag | Default | Meaning |
| --- | --- | --- |
| `-config` | `/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml` | Full path to the XML config. This is the **only** file talKKonnect reads |
| `-serverindex` | `0` | Start on account index *n* of the `default="true"` accounts |
| `-daemon` | `false` | Detach and run headless, writing `talkkonnect-daemon.pid` and `talkkonnect-daemon.log` in the working directory |
| `-cpuprofile`, `-memprofile` | — | Write a pprof profile to the named file |
| `-help` | — | Usage summary |

### Running unattended

talKKonnect has native daemon mode, so `screen` is no longer needed. Use the unit in
[`conf/systemd/talkkonnect.service`](../conf/systemd/talkkonnect.service) — fill in `User`, `Group`
and `WorkingDirectory`, and add `-config` to `ExecStart` if your config is not at the default path:

```ini
[Service]
Type=forking
User=talkkonnect
Group=talkkonnect
WorkingDirectory=/home/talkkonnect
ExecStart=/home/talkkonnect/bin/talkkonnect --daemon
Restart=always
RestartSec=10
```

```bash
sudo cp conf/systemd/talkkonnect.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now talkkonnect
journalctl -u talkkonnect -f
```

In daemon mode there is no terminal, so the bottom CLI is skipped automatically. To get a console on
a daemonised unit, enable the built-in SSH console — see
[`<remotesshconsole>`](#remotesshconsole) — and `ssh -p 9999 user@device`.

If you prefer `screen` on a Raspberry Pi, `screen -dmS talkkonnect /home/talkkonnect/bin/talkkonnect &`
still works; reattach with `screen -r` and detach with `Ctrl-A-D`.

---

## The welcome banner

```
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
```

> The `<Del>` line is a leftover from the old function-key interface. Since version 4 the menu is
> shown by typing `?` (or `help` / `menu`) at the bottom prompt.

The banner is followed by the startup log: the config sanity check, the ALSA/OpenAL devices found,
the server connection, and the channel list. Read the `warn:` lines on first boot — they tell you
which parts of your config were clamped or switched off. See
[Startup validation](#startup-validation-and-what-it-does-to-your-config).

---

## Driving talKKonnect from the terminal

When talKKonnect runs in a terminal it pins a command prompt to the bottom row while the log scrolls
above it. Single keys run the common actions; typed commands do everything else, with Tab completion
throughout. The same prompt is served over SSH when
[`<remotesshconsole>`](#remotesshconsole) is enabled.

Type `?`, `help` or `menu` at the prompt for this banner:

```
┌──────────────────────────────────────────────────────────────┐
│     _ __ ___   __ _(_)_ __    _ __ ___   ___ _ __  _   _     │
│    | '_ ` _ \ / _` | | '_ \  | '_ ` _ \ / _ \ '_ \| | | |    │
│    | | | | | | (_| | | | | | | | | | | |  __/ | | | |_| |    │
│    |_| |_| |_|\__,_|_|_| |_| |_| |_| |_|\___|_| |_|\__,_|    │
├─────────────────────────────┬────────────────────────────────┤
│ <1> to Display this Menu    | <Ctrl-C> to Quit talkkonnect   │
├─────────────────────────────┼────────────────────────────────┤
│ <2> Channel UP (+)          │ <3>  Channel Down (-)          │
│ <4> Mute/Unmute Speaker     │ <5>  Digital Volume Up (+)     │
│ <6> Digital Volume Down (-) │ <7>  Start Transmitting        │
│ <8> Stop Transmitting       │ <9> List Online Users          │
│ <0> Show Uptime             │                                │
├─────────────────────────────┼────────────────────────────────┤
│ <a> List API Commands       │<b> Playback/Stop Stream        │
│ <d> Dump XML Config         │<e> Send Email                  │
│ <g> GPS Position            │<h> XML Config Checker          │
│ <l> Clear Screen            │<m> Radio Channel (+)           │
│ <n> Radio Channel (-)       │<o> Ping Servers                │
│ <p> Panic Simulation        │<q> Repeat TX Loop Test         │
│ <r> Scan Channels On/Off    │<s> Thanks/Acknowledge          │
│ <t> Show Uptime             │<u> Display Version             │
│ <v> Online Radio On/Off     │<w> Dump XML Config             │
│ <x> Next Server             │<z> Next Server                 │
├─────────────────────────────┴────────────────────────────────┤
│ Voice Targets (whisper/shout), typed in the CLI:             │
│  vt list                     list configured + active target │
│  vt set <id> / vt clear      activate target <id> / target 0 │
│  vt next / vt prev           step through configured targets │
│  vt whisper <user>           whisper to one online user      │
│  vt add / vt help            add a target / full usage       │
├──────────────────────────────────────────────────────────────┤
│ Multicast (RTP to IP speakers), typed in the CLI:            │
│  mc status                   destination, filters, live mix  │
│  mc on / mc off / mc toggle  start or stop multicast output  │
├──────────────────────────────────────────────────────────────┤
│  Visit us at www.talkkonnect.com and github.com/talkkonnect  │
│  Thanks to Global Coders Co., Ltd. for their sponsorship     │
└──────────────────────────────────────────────────────────────┘
```

Two labels in that banner are wrong and are worth knowing: **`<1>` does nothing** — use `?`, `help`
or `menu` — and **`<x>` is *previous* server**, not next (`<z>` is next).

### Single keys

| Key | Action |
| --- | --- |
| `2` `3` | Channel up / down |
| `4` | Mute/unmute speaker (toggle) |
| `5` `6` | RX digital volume up / down |
| `7` `8` | Start / stop transmitting |
| `9` | List online users in the current channel |
| `0` `t` | Show uptime |
| `a` | Write the enabled HTTP API command list to the log |
| `b` | Playback / stop the configured stream into the channel |
| `d` `w` | Dump the whole XML config to the log |
| `e` | Send the SMTP alert email |
| `g` | Request and show the GPS position |
| `h` | Re-run the XML config sanity checker |
| `l` | Clear the terminal, LCD and OLED |
| `m` `n` | SA818 radio module channel up / down |
| `o` | Ping the configured servers |
| `p` | Panic simulation |
| `q` | Repeat TX loop test |
| `r` | Channel scan on / off |
| `s` | Thanks and acknowledgements screen |
| `u` | Show version, and whether a newer release exists |
| `v` | Internet radio on / off |
| `x` `z` | Connect to the previous / next server |

### Typed commands

| Command | What it does |
| --- | --- |
| `?` / `help` / `menu` | Show the banner above |
| `c` / `clear` / `cls` | Clear the screen and repaint the prompt |
| `q` / `quit` / `exit` | Close the CLI, leaving talKKonnect running |
| `...` | Shut talKKonnect down (SIGTERM) |
| `clearhist` | Clear the command history |
| `cfg keys` / `cfg list` / `cfg set <path> <value>` / `cfg save` / `cfg restart` | Read and change any setting in the running config by dotted path, e.g. `cfg set global.software.settings.loglevel debug`. `cfg save` writes it back to `talkkonnect.xml`, `cfg restart` re-execs |
| `vt …` | Voice targets — see [Voice Targets and the vt Command](#voice-targets-and-the-vt-command) |
| `mc …` | Multicast output — see [The Multicast Section](#the-multicast-section-rtp-to-ip-speakers-and-sip-phones) |
| any HTTP API action name | Typed directly, e.g. `announcement main_announcement`, `voicetargetset 3`, `joinchannel Root/Ops`, `showuptime`. These go through the same allow-list as the HTTP API, so the action must be enabled under `<http>` |

Arrow keys walk the history; Tab completes commands, `cfg set` paths, configured voice target ids and
online user names.

Set `TALKKONNECT_NO_BOTTOM_CLI=1` for plain unadorned log output — useful under a test harness or in
a pipeline. The prompt is also skipped automatically when stdout is not a terminal.

See [terminal-cli.md](./terminal-cli.md) for the short version of this, and
[api.md](./api.md) for the same actions over HTTP.

---

## Audio configuration

Configure and test your Linux sound system **before** troubleshooting talKKonnect. talKKonnect works
well with plain ALSA; PulseAudio is not required.

### USB sound cards

Raspberry Pi boards have audio output via the BCM2835 chip but, by design, no audio input — hence a
USB sound card. Cards built on the CM108, CM109, CM119 and CM6206 chips are cheap, common and work
well. Many other single board computers (Orange Pi, for example) have both input and output on board.

Identify a connected card with `lsusb`:

```
Bus 001 Device 004: ID 0d8c:000c C-Media Electronics, Inc. Audio Adapter
```

List playback devices with `aplay -l` and capture devices with `arecord -l`.

Optionally disable the Pi's internal sound so the USB card becomes card 0 — edit `/boot/config.txt`
(`/boot/firmware/config.txt` on Bookworm and later), add:

```
#Disable audio (loads snd_bcm2835)
dtparam=audio=off
```

then save and reboot.

If you keep BCM2835 enabled, the USB card is usually card 1 and you need to tell ALSA to use it.
Either edit `/usr/share/alsa/alsa.conf` globally:

```
#defaults.ctl.card 0
#defaults.pcm.card 0
defaults.ctl.card 1
defaults.pcm.card 1
```

or set it per user in `~/.asoundrc`:

```
pcm.!default {
    type asym
    capture.pcm "mic"
    playback.pcm "speaker"
}
pcm.mic {
    type plug
    slave {
        pcm "hw:1,0"
    }
}
pcm.speaker {
    type plug
    slave {
        pcm "hw:1,0"
    }
}
```

Match the card index to your system (`aplay -l`, `amixer`). If you configured it globally in
`alsa.conf` there is no need for `.asoundrc` as well.

The microphone must be **captured** for talKKonnect to work: run `alsamixer`, select your input
device (Mic or Line In), press the space bar, and check that a red `CAPTURE` label appears.

### Testing

```bash
speaker-test            # you should hear white noise
arecord -f CD | aplay   # you should hear yourself
```

Adjust microphone sensitivity and output gain with `alsamixer` or `amixer` — this takes some trial
and error.

### Telling talKKonnect which devices to use

The mixer control names in `<global><software><settings>` must match what `amixer` calls them on your
card — `Speaker`, `PCM`, `Master`, `Headphone`, `CAPTURE`, `Mic` and so on. Speaker muting on PTT
(`<simplexwithmute>`) works through `<outputmutecontroldevice>`, so that one in particular has to be
right.

```xml
<outputdevice>Speaker</outputdevice>
<outputdeviceshort>Spkr</outputdeviceshort>
<outputvolcontroldevice>Speaker</outputvolcontroldevice>
<outputmutecontroldevice>Speaker</outputmutecontroldevice>
<inputdevice>CAPTURE</inputdevice>
<openalinputdevice></openalinputdevice>
<openaloutputdevice></openaloutputdevice>
<localplaybackdevice></localplaybackdevice>
```

If you are unsure, set all four `outputdevice*` tags to whatever your default output control is
called. The Mumble audio path itself goes through OpenAL rather than ALSA mixer controls — leave
`<openalinputdevice>` and `<openaloutputdevice>` empty for the system default, or set
`<loglevel>debug</loglevel>` and talKKonnect will log every OpenAL capture and playback device it can
see at startup, which is the quickest way to get the exact names:

```
info: OpenAL capture devices: [ALSA Default]
info: OpenAL playback devices: [ALSA Default]
```

Full detail: [CONFIGURATION.md §5.1](./CONFIGURATION.md#51-settings).

### External tools

Some features shell out, so these need to be installed and on `PATH`:

| Tool | Needed for |
| --- | --- |
| `ffmpeg` | Repeater tone generation, `<multimedia>` announcements, multicast paging, internet radio |
| `ffplay` | Local playback of `<multimedia>` profiles and generated tones |
| `aplay` | `<sounds><input>` button feedback sounds |
| `yt-dlp` | Internet radio with `<YoutubeMusicPlayback>true</YoutubeMusicPlayback>` |

---

## LCD and OLED screen installation

Both screen types need `targetboard="rpi"` and the I²C interface enabled. Follow step 1 of
[enabling I2C](https://www.raspberrypi-spy.co.uk/2014/11/enabling-the-i2c-interface-on-the-raspberry-pi/),
then install the detection tool as root:

```bash
apt-get install -y i2c-tools
i2cdetect -y 1
```

`i2cdetect` reports the address in **hex**, and talkkonnect.xml wants it in **decimal**. Convert it:

| `i2cdetect` shows | Put in the XML | Tag |
| --- | --- | --- |
| `3c` | `60` | `<oleddefaulti2caddress>60</oleddefaulti2caddress>` |
| `3f` | `63` | `<lcdi2caddress>63</lcdi2caddress>` |

Other things worth knowing:

* GPIO pins **2 and 3 are the I²C bus** — they cannot be used for anything else once an I²C screen or
  an MCP23017 expander is enabled. talKKonnect warns if you try.
* For OLEDs, `<oledstartcolumn>` is `0` for 0.96 inch panels and `1` for 1.3 inch panels. Getting it
  wrong leaves garbage down one edge of the screen.
* If the OLED does not answer at startup, talKKonnect logs `Cannot Communicate with OLED` and carries
  on with the screen disabled.

Full tag detail: [`<lcd>`](#the-lcd-section-hd44780-character-lcd) and
[`<oled>`](#the-oled-section-i2c-oled) below,
[CONFIGURATION.md §6.4](./CONFIGURATION.md#64-lcd) and
[§6.5](./CONFIGURATION.md#65-oled).

---

## The talkkonnect.xml configuration File tags and their meaning

`talkkonnect.xml` is the single file that drives talKKonnect. What follows is a walkthrough of every
section, in file order, with the tags that matter and the traps that cost people time. For the
complete per-tag tables — every valid value, every default, and exactly what the sanity checker does
to a bad value — see the [Configuration Manual](./CONFIGURATION.md); each subsection below links to
its counterpart there.

### How the file is loaded

1. The XML is unmarshalled into the config struct.
2. `CheckConfigSanity()` runs and **mutates the config in memory** — bad values are clamped and
   misconfigured features are switched off, each with a `warn:` line. Anything logged as `alert:` is
   fatal and stops talKKonnect.
3. Derived state is built: the account list, the USB key map, the GPIO memory-channel and
   voice-target maps, the event-sound preload, the multicast sender and internet radio.

Ways to reload it:

| Method | Effect |
| --- | --- |
| `h` at the prompt | Re-run the sanity checker against the in-memory config, changing nothing |
| `Ctrl-B` on the terminal | Live reload — re-read the file and apply the reloadable subset ([see below](#live-reload--what-can-change-at-runtime)) |
| `http://<host>:<port>/config` | Built-in web editor: saves the posted XML to disk, then live-reloads it |
| `cfg set <path> <value>` | Change one leaf value at runtime; `cfg save` writes it back, `cfg restart` re-execs |
| Restart the process | The only way to apply the hardware sections (GPIO, LCD/OLED, GPS, keyboard devices) |

### Conventions you need to know

**`enabled` is an attribute, not an element.** Almost every feature block is switched on with an
attribute on its opening tag:

```xml
<beacon enabled="true">…</beacon>
```

A handful of settings use elements instead — `<insecure>true</insecure>`,
`<localplay>true</localplay>`, `<Enabled>true</Enabled>` under `<Radio>`. Where it matters, it is
called out below.

**Booleans** are `true` / `false`. Anything else parses as `false`.

**Unknown tags are silently ignored.** Go's XML decoder does not error on elements it does not
recognise, so a typo in a tag name means the value is simply never read, with no warning at all. If a
setting appears to do nothing, check its spelling against [CONFIGURATION.md](./CONFIGURATION.md)
first.

**Durations** are written as bare integers and the unit is applied in code:

| Tag | Unit |
| --- | --- |
| `<streamonstartafter>`, `<txonstartafter>`, `<repeattxdelay>` | seconds |
| `<voiceactivitytimermsecs>` | milliseconds |
| `<predelay value="…">` / `<postdelay value="…">` under `<multimedia>` | seconds |
| `<predelay>` / `<postdelay>` under `<ttsmessages>` | **nanoseconds** — effectively no delay, see [Known quirks](#known-quirks-and-traps) |
| anything named `*msecs` / `*secs` | as the name says |

**Volumes** come in three flavours depending on where they appear:

* `0–100` percent integers — event sounds, beacon local play, multimedia, internet radio
* `0.0–1.0` float gains for into-stream playback — `<beaconvolumeintostream>`,
  `<playvolumeintostream>`, `<panicfunction><volume>`
* `1–200` percent for multicast, where above 100 amplifies

**File paths must be absolute.** Local sound files are checked at startup and a missing file disables
that sound with a warning. Values starting `http`, `https` or `rtsp` are treated as URLs and not
existence-checked.

### Document skeleton

```xml
<?xml version="1.0" encoding="UTF-8"?>
<document>
  <accounts>
    <account name="…" default="true">…</account>
  </accounts>
  <global>
    <software>
      <settings/>          <channelscan/>        <remotesshconsole/>
      <autoprovisioning/>  <beacon/>             <tts/>
      <smtp/>              <sounds/>             <remotecontrol/>
      <printvariables/>    <ttsmessages/>        <ignoreuser/>
      <memorychannels/>    <presetvoicetargets/> <multicast/>
    </software>
    <hardware targetboard="pc">
      <ledstripenabled/>   <gpiooffset/>         <voiceactivitytimermsecs/>
      <io/>                <heartbeat/>          <comment/>
      <listening/>         <lcd/>                <oled/>
      <gps/>               <traccar/>            <panicfunction/>
      <usbkeyboard/>       <keyboard/>           <audiorecordfunction/>
      <radio/>             <analogrelays/>
    </hardware>
    <multimedia>…</multimedia>
    <Radio>…</Radio>
  </global>
</document>
```

Note the **two different radio sections**: `<hardware><radio>` is the SA818 RF transceiver module,
`<global><Radio>` (capital R) is internet streaming radio. They are unrelated features.

---

### The Accounts Section

One `<account>` per Mumble server. The accounts marked `default="true"` form the connection list;
talKKonnect starts on the one selected by `-serverindex` or `<nextserverindex>`, and `z` / `x` (or the
`connnextserver` / `previousserver` actions) cycle through them. Accounts with `default="false"` are
parsed but never used — including their voice targets.

```xml
<account name="talkkonnect-community" default="true">
  <serverandport>mumble.talkkonnect.com:64738</serverandport>
  <username>my-device</username>
  <password/>
  <insecure>true</insecure>
  <register>false</register>
  <certificate>/home/talkkonnect/mumble.pem</certificate>
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

| Tag | Notes |
| --- | --- |
| `name` (attr) | Label used in logs, on screen and as the CLI prompt. A missing name warns |
| `default` (attr) | `true` adds the account to the active list. **At least one is required** — zero is a fatal `alert:` |
| `<serverandport>` | Required, `host:port`. Mumble's default port is 64738; the community server is `mumble.talkkonnect.com:64738` |
| `<username>` | Identifies you on the server. Empty auto-generates `talkkonnect-<MAC>`, or a random suffix if no MAC is readable |
| `<password>` | Server password. Leave as `<password/>` if the server has none |
| `<insecure>` | Element, not an attribute. `true` skips TLS certificate verification — needed for self-signed servers |
| `<register>` | Request registration on the server |
| `<certificate>` | Absolute path to a combined cert+key PEM for certificate auth. Set-but-missing warns at startup and is fatal at connect time |
| `<channel>` | Channel to join after connecting; sub-channels work as a path. Leave empty to stay in Root |
| `<ident>` | Device identity string, sent in panic messages and shown in status output |
| `<listentochannels><channel>` | Extra channels to monitor without joining. Activated at boot by `<listentochannelsonstart>`, or on demand with `listeningstart` |
| `<tokens><token>` | Mumble access tokens sent on connect, for token-protected channels |
| `<voicetargets>` | Preset whisper/shout targets, below |

Full detail: [CONFIGURATION.md §4](./CONFIGURATION.md#4-accounts).

#### Voice Targets and the vt Command

A voice target is a Mumble whisper or shout destination — a set of users, or a channel — that your
audio goes to instead of the channel you are joined to. Targets are numbered **1 to 31** (0 means
normal channel speech) and live under `<voicetargets>` inside an account. Only the account
talKKonnect is actually using is consulted, so a target added to a `default="false"` account is never
sent.

```xml
<voicetargets>
    <id value="1">
        <iscurrent>true</iscurrent>
        <users>
            <user>zoran-laptop</user>
        </users>
    </id>
    <id value="3">
        <channels>
            <channel>
                <name>TEST1</name>
                <recursive>true</recursive>
                <links>true</links>
                <group>all</group>
            </channel>
        </channels>
    </id>
</voicetargets>
```

* a target may list several `<user>` entries, several `<channel>` entries, or both
* multiple `<id>` blocks with the same `value` are merged
* `<iscurrent>` marks the slot to select at startup
* `recursive` includes the sub-channels of the named channel, `links` includes channels linked to it,
  and `group` restricts the shout to one ACL group — leave it empty for everyone
* select a target with the `voicetargetset <id>` action (bottom CLI, SSH console, HTTP API, MQTT, GPIO
  preset buttons, rotary encoder) and go back to normal channel speech with `voicetargetset 0`
* a `<multimedia>` announcement can be aimed at a target with `<voicetarget id="N">true</voicetarget>`

Rather than editing the XML by hand, the `vt` command in the bottom CLI and the SSH console lists
targets, selects them and appends new ones:

```
vt list                                          show every configured target
vt set <id>                                      activate a configured target
vt clear                                         back to normal channel speech (same as vt set 0)
vt next | vt prev                                step through the configured targets
vt whisper <name>                                whisper to one online user, no config entry needed
vt add <id> user <name> [<name> ...]             whisper target: one or more users
vt add <id> channel <name> [recursive=true] [links=true] [group=<name>]
vt help                                          usage summary
```

* Tab completes the subcommand, the ids that are actually configured after `vt set`, and the users who
  are actually online after `vt whisper`
* `vt set` and `vt clear` call the same code as `voicetargetset` but do not go through the HTTP command
  allow-list, so they work whether or not `<command action="voicetargetset">` is enabled under
  `<http>`. They also say what became active, and refuse an id that has no entry for the account in
  use instead of silently doing nothing
* `vt whisper <name>` uses voice target 30, the slot reserved for run time whispers, so it needs no
  `<voicetargets>` entry; clear it with `vt clear`
* `vt list` prints the targets of every account, marks the account in use, marks the target that is
  currently active, and — while connected — flags any entry whose user is offline or whose channel does
  not exist on the server. A target naming a channel that is not there silently falls back to plain
  channel speech when selected, so this is the quick way to spot a stale target
* `vt add` appends to an existing `<id>` or creates it, writes `talkkonnect.xml` in place and updates
  the running configuration, so `voicetargetset <id>` works immediately without a restart
* only the `<voicetargets>` element is touched. Comments, tag order and formatting elsewhere in the
  file are left exactly as they were, and the previous contents are kept as `talkkonnect.xml.bak`
* the edited file has to parse and contain the new entry before it replaces the original, so a bad
  edit fails with a message and changes nothing
* names containing spaces go in double quotes, e.g. `vt add 5 channel "Zone One" recursive=true`
* `recursive` and `links` default to true when given without a value (`recursive=`), and to false when
  the option is left out entirely

See also [voice-targets.md](./voice-targets.md).

---

### The Global Software Section

Everything under `<global><software>`: behaviour, logging, remote control, sounds and announcements.
Full reference: [CONFIGURATION.md §5](./CONFIGURATION.md#5-globalsoftware).

#### Settings

Core behaviour. All children are elements, not attributes. Audio device tags are covered under
[Audio configuration](#telling-talkkonnect-which-devices-to-use) above.

**Logging**

| Tag | Notes |
| --- | --- |
| `<logging>` | `screen`, `screenwithlineno`, `screenandfile`, `screenandfilewithlineno`. The `…withlineno` variants add the Go `file:line` to each entry. **Anything other than `screen` also writes to the log file** |
| `<logfilenameandpath>` | Absolute path to the log file, opened append-only and created if missing. If empty and logging is not `screen`, talKKonnect picks the first writable of: the config directory, the working directory, `/var/log/talkkonnect.log`, `/tmp/talkkonnect.log` |
| `<loglevel>` | `trace`, `debug`, `info`, `warning`, `error`, `alert`. Minimum severity printed; unrecognised or empty falls back to `info`. One of the few values applied by live reload |

**Startup and transmit behaviour**

| Tag | Notes |
| --- | --- |
| `<singleinstance>` | `true` takes a `talkkonnect.lock` lock file; a second instance announces the clash, waits 5 s and exits |
| `<cancellablestream>` | Lets an incoming transmission stop a stream you are playing into the channel |
| `<streamonstart>` / `<streamonstartafter>` | Start the configured stream automatically, this many **seconds** after connecting |
| `<streamsendmessage>` | Post a channel text message when streaming starts |
| `<txonstart>` / `<txonstartafter>` | Begin transmitting automatically, this many **seconds** after connecting |
| `<repeattxtimes>` / `<repeattxdelay>` | Number of transmissions and the gap between them for the `q` repeat-TX loop test. `0` in either disables the loop |
| `<simplexwithmute>` | Mute the speaker while transmitting — half-duplex radio behaviour, and what stops acoustic feedback on a single speaker-plus-mic build |
| `<txcounter>` | Count and display PTT presses since startup, for debugging |
| `<nextserverindex>` | Account index to connect to at startup, overriding `-serverindex`. Counting starts at 0 from the first `default="true"` account. Larger than the number of default accounts resets to `0` with a warning |
| `<txlockout>` | Refuse to transmit while someone else is speaking on the channel, play the `txlockout` event sound when TX is refused, and enable the panic function's `txlockenabled` |
| `<listentochannelsonstart>` | Activate the account's `<listentochannels>` list at boot |

Full detail: [CONFIGURATION.md §5.1](./CONFIGURATION.md#51-settings).

#### Channel Scan Section

Channel scanning makes talKKonnect behave like a scanning radio. It is started and stopped with the
same command — `r` at the prompt, or the `scanchannels` action over the HTTP API and MQTT.

While scanning, talKKonnect moves through all channels it has permission to enter (the same accessable
channel list used by channel up/down), one at a time. It waits on each channel for the dwell time, and
if someone starts speaking it holds there until the traffic stops plus the hang time, then continues.
Pressing PTT while parked on a channel also holds the scan until you release it. Changing channel by
hand — channel up/down, GPIO, keyboard or remote API — stops the scan and leaves talKKonnect on the
channel you chose.

```xml
<channelscan>
    <dwelltimemsecs>3000</dwelltimemsecs>
    <hangtimemsecs>4000</hangtimemsecs>
    <returntostartchannel>true</returntostartchannel>
    <skipchannels>Root,Test Channel</skipchannels>
</channelscan>
```

* `<dwelltimemsecs>` is how long to listen to each channel before moving on. Minimum **500** — a
  smaller non-zero value warns and 500 is used; `0` means use the built-in default
* `<hangtimemsecs>` is the extra time to stay on a busy channel after traffic stops. Same minimum
* `<returntostartchannel>` returns to the channel the scan started from when scanning stops
* `<skipchannels>` is a comma-separated list of channel **names or numeric ids** to leave out.
  Whitespace is trimmed and matching is case-insensitive
* the optional `startscanning` and `stopscanning` TTS sound actions are played when scanning starts
  and stops
* scanning needs at least 2 accessable channels, otherwise it refuses to start and logs a warning

Full detail: [CONFIGURATION.md §5.2](./CONFIGURATION.md#52-channelscan).

#### remotesshconsole

An embedded SSH server that serves the same bottom-CLI console as the terminal — the practical way to
reach a unit running in daemon mode.

```xml
<remotesshconsole enabled="true">
    <username>suvir</username>
    <password>secret</password>
    <idrsafile>/root/.ssh/id_rsa_tk</idrsafile>
    <listen>0.0.0.0:9999</listen>
</remotesshconsole>
```

All four children are mandatory when enabled — if any is empty the whole feature is disabled with a
warning. `<idrsafile>` is the server's own host key; generate it with
`ssh-keygen -t rsa -f /root/.ssh/id_rsa_tk`. `<listen>` is `ip:port`; use `0.0.0.0` to accept from any
interface. Access is additionally filtered by [`<networkacl>`](#networkacl).

Full detail: [CONFIGURATION.md §5.3](./CONFIGURATION.md#53-remotesshconsole).

#### Autoprovisioning Section

Autoprovisioning fetches the config from an HTTP server at boot and then reloads from the downloaded
copy, so a fleet of units can be provisioned centrally.

```xml
<autoprovisioning enabled="false">
    <tkid>device_id_12345</tkid>
    <url>http://provisioning.example.com/configs/</url>
    <savefilepath>/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect</savefilepath>
    <savefilename>talkkonnect.xml</savefilename>
</autoprovisioning>
```

* `<tkid>` identifies this device to the provisioning server and selects the file it serves
* `<url>` is the provisioning web server hosting the XML
* `<savefilepath>` and `<savefilename>` are where the fetched file is written locally

All four are mandatory when enabled, otherwise provisioning is disabled with a warning. **A
provisioning failure is fatal** — a unit that cannot reach its server will not start.

Full detail: [CONFIGURATION.md §5.4](./CONFIGURATION.md#54-autoprovisioning).

#### Beacon Section

The beacon emulates a radio repeater identification beacon, playing a WAV file at a fixed interval so
everyone on the channel knows the repeater is alive. It can play into the Mumble stream, out of the
local speaker (the RF side, when used as a repeater controller), or both.

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

* `<beacontimersecs>` is the interval in seconds. `0` disables the beacon with a warning
* `<beaconfileandpath>` is required when enabled
* `<localplay>` / `<localvolume>` play on the local speaker at `0–100` percent
* `<playintostream>` / `<beaconvolumeintostream>` transmit into the channel at gain `0.0–1.0`
* `<gpioenabled>` / `<gpioname>` drive an output pin while the beacon plays — used to key up a
  transmitter on a repeater build. The name must be a configured output pin
* **at least one of `<localplay>` or `<playintostream>` must be true**, otherwise the beacon is
  disabled with a warning

Full detail: [CONFIGURATION.md §5.5](./CONFIGURATION.md#55-beacon).

#### The TTS Section

Pre-recorded voice prompts for UI events — for users without an LCD or OLED who want audible feedback.
Each `<sound>` maps an internal event name to a WAV file.

```xml
<tts enabled="true" language="en-US">
    <volumelevel>10</volumelevel>
    <sound action="talkkonnectloaded" file="/…/Loaded.wav" blocking="true" enabled="true"/>
    <sound action="channelup" file="/…/ChannelUp.wav" blocking="false" enabled="false"/>
</tts>
```

* `enabled` on `<tts>` is the master switch — `false` and no prompt plays
* `<volumelevel>` (`0–100`) applies to **all** prompts
* `blocking="true"` waits for playback to finish before continuing — use it for shutdown prompts
* `enabled` on each `<sound>` lets you turn on only the events you care about

Valid `action` values, which are the names the code actually asks for:

`channelup`, `channeldown`, `currentrxvolumelevel`, `currenttxvolumelevel`, `digitalvolumeup`,
`digitalvolumedown`, `displaymenu`, `listonlineusers`, `listserverchannels`, `mutespeaker`,
`unmutespeaker`, `nextserver`, `previousserver`, `panicsimulation`, `pingservers`, `playstream`,
`printxmlconfig`, `quittalkkonnect`, `requestgpsposition`, `sendemail`, `startscanning`,
`stopscanning`, `starttransmitting`, `stoptransmitting`, `talkkonnectloaded`.

**Matching is exact and case-sensitive.** `muteSpeaker` or `currentvolumelevel` will never fire — see
[Known quirks](#known-quirks-and-traps). Entries with an unrecognised `action` are harmless but dead.
The `language` attribute on `<tts>` is parsed and printed but not used; generated-speech language
comes from `<ttsmessages><ttslanguage>`.

Full detail: [CONFIGURATION.md §5.6](./CONFIGURATION.md#56-tts).

#### The SMTP Section

Email alerts, used by the `e` key / `sendemail` action and by the panic function. talKKonnect talks to
one SMTP server, Gmail's.

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

`<username>`, `<password>` and `<receiver>` are mandatory when enabled; missing any of them disables
SMTP with a warning. Use a Gmail **app password**, not the account password. The three GPS flags
append the current fix time, the latitude/longitude, and a clickable Google Maps link to the message
body — all three need a working [`<gps>`](#the-gps-section).

Full detail: [CONFIGURATION.md §5.7](./CONFIGURATION.md#57-smtp).

#### The Sounds Section

Three independent groups: sounds for channel events, feedback for physical controls, and the repeater
tone.

**Event sounds** fire on things that happen on the server or in the channel:

```xml
<sounds>
    <sound event="rogerbeep" file="/…/rogerbeeps/Waterdrop.wav" volume="50" blocking="false" enabled="true"/>
</sounds>
```

| `event` | Fires when |
| --- | --- |
| `joinedchannel` | A user joins your channel |
| `leftchannel` | A user leaves your channel |
| `message` | A text message arrives |
| `incommingbeep` | An incoming transmission starts |
| `rogerbeep` | Your transmission ends — pre-loaded into PCM at startup for low latency |
| `txlockout` | TX was refused because of lock-out |
| `stream` | The stream played into the channel by the `b` key |

`file` is an absolute path or an `http`/`rtsp` URL; a missing local file disables that entry with a
warning. `volume` is `0–100`, and **`volume="0"` disables the sound** with a warning. `blocking` waits
for the sound to finish before carrying on.

The `stream` event is more powerful than it looks: point it at a local file or a network stream and
the `b` key plays it into the Mumble channel — useful for testing and for piping an audio source into
a channel.

**`<input>` — feedback for physical controls.** The `enabled` attribute on `<input>` is a master
switch. Event names must match exactly what the code asks for: GPIO events are prefixed `io`, USB
keyboard events `usb`.

```xml
<input enabled="true">
    <sound event="iotxpttstart" file="/…/YellowJacket.wav" enabled="true"/>
    <sound event="iovolup"     file="/…/RC210#2.wav"      enabled="true"/>
</input>
```

*GPIO:* `iotxpttstart`, `iotxpttstop`, `txtogglestart`, `iotxtogglestop`, `iochannelup`,
`iochanneldown`, `iovolup`, `iovoldown`, `iopanic`, `iostreamtoggle`, `iocommenton`, `iocommentoff`,
`iorotarycw`, `iorotaryccw`, `iorotarybutton`, `iotrackingon`, `iotrackingoff`, `iolisteningstart`,
`iolisteningstop`, `iorepeatertone`, `iocnextserver`, `memorychannel`, `changechannel`, `shutdown`

*USB keyboard:* `usbchannelup`, `usbchanneldown`, `usbvolup`, `usbvoldown`, `usbcurrentrxvol`,
`usbcurrenttxvol`, `usbmute`, `usbunmute`, `usbmutetoggle`, `usbstarttx`, `usbstoptx`,
`usbstartlisten`, `usbstopliosten`, `usbstreamtoggle`, `usbrecord`, `usbsetcomment`, `usbserverup`,
`usbpreviousserver`, `usbvoicetarget`, `usbmqttpubpayloadset`

There is no volume attribute here — these play through `aplay` at the device volume. `usbstopliosten`
and `iocnextserver` really are spelled that way in the code; use them verbatim.

**`<repeatertone>`** generates a sine wave of a given frequency and duration, to open an RF repeater.
Requested by European hams, who need 1750 Hz.

```xml
<repeatertone enabled="true">
    <sound event="repeatertone" tonefrequencyhz="1750" volume="50"
           tonedurationsec="1" direction="local" blocking="false" enabled="true"/>
</repeatertone>
```

`direction="intostream"` transmits the tone into the Mumble channel; **any other value** (such as
`local`) plays it on the local speaker or radio input. `tonedurationsec` accepts fractions. The WAV is
generated on demand with `ffmpeg` into `soundfiles/repeatertones/sine_<freq>_<duration>.wav` and
cached, so `ffmpeg` and `ffplay` must be installed.

Full detail: [CONFIGURATION.md §5.8](./CONFIGURATION.md#58-sounds).

#### The Remote Control Section

`<remotecontrol>` holds the HTTP API, the JSON status endpoint, the network ACL and MQTT.

##### The HTTP API

`<http>` granularly controls which remote actions are reachable over HTTP. `listenport` is the port
talKKonnect listens on.

```xml
<http enabled="true" listenport="8080">
    <command action="channelup" funcparamname="" message="Channel Up" enabled="true"/>
    <command action="joinchannel" funcparamname="value" message="Join Channel" enabled="true"/>
</http>
```

Commands are invoked as `http://<host>:<listenport>/?command=<action>` plus any query parameters:

```
List the enabled API      http://{your-talkkonnect-ipaddress}:8080/?command=listapi
Channel Up                http://{your-talkkonnect-ipaddress}:8080/?command=channelup
Channel Down              http://{your-talkkonnect-ipaddress}:8080/?command=channeldown
Mute/Unmute Toggle        http://{your-talkkonnect-ipaddress}:8080/?command=mute-toggle
Volume Up                 http://{your-talkkonnect-ipaddress}:8080/?command=volumerxup
Start Transmitting        http://{your-talkkonnect-ipaddress}:8080/?command=starttransmitting
Stop Transmitting         http://{your-talkkonnect-ipaddress}:8080/?command=stoptransmitting
Play/Stop Stream          http://{your-talkkonnect-ipaddress}:8080/?command=playback
Request GPS Position      http://{your-talkkonnect-ipaddress}:8080/?command=gpsposition
Send Email                http://{your-talkkonnect-ipaddress}:8080/?command=sendemail
Join a Channel            http://{your-talkkonnect-ipaddress}:8080/?command=joinchannel&channel=Root/Ops
Set Voice Target          http://{your-talkkonnect-ipaddress}:8080/?command=voicetargetset&id=1
Play an Announcement      http://{your-talkkonnect-ipaddress}:8080/?command=announcement&mediaid=main_announcement
Multicast Toggle          http://{your-talkkonnect-ipaddress}:8080/?command=multicasttoggle
```

A command runs only if **both** conditions hold: the `action` is one of the built-in handlers, **and**
a `<command>` element with that action exists in this section. An action that is not built in gets a
`400`; one missing from the XML gets a `404`. `enabled="false"` keeps a command out of the `listapi`
listing.

Built-in actions:

`displaymenu`, `channelup`, `channeldown`, `mute`, `unmute`, `mute-toggle`, `currentrxvolume`,
`currenttxvolume`, `volumerxup`, `volumerxdown`, `volumetxup`, `volumetxdown`, `setrxvolume`,
`listserverchannels`, `joinchannel`, `whisperuser`, `whisperclear`, `starttransmitting`,
`stoptransmitting`, `listonlineusers`, `playback`, `gpsposition`, `sendemail`, `previousserver`,
`connnextserver`, `clearscreen`, `pingservers`, `panicsimulation`, `repeattxloop`, `scanchannels`,
`thanks`, `showuptime`, `showversion`, `dumpxmlconfig`, `ttsannouncement`, `announcement`,
`voicetargetset`, `listeningstart`, `listeningstop`, `radiotoggle`, `radionext`, `radioprev`,
`radiovolup`, `radiovoldown`, `multicaston`, `multicastoff`, `multicasttoggle`, `listapi`.

Query parameters — all case-insensitive. The `funcparamname` attribute is documentation only; these
names are what the server actually reads:

| Parameter | Used by | Type |
| --- | --- | --- |
| `id` | `voicetargetset` | integer 0–31 |
| `channel` | `joinchannel` | channel name or path |
| `user` | `whisperuser` | username |
| `volume` | `setrxvolume` | integer |
| `mediaid` | `announcement` | `<multimedia>` profile id |
| `ttsmessage` | `ttsannouncement` | text to speak |
| `ttslocalplay`, `ttsplayintostream` | `ttsannouncement` | `true`/`false` |
| `gpioenabled`, `gpioname` | `ttsannouncement` | `true`/`false`, pin name |
| `predelay`, `postdelay` | `ttsannouncement` | integer |
| `language` | `ttsannouncement` | language code, e.g. `en` |

```bash
curl "http://192.168.1.50:8080/?command=ttsannouncement&ttsmessage=Muster+at+gate+3&ttslocalplay=true"
curl "http://192.168.1.50:8080/?command=joinchannel&channel=Root/Operations"
```

The same listener serves the **XML config editor** at `/config`, which saves the posted file to disk
and live-reloads it. Full request and response detail is in [api.md](./api.md); the tag reference is
[CONFIGURATION.md §5.9](./CONFIGURATION.md#59-remotecontrol).

##### uistatus

A pretty-printed JSON snapshot of the current state — channel, users, talk state, version, multicast —
for an external framebuffer or web UI.

```xml
<uistatus enabled="true" listenip="127.0.0.1" listenport="8080" url="/uistatus"/>
```

If the `<uistatus>` element is absent entirely, `enabled` follows `<http enabled>`. Set
`enabled="false"` when you are using the built-in LCD/OLED rather than an external client. `listenip`
defaults to `127.0.0.1` (use `0.0.0.0` for all interfaces) and `listenport` defaults to the HTTP API
port — if the two match, the same listener serves both, otherwise a second listener starts. `url`
defaults to `/uistatus` and a leading `/` is added if you omit it.

##### networkacl

```xml
<networkacl enabled="true">
    <network cidr="127.0.0.1/32"/>
    <network cidr="192.168.1.0/24"/>
</networkacl>
```

Restricts the HTTP API, `/config`, `/uistatus` **and** the SSH console to these CIDRs; non-matching
clients get a `403`. Invalid CIDRs are skipped with a warning. **Enabled with no valid entries means
everything is allowed** — you get a warning saying so, which is easy to miss.

##### The MQTT Section

MQTT lets you remote-control talKKonnect over a public or local broker, which sidesteps the problem of
reaching units behind NAT anywhere on the internet. You can drive commands and relays for external
devices, and publish status back.

```xml
<mqtt enabled="true">
    <settings enabled="true">
        <mqttsubtopic>event/cameraman1/</mqttsubtopic>
        <mqttpubtopic>response/cameraman1</mqttpubtopic>
        <mqttbroker>tcp://mqtt.yourserver.com:1883</mqttbroker>
        <mqttuser>camera1</mqttuser>
        <mqttpassword>yourpassword</mqttpassword>
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
        <command action="channelup" message="Channel Up" enabled="true"/>
        <command action="channeldown" message="Channel Down" enabled="true"/>
        <command action="muteunmute" message="Mute-Toggle" enabled="true"/>
        <command action="volumetxup" message="Mic Volume Up" enabled="true"/>
        <command action="volumetxdown" message="Mic Volume Down" enabled="true"/>
        <command action="scanchannels" message="Scan Channels" enabled="true"/>
        <command action="voicetargetset" message="Set Voice Target" enabled="true"/>
        <command action="multicasttoggle" message="Multicast Toggle" enabled="true"/>
        <command action="attention" message="Attention LED" enabled="true"/>
    </commands>
</mqtt>
```

`<mqttsubtopic>`, `<mqttpubtopic>`, `<mqttbroker>`, `<mqttpassword>` and `<mqttid>` are all mandatory
when enabled — an empty one disables MQTT with a warning. (`<mqttuser>` is not checked.) `+` and `#`
wildcards work in the subscribe topic; `<store>` is the persistence directory used when `cleansess`
is false; `<attentionblinktimes>` / `<attentionblinkmsecs>` set the blink pattern for the `attention`
command; `<pubpayload>` holds payloads selectable by the `mqttpubpayloadset` keyboard action,
addressed by `item`.

**Payload format.** The payload is lower-cased, `:` is stripped, then split on spaces. Word 0 is the
action and must match a `<command action>` entry. So with the config above, publishing `volumetxup`
to `event/cameraman1` makes that unit's microphone louder.

| Payload | Effect |
| --- | --- |
| `muteunmute mute` / `unmute` / `toggle` | Speaker mute control |
| `attention on` / `off` / `blink` | Drive the `attention` output pin — for example to get a user's attention |
| `voicetargetset <0-31>` | Select a voice target slot |
| `announcement <profileid>` | Play a `<multimedia>` profile |
| `relay <1\|2> on\|off\|pulse` | Intended to pulse a relay, e.g. to open an access-control door. **Currently broken** — see [Known quirks](#known-quirks-and-traps) |
| any other action, no arguments | Runs the command |

**MQTT actions are a different, smaller set than the HTTP ones** — note `muteunmute` in place of
`mute`/`unmute`/`mute-toggle`, and no `ttsannouncement`, `showversion`, `joinchannel`, `whisperuser`,
`setrxvolume` or `listapi`:

`displaymenu`, `channelup`, `channeldown`, `muteunmute`, `currentrxvolume`, `currenttxvolume`,
`volumerxup`, `volumerxdown`, `volumetxup`, `volumetxdown`, `listserverchannels`,
`starttransmitting`, `stoptransmitting`, `listonlineusers`, `playback`, `gpsposition`, `sendemail`,
`previousserver`, `connnextserver`, `clearscreen`, `pingservers`, `panicsimulation`, `repeattxloop`,
`scanchannels`, `thanks`, `showuptime`, `dumpxmlconfig`, `announcement`, `voicetargetset`,
`listeningstart`, `listeningstop`, `radiotoggle`, `radionext`, `radioprev`, `radiovolup`,
`radiovoldown`, `multicaston`, `multicastoff`, `multicasttoggle`, `attention`, `relay`.

`attention` and `relay` need the matching output pin defined under [`<pins>`](#pins).

#### The PrintVariables Section

Pure diagnostics: each flag controls whether that section is dumped to the log at startup and on
`dumpxmlconfig` (the `d` / `w` keys). All are `true`/`false` elements:

`printaccount`, `printsystemsettings`, `printremotesshconsole`, `printprovisioning`, `printbeacon`,
`printtts`, `printsmtp`, `printsounds`, `printhttpapi`, `printmqtt`, `printttsmessages`,
`printignoreuser`, `printhardware`, `printgpioexpander`, `printmax7219`, `printpins`, `printrotary`,
`printpulse`, `printvolumebuttonstep`, `printheartbeat`, `printcomment`, `printlcd`, `printoled`,
`printgps`, `printtraccar`, `printpanic`, `printusbkeyboard`, `printaudiorecord`, `printkeyboardmap`,
`printradiomodule`, `printmultimedia`, `printmemorychannels`, `printpresetvoicetargets`,
`printmulticast`.

Turning these off is worth doing on a production unit — `printaccount` dumps account details, though
passwords are masked. `<printlistentochannels>` is a string and is currently parsed but not consulted.

#### TTSMessages Section

Generated speech via Google Translate, as opposed to the pre-recorded prompts in
[`<tts>`](#the-tts-section). This is what makes talKKonnect an annunciator: it is used by
`ttsannouncement`, station announcements and spoken incoming messages.

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

* `<ttslanguage>` is the language code used for synthesis — `en`, `de`, `th` and so on. See Google
  Translate for the codes. `<ttslanguagethai>` is used instead when the text is detected as Thai
* `<ttsmessagefromtag>` set to true prefixes the announcement with "message from". For plain
  announcements leave it false
* `<ttstone>` is an attention tone (an absolute path to a WAV) played before the announcement
* `<localblocking>` waits for local playback to finish, so sounds play in order
* `<ttssounddirectory>` is where generated audio is cached — relative paths resolve against the
  working directory, and the directory is created if missing
* `<localplay>` plays out of the local speaker or amplifier; `<playintostream>` transmits into the
  channel so everyone hears it
* `<speakvolumeintostream>` (`0–100`) is the volume of synthesised speech into the stream;
  `<playvolumeintostream>` (`0.0–1.0`) is the gain for file playback into the stream
* `<gpio name= enabled=>` drives an output pin for the whole announcement — for switching an external
  amplifier or a flashing attention light
* `<predelay>` and `<postdelay>` are **documented as seconds but treated as nanoseconds**, so they
  produce no usable delay. Use a `<multimedia>` profile when you need a real gap

Generated speech needs outbound internet access. Cached files are reused, so repeated announcements
keep working offline.

Full detail: [CONFIGURATION.md §5.11](./CONFIGURATION.md#511-ttsmessages).

#### Ignore User Section

For when you do not want to hear a particular user — most often yourself, because you are also
transmitting from another device in the same room.

```xml
<ignoreuser enabled="false">
    <ignoreuserregex>(?i)bot|spam</ignoreuserregex>
</ignoreuser>
```

`<ignoreuserregex>` is a Go regular expression matched against the username; audio and messages from
matching users are dropped. `(?i)` makes it case-insensitive. The regex must be at least 4 characters
or the feature is disabled with a warning. Ignored users are left out of multicast too.

#### Memory Channels and Preset Voice Targets

Both bind a GPIO button to a fixed setting: press it and jump straight there.

```xml
<memorychannels enabled="true">
    <channel gpioname="memorychannel1" channelname="HAM-CB" enabled="true"/>
    <channel gpioname="memorychannel2" channelname="Root/Favorite2" enabled="true"/>
</memorychannels>

<presetvoicetargets enabled="true">
    <voicetargetset gpioname="presetvoicetarget1" id="10" enabled="true"/>
</presetvoicetargets>
```

`gpioname` **must** be one of `memorychannel1` … `memorychannel4` or `presetvoicetarget1` …
`presetvoicetarget5`; any other value disables the **whole section** with a warning. A matching input
pin with the same `name` must also exist under [`<pins>`](#pins). For preset voice targets, `id` is a
slot (1–31) defined under the account's `<voicetargets>`.

### The Multicast Section (RTP to IP Speakers and SIP Phones)

talKKonnect can re-transmit the audio it receives from Mumble as an RTP stream to a multicast group,
so hardware IP PA speakers (CyberData, Algo, Barix, Advanced Network Devices) and SIP desk phones such
as Yealink can hear the channel without being Mumble clients. This is the same wire format
[gochimesd](https://github.com/talkkonnect/gochimesd) sends, and the tags are named to match, so one
receiver configuration works against both programs.

Audio leaves as **8 kHz mono, 20 ms per packet**, which is what those devices decode. Several people
talking at once are mixed into a single RTP stream with one SSRC, and nothing is sent while the channel
is quiet; the first packet after a silent gap carries the RTP marker bit so a hardware jitter buffer
resets.

```xml
<multicast enabled="false">
    <group>239.0.1.10</group>
    <port>5004</port>
    <codec>pcmu</codec>
    <ttl>1</ttl>
    <interface></interface>
    <packetms>20</packetms>
    <l16payloadtype>96</l16payloadtype>
    <volume>100</volume>
    <allchannels>false</allchannels>
    <hangoverms>200</hangoverms>
    <include>
        <!-- <user>somebody</user> -->
    </include>
    <exclude>
        <!-- <user>noisybox</user> -->
    </exclude>
</multicast>
```

* `enabled` on the section switches the whole feature on and off at startup. It can also be switched at
  run time, see below
* `group` is the IPv4 multicast group (`224.0.0.0/4`, typically `239.x.y.z`) and `port` the UDP port. A
  non-multicast or IPv6 address disables the section. Keep the port **even**: by RTP convention the odd
  port above it belongs to RTCP, and some receivers insist on it
* `codec` is `pcmu` (G.711 u-law, the default and the only one every hardware decoder accepts), `pcma`
  (G.711 A-law) or `l16` (uncompressed PCM). **l16 rides a dynamic RTP payload type and G.711-only
  receivers drop it silently** — the group transmits correctly yet the speaker stays mute — so the
  config checker warns when it is selected. Use `pcmu` unless every listener on the group is known to
  support L16. Unrecognised values fall back to `pcmu` with a warning
* `l16payloadtype` is the dynamic payload type used for `l16`, 96 to 127
* `ttl` is 1 by default, which keeps the stream on the local subnet. Crossing a routed VLAN needs a
  higher value **and** IGMP/PIM configured on the network
* `interface` is the network interface the group leaves by, e.g. `eth0` or `ens18`. Leave it empty to
  follow the default route. A name that does not exist on the host is reported as a warning at startup
  rather than silently going out of the wrong interface
* `packetms` is the packetization interval, one of 10, 20, 30, 40 or 60. 20 ms is the interoperable
  default; other values warn and reset to 20
* `volume` is a software gain in percent applied to the outgoing stream, `1–200`, where 100 is unity
  and above 100 amplifies
* `hangoverms` (`0–5000`, default 200) is how long a silent talker holds the RTP stream open
* `allchannels` false, the default, carries only the channel talKKonnect is joined to. Set it true to
  also carry audio heard through `<listentochannels>` and incoming whispers — useful for a monitoring
  node, but it does mean a monitored channel is re-broadcast over the PA
* `include` lists the users whose audio is carried. **An empty list means every talker**
* `exclude` lists users whose audio is never carried, and **exclude wins over include**. Names are
  matched case insensitively, a name in both lists warns, and a user matched by `<ignoreuser>` is left
  out of the multicast too

Announcements can be multicast as well: set `<multicast>true</multicast>` in a `<multimedia>` profile
(see [the Multimedia section](#the-multimedia-section-pre-recorded-announcements--ip-speaker)) and the
profile's tone and media files go to the group, independently of `localplay` and `playintostream`.
This needs `ffmpeg` on the `PATH`.

Run time control, all of which report through the same status:

* `mc status`, `mc on`, `mc off`, `mc toggle` in the bottom CLI and the SSH console (`mc help` for the
  full usage, Tab completes the subcommand). These always work, whether or not the actions below are
  enabled in the XML
* the `multicaston`, `multicastoff` and `multicasttoggle` actions over the HTTP API and MQTT, each
  gated by its own `<command action="…">` entry under `<http>` and `<mqtt>` like every other action
* the `/uistatus` JSON snapshot carries a `multicast` object with the destination, codec, whether the
  sender is running, who is being mixed right now and how many packets have been sent
* `cfg set global.software.multicast.…` followed by a config reload restarts the sender with the new
  settings. The `include`, `exclude` and `allchannels` values take effect immediately without a restart

To check the stream from another machine on the same subnet: `ffplay rtp://@239.0.1.10:5004`.

See also [multicast.md](./multicast.md) and
[CONFIGURATION.md §5.15](./CONFIGURATION.md#515-multicast).

---

## The Hardware Section

Everything under `<global><hardware>`. Full reference:
[CONFIGURATION.md §6](./CONFIGURATION.md#6-globalhardware).

```xml
<hardware targetboard="rpi">
    <ledstripenabled>false</ledstripenabled>
    <gpiooffset>0</gpiooffset>
    <voiceactivitytimermsecs>200</voiceactivitytimermsecs>
    …
</hardware>
```

* `targetboard` is the master switch for everything hardware. **`rpi`** enables the GPIO, LCD and OLED
  paths; **`pc`** — for a PC, server or VM with no GPIO, buttons or screen — makes GPIO calls no-ops
  and forces the LCD backlight timer off. Any value other than `rpi` behaves as `pc`
* `<ledstripenabled>` turns on the addressable LED strip, as used on the ReSpeaker HAT. When on, pins
  7–11 clash with SPI and warn. `<voiceactivitytimermsecs>` is how long the voice-activity LED stays
  lit after audio — minimum **200**, and anything smaller is raised to 200
* `<gpiooffset>` is added to every `<pin pinno>` at startup when greater than 0. Needed on boards whose
  gpiochip base is not 0, which is many Orange Pi and mainline-kernel systems

### The IO Section

The IO section was rewritten to be flexible enough to add new functions without new tags. It covers
port expanders, the 7-segment driver, every GPIO pin, the rotary encoder and pulse timing.

**`<gpioexpander>` — MCP23017 I²C port expanders.** Up to 8 chips on one bus, currently used for
outputs.

```xml
<gpioexpander enabled="true">
    <chip id="0" i2cbus="1" mcp23017device="32" enabled="true"/>
</gpioexpander>
```

`id` (≤ 8) is what a pin's `chipid` refers to, `i2cbus` is the bus number (usually `1` on a Pi), and
`mcp23017device` is the I²C address **in decimal** (`32` = 0x20). Enabling any expander makes GPIO
pins 2 and 3 warn as I²C clashes.

**`<max7219>` — 8-digit 7-segment driver on SPI**, used to show the channel id.

```xml
<max7219 enabled="false" max7219cascaded="1" spibus="0" spidevice="0" brightness="5"/>
```

`brightness` is 0–7. **These must be written as attributes**, as above — some sample configs write
them as child elements, which leaves them all at zero.

#### pins

The heart of a hardware build: one `<pin>` per physical connection.

```xml
<pins>
    <pin direction="output" device="led/relay"  name="transmit" pinno="22" type="gpio" chipid="0" inverted="false" enabled="true"/>
    <pin direction="input"  device="pushbutton" name="txptt"    pinno="26" type="gpio" chipid="0" enabled="true"/>
</pins>
```

| Attribute | Valid values | Notes |
| --- | --- | --- |
| `direction` | `input`, `output` | Anything else disables the pin |
| `device` | inputs: `pushbutton`, `toggleswitch`, `rotaryencoder`<br>outputs: `led/relay`, `lcd` | A device that does not match the direction disables the pin with a warning |
| `name` | see below | Must be a recognised name or the pin is disabled |
| `pinno` | **1–27** or **513–539** | BCM numbering. `<gpiooffset>` is added if set. Anything else disables the pin |
| `type` | `gpio`, `mcp23017` | `gpio` is a SoC pin; `mcp23017` routes through the expander named by `chipid` |
| `chipid` | 0–8 | Which `<gpioexpander><chip id>` this pin belongs to. **Use 0 when no expander is fitted** |
| `inverted` | bool | Outputs only: `true` drives the pin low for "on", for active-low relay boards and sinking LED wiring. Normally `false` |
| `enabled` | bool | Per-pin switch. Disabled pins are not validated |

**Output names:** `voiceactivity` (lit while receiving), `participants` (lit when others are in the
channel), `transmit` (lit while transmitting), `online` (lit while connected), `attention` (driven by
the MQTT `attention` command), `voicetarget` (lit while a voice target is active), `heartbeat` (blinks
per [`<heartbeat>`](#the-heartbeat-section-output)), `backlight` (LCD backlight — use `device="lcd"`),
`relay0`, `analogrelay1`, `analogrelay2`.

**Input names:** `txptt` (momentary PTT), `txtoggle` (latching transmit), `channelup`, `channeldown`,
`volup`, `voldown`, `panic`, `streamtoggle`, `comment` (use `device="toggleswitch"`), `rotarya`,
`rotaryb`, `rotarybutton` (use `device="rotaryencoder"`), `nextserver`, `repeatertone`,
`memorychannel1` … `memorychannel4`, `presetvoicetarget1` … `presetvoicetarget5`, `shutdown`.

> **Names that work in code but are rejected by the validator.** `listening`, `tracking`, `mqtt0`,
> `mqtt1`, `internetradiotoggle`, `internetradiochannelup`, `internetradiochanneldown`,
> `internetradiovolup` and `internetradiovoldown` all have working handlers but are **not** in the
> validator's allow-list, so an enabled pin using one is disabled at startup with `Invalid Name`.
> Reach those functions from the keyboard, HTTP API or MQTT instead. Likewise `radiomodule` is not a
> valid `device` value even though some `<pins>` examples use it.

Only `type="gpio"` inputs are polled as buttons — MCP23017 pins are configured as inputs but the
button-press paths run on SoC pins.

#### rotaryencoder

The rotary encoder can control the Mumble channel, the local volume, the SA818 radio channel or the
voice target. Enabled controls form a ring: pressing the encoder's button cycles to the next function
and turning the knob acts on the current one.

```xml
<rotaryencoder enabled="true">
    <control function="mumblechannel" enabled="true"/>
    <control function="localvolume" enabled="true"/>
</rotaryencoder>
```

Valid `function` values are `mumblechannel`, `localvolume`, `radiochannel` and `voicetarget`. You also
need `rotarya`, `rotaryb` and `rotarybutton` input pins with `device="rotaryencoder"` and `chipid="0"`.
If the encoder responds too quickly or too slowly, adjust `<pulse>`.

#### pulse and volumebuttonstep

```xml
<pulse leadingmsecs="1000" pulsemsecs="1000" trailingmsecs="1000"/>

<volumebuttonstep>
    <volupstep>1</volupstep>
    <voldownstep>-1</voldownstep>
</volumebuttonstep>
```

`<pulse>` takes **attributes**, not elements: the delay before a pulsed output, the pulse width, and
the delay after. `<volumebuttonstep>` is the mixer step per button press — tune it to taste and to
your sound card. `0` is replaced by `+1` / `-1`, and `<voldownstep>` should be negative.

#### The Heartbeat Section (OUTPUT)

A "the software is alive" blink on an output pin named `heartbeat`.

```xml
<heartbeat enabled="true">
    <heartbeatledpin>heartbeat</heartbeatledpin>
    <periodmsecs>2000</periodmsecs>
    <ledonmsecs>1000</ledonmsecs>
    <ledoffmsecs>1010</ledoffmsecs>
</heartbeat>
```

All three timings should be ≥ 100 ms. The heartbeat can share a pin with `voiceactivity` so one LED
does both jobs — but **do not** do that, or enable the heartbeat at all on that pin, if talKKonnect is
wired to key a transceiver.

#### The Comment Section

Two Mumble comment strings selected by the position of a toggle switch — an away-message switch.
When another user lists online users they see your username with the current message in square
brackets.

```xml
<comment>
    <commentbuttonpin>comment</commentbuttonpin>
    <commentmessageoff>Status: Available</commentmessageoff>
    <commentmessageon>Status: Busy</commentmessageon>
</comment>
```

`<commentbuttonpin>` is parsed but **ignored** — the binding comes from the input pin *named* `comment`
under `<pins>`, with `device="toggleswitch"`. The same is true of `<listening>`: neither of its
children is read, and listen-to-channel is driven by `<listentochannelsonstart>` and the
`listeningstart` / `listeningstop` actions.

#### The LCD Section (HD44780 character LCD)

talKKonnect supports the common 4-line, 20-character HD44780 module, in 4-bit parallel mode or over an
I²C backpack. Requires `targetboard="rpi"`; set `enabled="false"` to turn it off.

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

* `<lcdinterfacetype>` is `i2c` or `parallel`; anything else disables the LCD
* `<lcdi2caddress>` is the **decimal** I²C address — `0x3f` from `i2cdetect` is `63`. Required and
  non-zero for `i2c`
* `<lcdrspin>` … `<lcdd7pin>` are the BCM pins wired to the display in `parallel` mode. Required and
  non-zero — a single zero disables the LCD
* `<lcdbacklighttimerenabled>` / `<lcdbacklighttimeoutsecs>` blank the backlight after that many idle
  seconds. `0` seconds disables the timer with a warning, and the timer is forced off unless
  `targetboard="rpi"`. It also has an odd guard — see [Known quirks](#known-quirks-and-traps)
* **GPIO pins 2 and 3 cannot be used for anything else** if you are connecting an I²C display

#### The OLED Section (I2C OLED)

talKKonnect supports the common 0.96 and 1.3 inch SSD1306-class I²C OLED panels. `i2c` is the only
interface — SPI has not been developed. Requires `targetboard="rpi"`. There is no backlight function
for OLEDs.

```xml
<oled enabled="true">
    <oledinterfacetype>i2c</oledinterfacetype>
    <oleddisplayrows>8</oleddisplayrows>
    <oleddisplaycolumns>21</oleddisplaycolumns>
    <oleddefaulti2cbus>1</oleddefaulti2cbus>
    <oleddefaulti2caddress>60</oleddefaulti2caddress>
    <oledscreenwidth>132</oledscreenwidth>
    <oledscreenheight>64</oledscreenheight>
    <oledcommandcolumnaddressing>33</oledcommandcolumnaddressing>
    <oledaddressbasepagestart>176</oledaddressbasepagestart>
    <oledcharlength>5</oledcharlength>
    <oledstartcolumn>1</oledstartcolumn>
</oled>
```

* `<oleddefaulti2caddress>` is **decimal** — `3c` from `i2cdetect` is `60`. The bus is usually `1`
* `<oleddisplayrows>` / `<oleddisplaycolumns>` are the text grid; 8 rows by 21 columns suits a typical
  panel. `<oledscreenwidth>` / `<oledscreenheight>` are the panel pixels, commonly 132 by 64
* `<oledstartcolumn>` is `0` for 0.96 inch panels and `1` for 1.3 inch panels — this is what clears
  garbage along the edge of the screen
* `<oledcommandcolumnaddressing>` (33 = 0x21), `<oledaddressbasepagestart>` (176 = 0xB0) and
  `<oledcharlength>` (font width in pixels) rarely need changing
* if the panel does not answer at startup, OLED support switches itself off and logs
  `Cannot Communicate with OLED`
* `<oledttf …>` appears in some shipped configs but is **not parsed**
* **GPIO pins 2 and 3 cannot be used for anything else** with an I²C display

#### The GPS Section

An NMEA GPS on a serial port — a u-blox 6 USB module, for example, which usually appears as
`/dev/ttyACM0`. Used by the panic function, `sendemail` and Traccar.

```xml
<gps enabled="true">
    <port>/dev/ttyAMA0</port>
    <baud>9600</baud>
    <txdata/>
    <even>false</even>
    <odd>false</odd>
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

Set `enabled="false"` if you have no GPS. `<port>` **must exist at startup** or GPS is disabled with a
warning. `<baud>` must be one of 2400, 4800, 9600, 14400, 19200, 38400, 57600 or 115200 — anything
else disables GPS. Setting **both** `<even>` and `<odd>` disables GPS; both false means no parity.
`<stopbits>` and `<databits>` must be non-zero, typically 1 and 8. `<gpsinfoverbose>` logs every
parsed sentence and is for bring-up only; `<gpsdiagsounds>` gives audible feedback when a fix is
acquired or lost; `<gpsdisplayshow>` puts the position on the LCD/OLED. The `<rs485*>` tags handle
RS-485 transceiver direction control.

#### The Traccar Section

Reports the GPS position to a [Traccar](https://www.traccar.org/) server.

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

**Both `enabled` and `<track>` must be true** for positions to be sent. `<clientid>` is the device
identifier registered in Traccar. Only the child block matching `<protocol name>` is used, so the
others can stay for reference: `osmand` and `opengts` take a `serverurl` including the scheme, `t55`
takes a bare `serverip`. The default Traccar ports are 5055 (OsmAnd), 5001 (T55) and 5177 (OpenGTS).
Note the typo in `traccardispayshow` — it is spelled that way in the parser.

#### The PanicFunction Section

An emergency "I need help" alert, triggered by the `panic` input pin or the `panicsimulation` action.

```xml
<panicfunction enabled="true" blocking="false">
    <filenameandpath>/…/soundfiles/alerts/alert.wav</filenameandpath>
    <volume>10</volume>
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

* `<filenameandpath>` is the WAV transmitted into the channel, at gain `<volume>` (`0.0–1.0`)
* `blocking` is an **attribute** on `<panicfunction>`, not an element
* `<sendident>` includes the account's `<ident>` in the message — use it when you want your name or an
  alternate id sent
* `<panicmessage>` is the text sent to the channel, and to sub-channels as well when
  `<recursivesendmessage>` is true
* `<sendgpslocation>` appends the current fix, so needs [`<gps>`](#the-gps-section)
* `<panicemail>` also sends the alert by email, so needs [`<smtp>`](#the-smtp-section) enabled
* `<eavesdrop>` opens the microphone so the channel can hear the scene
* `<txlockenabled>` / `<txlocktimeoutsecs>` hold talKKonnect in transmit for that many seconds after
  the button is pressed, so the requester can talk without holding PTT. Honoured only when
  `<txlockout>` is true in `<settings>`
* `<lowprofile>` is a silent panic: local sound, LEDs and the screen are suppressed so a bystander
  cannot tell it fired

#### The USB Keyboard Section

A wired or wireless USB numpad or PTT handset for controlling a headless unit — no terminal login
needed, just plug and play. Many CM-series sound cards and USB PTT handsets present their buttons as a
HID keyboard, which is why several devices are supported.

```xml
<usbkeyboard enabled="true">
    <usbkeyboarddevs>
        <usbkeyboarddevpath>/dev/input/by-id/usb-C-Media_…-event-if03</usbkeyboarddevpath>
        <usbkeyboarddevpath>/dev/input/by-id/usb-Another_Device-event-kbd</usbkeyboarddevpath>
    </usbkeyboarddevs>
    <numlockscanid>69</numlockscanid>
</usbkeyboard>
```

Use the stable `/dev/input/by-id/...` paths, not `/dev/input/eventN`, which renumber across reboots. A
single legacy `<usbkeyboarddevpath>` directly under `<usbkeyboard>` is still accepted.

#### The KeyboardCommands Section

Maps key scan codes to actions. Each `<command>` is one action; the nested `<usbkeyboard>` element
binds a key to it.

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

* `action` must be in the list below or the command is disabled with a warning
* `paramname` / `paramvalue` carry the argument for actions that need one — the slot for
  `voicetargetset`, the text for `setcomment`
* `scanid` is the Linux input scan code, **1–255**; `0` or over 255 disables the binding with a
  warning. Find codes with `evtest` or `showkey --scancodes`
* `keylabel` is the label shown in the on-screen key map

Valid `action` values: `channelup`, `channeldown`, `serverup`, `serverdown`, `mute`, `unmute`,
`mute-toggle`, `stream-toggle`, `volumeup`, `volumedown`, `volumerxup`, `volumerxdown`, `volumetxup`,
`volumetxdown`, `volup`, `voldown`, `setcomment`, `transmitstart`, `transmitstop`, `pttkey`,
`soundinterfacepttkey`, `record`, `voicetargetset`, `mqttpubpayloadset`, `changechannel`,
`listentochannelon`, `listentochanneloff`, `gpioinput`, `gpiooutput`, `radiotoggle`, `radionext`,
`radioprev`, `radiovolup`, `radiovoldown`.

**`<ttykeyboard>` elements appear in shipped configs but are not parsed** — only `<usbkeyboard>`
bindings take effect. Terminal keys come from the [bottom CLI](#driving-talkkonnect-from-the-terminal)
instead.

#### The AudioRecordFunction Section

Records channel traffic to disk as raw Opus packets in `.mrec` files, optionally indexed in
MySQL/MariaDB for the web recording player.

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

**`traffic` is the only supported `<recordmode>`** — any other non-empty value logs a warning and
records nothing. `<recordsavepath>` must exist and be writable. `<maxfilesize>` is bytes per file
before rolling over, `<channelbuffersize>` is the in-memory audio buffer in samples (raise it if you
see dropped audio under load) and `<writeflushinterval>` is milliseconds between flushes to disk. The
`<recorddb>` database and its schema must already exist. Recording also starts automatically for a
panic event, and can be toggled at runtime with the `record` keyboard action.

#### Radio Section (SA818 transceiver module)

Controls an SA818/DRA818 VHF/UHF module over serial, which is how talKKonnect becomes an RF gateway or
repeater controller. Channels are numbered and switched with the `m` / `n` keys or the rotary encoder's
`radiochannel` function.

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
                <bandwidth>0</bandwidth>
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

* `<connectchannelid>` is the `id` of the channel programmed at startup
* `<sa818 enabled>` (attribute) gates the section; the separate `<enabled>` element maps to the
  module's power-down control
* `<bandwidth>` is passed straight through as the *band* field of the module's `AT+DMOSETGROUP`
  command, which accepts only `0` (12.5 kHz) or `1` (25 kHz). **The `12500` / `25000` values in the
  sample configs are not values that command accepts**
* `<rxfreq>` / `<txfreq>` are MHz, sent with 4 decimal places. Equal values are simplex, different
  values a repeater split
* `<squelch>` is `0`–`8`, `<volume>` is `1`–`8` (`AT+DMOSETVOLUME`), `<ctcsstone>` is a CTCSS tone
  index and `<dcstone>` a DCS code, `0` meaning off for both
* `<predeemph>`, `<highpass>` and `<lowpass>` are audio filters, `0` off and `1` on (`AT+SETFILTER`)
* `<txpower>` (`H` or `L`) is parsed and logged only — no command is sent for it

**Transmit only on frequencies you are licensed for.**

#### The AnalogRelays Section

Drives relay output pins when traffic appears on a specific channel — for PA zones and sirens.

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

`listenchannel` is matched against the talker's channel exactly. The pin names must be output pins
defined under [`<pins>`](#pins) — use `analogrelay1` / `analogrelay2`, which are the relay names the
validator accepts. Requires `targetboard="rpi"`; on any other target the zones are not created.

---

## The Multimedia Section (Pre-Recorded Announcements / IP-Speaker)

`<multimedia>` sits directly under `<global>`, next to `<hardware>`. Each `<id value="...">` is a named
announcement profile: an optional attention tone followed by an ordered list of media sources, played
out of the local speaker, into the Mumble channel, to the multicast group, or any combination.

A profile is triggered by name, never automatically unless you give it a schedule:

* HTTP: `http://{your-talkkonnect-ipaddress}:8080/?command=announcement&mediaid=main_announcement`
* MQTT: publish `announcement main_announcement` to the subscribed topic
* Bottom CLI or SSH console: `announcement main_announcement`
* `<schedule>` inside the profile, for unattended repetition

The `announcement` action must also be enabled under `<http>` and `<mqtt>` before the first two work.

```xml
<multimedia>
    <id value="main_announcement" enabled="true">
        <params>
            <announcementtone file="/path/to/announcement-01.wav" volume="10" blocking="true" enabled="true"/>
            <localplay>true</localplay>
            <gpio name="multimedia_active_led" enabled="true"/>
            <predelay value="1" enabled="true"/>
            <postdelay value="1" enabled="true"/>
            <playintostream>true</playintostream>
            <streamvolume>50</streamvolume>
            <voicetarget id="0">false</voicetarget>
            <multicast>false</multicast>
        </params>
        <schedule intervalsecs="0" enabled="false"/>
        <media>
            <source name="1st-song" file="/path/to/song.mp3" volume="10" duration="0" offset="0" loop="1" blocking="true" enabled="true"/>
        </media>
    </id>
</multimedia>
```

* the `value` attribute is the profile name used as the `mediaid` when triggering the announcement, and
  `enabled` on the `<id>` turns the whole profile on or off. **`value` must not be empty**
* `announcementtone` is an attention tone played before the media sources; `file` may be an absolute
  path or an `http`, `https` or `rtsp` url, `volume` is 0–100 and `blocking` waits for the tone to
  finish before the first source starts
* `localplay` set to true plays the profile out of the local speaker or amplifier using `ffplay`
* `playintostream` set to true transmits the profile into the currently joined Mumble channel so
  everyone in the channel hears it, and `streamvolume` (0–100) is the default volume used for that
* `gpio name=...` if enabled is driven high for the entire announcement, for switching an external
  amplifier or an attention light; the pin drops as soon as playback returns, so set `blocking="true"`
  on the tone and the sources if you want it to track the audio accurately
* `predelay` and `postdelay` are in **seconds** and add a pause before and after the announcement
* `voicetarget` decides where an into-stream announcement is sent. Left as false the announcement goes
  to the channel you are joined to, temporarily bypassing any voice target that happens to be
  selected. Set to true with an `id` of 1–31 it shouts the announcement at that `<voicetargets>` slot
  of the active account (zone paging), and set to true with `id="0"` it follows whichever voice target
  is already active. The previous routing is restored when the announcement finishes. An out-of-range
  `id` is reset to `0` with a warning, and this tag has no effect on `localplay`
* `multicast` set to true also sends the profile to the RTP multicast group configured in the
  [Multicast section](#the-multicast-section-rtp-to-ip-speakers-and-sip-phones), for paging hardware IP
  speakers and SIP desk phones. It is independent of `localplay` and `playintostream`, so a profile can
  page the speakers only, go into Mumble only, or do both at once; it needs
  `<multicast enabled="true">` and `ffmpeg` on the `PATH`
* `schedule intervalsecs="N" enabled="true"` replays the profile every N seconds for unattended
  announcements. Enabled with `intervalsecs` of 0 or less is disabled with a warning

Each `<source>` under `<media>` is played in the order it appears in the file:

* `file` is an absolute path or an `http`, `https` or `rtsp` url, so a source can also be an internet
  radio stream
* `volume` is 0–100 and applies to both local and into-stream playback, overriding `streamvolume` for
  that source
* `offset` in seconds seeks into the file before playing, for skipping a long intro
* `duration` in seconds cuts playback short, which is how you page a fixed slice out of a long file or
  a stream
* `loop` repeats the source, **clamped to 3** so a typo cannot occupy the channel indefinitely
* `blocking` set to true finishes the source before the next one starts. Set it on every source unless
  you actually want them mixed on top of each other, since local playback is otherwise fired off in
  parallel
* `enabled` skips the source when false

**A profile must have at least one of `<localplay>`, `<playintostream>` or `<multicast>` set to
true**, or it is disabled with a warning at startup, since it would have nowhere to play. Profiles with
an empty `value` or with files that cannot be found are disabled the same way, so watch the `warn:`
lines in the log.

Full detail: [CONFIGURATION.md §7](./CONFIGURATION.md#7-globalmultimedia).

---

## The Internet Radio Section

`<global><Radio>` — note the **capitalised tag names** throughout this section, which differ from the
rest of the file and are case-sensitive. It plays internet audio streams through `ffmpeg` to ALSA in
the background, and gets out of the way when Mumble audio arrives.

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

* `<Enabled>` is an element here, not an attribute
* `<InterruptionMode>` is what happens when Mumble audio arrives: `stop`, `pause`, or `duck` (lower the
  volume to `<DuckVolumePercent>`). Invalid values warn and fall back to `stop`
* `<AutoResumeDelay>` is the silence in seconds after which the radio resumes; 0 or negative becomes 15
* `<MasterVolume>` and `<DuckVolumePercent>` are 1–100; an out-of-range master volume resets to 50
* `<AlsaDevice>` is the output device, e.g. `plughw:1,0`. `<FFmpegPath>` is required — it is the
  playback engine
* `<StreamRetrySecs>` is the reconnect delay after a stream drops; 0 or negative becomes 5
* `<AnnounceStationTTS>` speaks the station name on change, using
  [`<ttsmessages>`](#ttsmessages-section)
* `<YoutubeMusicPlayback>` enables YouTube and YouTube Music URLs through `yt-dlp`. If it is on and
  `<YtDlpPath>` is not found on disk or on `PATH`, you get a warning
* per station, `<Backend>` empty or `auto` decides from `YoutubeMusicPlayback` plus URL detection,
  `youtube` forces `yt-dlp`, and `http`, `direct` or `ffmpeg` force plain `ffmpeg -i`. For YouTube
  Music use a specific track or watch URL, not the site home page
* enabling the section with no stations warns and falls back to built-in demo stations

Control it with the `v` key, or `radiotoggle`, `radionext`, `radioprev`, `radiovolup` and
`radiovoldown` over the HTTP API, MQTT and the keyboard. The shipped configs omit the `radiovolup` /
`radiovoldown` `<command>` entries, so add them to reach those two over HTTP.

Full detail: [CONFIGURATION.md §8](./CONFIGURATION.md#8-globalradio-internet-streaming-radio).

---

## Live reload — what can change at runtime

`Ctrl-B`, or a save from the `/config` web editor, re-reads the file but applies only an allow-list of
settings. Everything else keeps its startup value until you restart.

**Applied on live reload:** `<settings>` `loglevel`, `cancellablestream`, `streamsendmessage`,
`repeattxtimes`, `repeattxdelay` and `simplexwithmute`; `<channelscan>`, `<beacon>`, `<tts>`,
`<sounds>` (event sounds are re-preloaded); `<remotecontrol>` `<http enabled>` and its `<command>`
list, `<uistatus>`, `<networkacl>` and the MQTT `<commands>`; `<printvariables>`, `<ttsmessages>`,
`<ignoreuser>`; `<multicast>` (re-applied to the running sender); `<panicfunction>`; `<keyboard>`
commands; `<multimedia>`; `<Radio>`.

**Requires a restart:** accounts and the server list, all audio device settings, `<logging>` and
`<logfilenameandpath>`, `<singleinstance>`, `<remotesshconsole>`, `<autoprovisioning>`, `<smtp>`,
`<memorychannels>`, `<presetvoicetargets>`, the MQTT `<settings>`, and **every** `<hardware>` section —
GPIO, LCD, OLED, GPS, Traccar, USB keyboard devices, audio recording, the SA818 radio and analog
relays.

The `cfg` commands give finer-grained runtime edits:

```
cfg keys                                       # list settable paths
cfg list                                       # show current values
cfg set global.software.settings.loglevel debug
cfg set accounts.account.0.username mycall
cfg save                                       # write back to talkkonnect.xml
cfg restart                                    # re-exec talkkonnect
```

Some runtime data — notably the account slices built at startup — only takes full effect after
`cfg restart`.

---

## Startup validation and what it does to your config

`CheckConfigSanity` runs on every load, and its output is worth reading on first boot:

```
info: Starting XML Configuration Sanity and Logical Checks
warn: Config Error [Section GPIO] Enabled GPIO Name listening Pin Number 14 Invalid Name
warn: Non-Critical Errors Found In talkkonnect.xml config file please fix errors …
```

* `warn:` — the value was clamped or the feature was switched off. talKKonnect keeps running, possibly
  without the feature you thought you had configured
* `alert:` — fatal, talKKonnect stops. The two fatal cases are **no default/enabled account** and an
  unusable `<nextserverindex>`
* `info: Finished XML Configuration Sanity and Logical Checks Without Any Alerts/Errors/Warnings` — a
  clean file

Re-run it at any time with the `h` key.

The corrections it makes automatically:

| Condition | Result |
| --- | --- |
| `<outputdeviceshort>`, `<outputvolcontroldevice>`, `<outputmutecontroldevice>` empty | Copied from `<outputdevice>` |
| `<logfilenameandpath>` empty and logging ≠ `screen` | First writable of config dir, CWD, `/var/log`, `/tmp` |
| `<voiceactivitytimermsecs>` < 200 | Set to 200 |
| `<volupstep>` / `<voldownstep>` = 0 | Set to `+1` / `-1` |
| `<nextserverindex>` > number of default accounts | Reset to 0, fatal if there are no accounts |
| Channel scan dwell/hang < 500 ms | 500 ms used |
| Sound file missing, and not an `http`/`rtsp` URL | That sound disabled |
| Sound `volume="0"` | That sound disabled |
| Media `loop` > 3 | Clamped to 3 |
| Multimedia profile with no output destination | Profile disabled |
| Multicast group / port / codec / packetms / ttl / volume / hangover invalid | Clamped to defaults |
| Beacon, SMTP, SSH console, autoprovisioning or MQTT missing required children | Feature disabled |
| GPIO pin with a bad direction, device, name, number or chipid | That pin disabled |
| LCD interface type or required pins invalid | LCD disabled |
| GPS port missing, bad baud, both parities, zero stop/data bits | GPS disabled |
| `<ignoreuserregex>` shorter than 4 characters | Ignore-user disabled |
| Internet radio `InterruptionMode` / `MasterVolume` invalid | Reset to `stop` / `50` |
| Keyboard action not recognised, or scan id 0 or > 255 | That binding disabled |
| `<memorychannels>` / `<presetvoicetargets>` with an unrecognised `gpioname` | The **whole section** disabled |

---

## Known quirks and traps

These are real behaviours of the current code. Knowing them saves hours.

1. **Unknown tags are silently ignored.** Misspell a tag and the value is never read, with no warning.
   `<oledttf>` and `<ttykeyboard>` in the shipped configs are exactly this — they parse fine and do
   nothing.

2. **`<tts>` action names are case-sensitive and must match exactly.** The shipped config contains
   `muteSpeaker`, `unmuteSpeaker` and `currentvolumelevel`; the code asks for `mutespeaker`,
   `unmutespeaker`, `currentrxvolumelevel` and `currenttxvolumelevel`. Prompts with the wrong spelling
   never play. Several other shipped entries (`participants`, `leftjoinedchannel`, `message`,
   `listentochannelson/off`, `printxmlconfig`) are not requested by any code path.

3. **`<sounds><input>` event names need the `io` / `usb` prefix.** The shipped config uses short names
   like `txpttstart` and `volup`; the code asks for `iotxpttstart` and `iovolup`. Only
   `txtogglestart` matches without a prefix.

4. **`<sound event="alert">`** appears in the shipped config but is never requested. The panic alert
   sound comes from `<panicfunction><filenameandpath>` instead.

5. **Some working GPIO names are rejected by the validator.** `listening`, `tracking`, `mqtt0`,
   `mqtt1` and the five `internetradio*` names have handlers in `gpio.go` but are not in the
   validator's allow-list, so enabling such a pin logs `Invalid Name` and disables it.
   `device="radiomodule"` is likewise not a valid device value.

6. **`<listening>` and `<comment><commentbuttonpin>` are ignored.** Button binding is by the pin's
   `name` attribute under `<pins>` (`comment`, `listening`), never by these fields.

7. **The LCD backlight timer has an inverted-looking guard.** It is disabled unless *both*
   `<oled enabled>` and `<lcd enabled>` are true, and it is also forced off when `targetboard` is not
   `rpi`.

8. **`<ttsmessages><predelay>` / `<postdelay>` produce no useful delay.** The value is treated as
   nanoseconds there. The `<multimedia>` equivalents are correctly interpreted as seconds — use those
   when you need a real gap.

9. **`<max7219>` children must be attributes.** The parser declares `max7219cascaded`, `spibus`,
   `spidevice` and `brightness` as attributes; writing them as child elements, as some sample configs
   do, leaves them at zero.

10. **HTTP and MQTT action names differ.** MQTT uses `muteunmute` with a `mute`/`unmute`/`toggle`
    argument where HTTP uses three separate actions, and MQTT has no `ttsannouncement`, `showversion`,
    `joinchannel`, `whisperuser`, `setrxvolume` or `listapi`. Declaring an MQTT command that has no
    handler passes the "is it defined?" check and then fails at dispatch time.

11. **An HTTP command must be declared in the XML *and* be a built-in.** Adding `radiovolup` /
    `radiovoldown` `<command>` entries is needed to reach those handlers — the shipped config omits
    them.

12. **`<tokens enabled="…">` is read from the `<account>` element, not `<tokens>`.** The attribute is
    declared on the account struct, so the flag on `<tokens>` has no effect and tokens are always sent.

13. **`<hardware><radio>` and `<global><Radio>` are different features.** The first is the SA818 RF
    module; the second is internet streaming radio with capitalised tag names.

14. **The MQTT `relay` command does not work as written.** It accepts a relay number of `1` or `2` and
    drives a pin named `relay1` / `relay2` — but the pin validator only allows `relay0`, so such a pin
    is disabled at startup. Separately, the handler reads the fourth word of a three-word payload, so
    the argument lookup is out of range. Use `<analogrelays>` or a GPIO output driven by another action
    instead.

15. **The menu banner has two wrong labels.** `<1>` does nothing (use `?`), and `<x>` is *previous*
    server, not next.

---

## Minimal working config

Enough to connect and talk, with nothing hardware-specific. Build up from here one section at a time,
reading the sanity-check output after each change.

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

---

## Where to go next

* [talkkonnect.xml Configuration Manual](./CONFIGURATION.md) — the exhaustive per-tag reference
* [Using the Terminal CLI](./terminal-cli.md) — the bottom prompt and the SSH console
* [HTTP API](./api.md) — request and response format for every action
* [Voice Targets (Whisper and Shout)](./voice-targets.md)
* [Multicast to IP Speakers and SIP Phones](./multicast.md)
* [Functionality](./functionality.md) and [Hardware Builds](./hardware-builds.md)
