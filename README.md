# talKKonnect

### A Headless Mumble Client that can function as a Talkie / Intercom / Gateway for Linux or Single Board Computers (IP Radio / Push-to-Talk)

*If you use talkkonnect, please let us know your use case by sending us pictures, and please also STAR the
talkkonnect/talkkonnect repo on github.com!*

---

* [What is talKKonnect?](#what-is-talkkonnect)
* [Installation / Getting Started](#installation--getting-started)
* [Functionality and Configurability](#functionality-and-configurability)
* [Using the Terminal CLI](#using-the-terminal-cli)
* [Voice Targets (Whisper and Shout)](#voice-targets-whisper-and-shout)
* [Multicast to IP Speakers and SIP Phones](#multicast-to-ip-speakers-and-sip-phones)
* [HTTP API Specification / Building Your Own UI](./docs/api.md)
* [Configuring and Running talKKonnect](./docs/running-talkkonnect.md)
* [Extra Multimedia Features (IP-Speaker)](./docs/functionality.md#extra-multimedia-features-ip-speaker)
* [Software Configurable Features](./docs/functionality.md#software-configurable-features)
* [Additional Optional Hardware and Precautions](./docs/functionality.md)
* [Common Information for all the Pre-Made Images For Various Hardware Configurations](./docs/hardware-builds.md#common-information-for-the-all-the-pre-made-images-for-various-hardware-configurations)
* [Why Was talKKonnect created?](./docs/history.md#why-was-talkkonnect-created)
* [Questions & Contributing](#questions--contributing)
* [License](#license)

----

## What is talKKonnect?

<a id="what-is-talkkonnect"></a>

[talKKonnect](https://www.talkkonnect.com) is an open-source customizable, headless, self-contained Mumble Push to Talk (
PTT) client.

It was designed for Linux single-board computers (SBCs) such as the Raspberry Pi and Orange Pi. It works equally well on
any reasonably modern Linux distribution.

#### Functionality for SBCs or as a Hardware Appliance:

talKKonnect offers a flexible form factor with an LCD/OLED display, channel and volume control, making it ideal for
group communication scenarios. Common use cases include amateur radio enthusiasts, ad-hoc group communications, and
replacing expensive commercial intercom
systems. [Read more about talKKonnect as an appliance and some potential use-cases.](./docs/appliances.md)

## Installation / Getting Started

Diving right in? talKKonnect works on a variety of devices and form factors, and even provides pre-built images for use
on common Raspberry Pi or similar single board computer (SBC) architectures.

* [General Installation instructions](./docs/getting-started.md)
* [Raspberry Pi / Other Pre-made Image Instructions](./docs/hardware-builds.md)
* [Configuration and Running](./docs/running-talkkonnect.md)

----

## Functionality and Configurability

Because talKKonnect was originally created as software to power hardware-based IP communication devices, it has
extensive optional functionality and configurability for several common hardware
devices. [See a full overview of talKKonnect configuration and support details](./docs/functionality.md)

----

## Using the Terminal CLI

When talKKonnect runs in a terminal it keeps a command prompt pinned to the bottom row while the log
scrolls above it. Everything the device can do is reachable from there: single keys for the common
actions, and typed commands for the rest. The same prompt is available remotely over SSH when
`<remotesshconsole enabled="true">` is set in `talkkonnect.xml` — the sample configs listen on port 9999,
so `ssh -p 9999 user@device`.

Type `?`, `help` or `menu` at the prompt for the full banner. In short:

| Input | What it does |
| --- | --- |
| `2` `3` | Channel up / down |
| `4` `5` `6` | Mute-unmute speaker, digital volume up / down |
| `7` `8` | Start / stop transmitting |
| `9` `0` | List online users, show uptime |
| `r` `v` `b` | Channel scan on/off, online radio on/off, playback/stop stream |
| `d` `h` `u` | Dump the XML config, run the config sanity checker, show the version |
| `a` `o` `g` | List API commands, ping the configured servers, show the GPS position |
| `c` / `clear` | Clear the screen and repaint the prompt |
| `q` / `quit` | Close the CLI, leaving talKKonnect running |
| `...` | Shut talKKonnect down |

Typed commands, with **Tab completion** throughout:

* `cfg keys` / `cfg list` / `cfg set <path> <value>` / `cfg save` / `cfg restart` — read and change any
  setting in the running configuration by its dotted path, e.g.
  `cfg set global.software.settings.loglevel debug`. `cfg save` writes it back to `talkkonnect.xml`
* `vt …` — voice targets, see [the section below](#voice-targets-whisper-and-shout)
* `mc …` — multicast output, see [the section below](#multicast-to-ip-speakers-and-sip-phones)
* any HTTP API action name, typed directly — for example `announcement main_announcement`,
  `voicetargetset 3` or `showuptime`. These go through the same allow-list as the HTTP API, so the action
  must be enabled under `<http>` in the XML

Set `TALKKONNECT_NO_BOTTOM_CLI=1` to turn the bottom prompt off when running talKKonnect under a
harness, in a pipeline, or anywhere you want plain unadorned log output. It is also skipped
automatically when there is no terminal, such as under `--daemon` or systemd.

----

## Voice Targets (Whisper and Shout)

A voice target sends your transmitted audio to specific users or a specific channel instead of the
channel you are joined to — Mumble calls this whispering (to users) and shouting (to a channel). It is
how one device pages a single zone rather than everybody.

Targets are numbered 1 to 31 and are defined per account in `talkkonnect.xml`:

```xml
<voicetargets>
    <id value="1">
        <users>
            <user>zoran-laptop</user>
        </users>
    </id>
    <id value="3">
        <channels>
            <channel>
                <name>Zone-One</name>
                <recursive>true</recursive>
                <links>true</links>
                <group></group>
            </channel>
        </channels>
    </id>
</voicetargets>
```

You do not have to edit the file by hand. The `vt` command manages targets from the CLI or the SSH
console:

| Command | What it does |
| --- | --- |
| `vt list` | Every configured target, which one is active, and any user or channel that is not on the server right now |
| `vt set <id>` | Activate a configured target |
| `vt clear` | Back to normal channel speech (the same as `voicetargetset 0`) |
| `vt next` / `vt prev` | Step through the configured targets |
| `vt whisper <user>` | Whisper to one online user, no config entry needed |
| `vt add <id> user <name> …` | Add users to a target and write it to `talkkonnect.xml` |
| `vt add <id> channel <name> [recursive=true] [links=true] [group=<name>]` | Add a channel to a target |

`vt add` edits only the `<voicetargets>` element, leaves the rest of the file — comments, formatting and
tag order — exactly as it was, keeps a `.bak` copy, and applies the change to the running configuration
immediately, so a new target is usable without a restart.

Targets can also be selected by GPIO preset buttons, a rotary encoder, the `voicetargetset` action over
the HTTP API and MQTT, and per-announcement with the `<voicetarget>` tag of a `<multimedia>` profile.
[Full details and every tag](./docs/running-talkkonnect.md#voice-targets-and-the-vt-command)

----

## Multicast to IP Speakers and SIP Phones

talKKonnect can re-transmit the audio it receives from Mumble to a multicast group as RTP, so hardware IP
PA speakers (CyberData, Algo, Barix, Advanced Network Devices) and SIP desk phones such as Yealink can
hear a Mumble channel without being Mumble clients. This turns a talKKonnect node into a gateway between
a Mumble channel and an existing paging system.

Audio goes out as **8 kHz mono G.711, 20 ms per packet** — the format those devices decode — which is the
same wire format [gochimesd](https://github.com/talkkonnect/gochimesd) sends, so one receiver
configuration serves both. Several people talking at once are mixed into a single RTP stream, and nothing
is sent while the channel is quiet.

Add a `<multicast>` section inside `<global><software>` of `talkkonnect.xml`:

```xml
<multicast enabled="true">
    <group>239.0.1.10</group>
    <port>5004</port>
    <codec>pcmu</codec>
    <ttl>1</ttl>
    <interface>eth0</interface>
    <packetms>20</packetms>
    <allchannels>false</allchannels>
    <include>
        <!-- empty means every talker in the channel -->
    </include>
    <exclude>
        <!-- <user>noisybox</user> -->
    </exclude>
</multicast>
```

* `include` and `exclude` decide **whose** audio is carried. An empty `include` carries every talker in
  the channel; a name in `exclude` is never carried, and exclude wins
* `allchannels` left false carries only the channel talKKonnect is joined to, so a monitored channel is
  not re-broadcast over the PA by accident
* `codec` should stay `pcmu` (G.711 u-law) unless every listener on the group is known to support
  something else. `l16` is uncompressed PCM on a dynamic payload type and G.711-only receivers drop it
  silently — the stream looks correct on the network while the speaker stays mute
* a `<multimedia>` profile with `<multicast>true</multicast>` pages the same group with a pre-recorded
  announcement, independently of local playback and playing into the Mumble channel

Control it at run time with `mc status`, `mc on`, `mc off` and `mc toggle` in the CLI or SSH console, with
the `multicaston`, `multicastoff` and `multicasttoggle` actions over the HTTP API and MQTT, and watch it
in the `multicast` object of the `/uistatus` telemetry snapshot. To check the stream from another machine
on the same subnet: `ffplay rtp://@239.0.1.10:5004`.

[Every tag, and the run time and network caveats](./docs/running-talkkonnect.md#the-multicast-section-rtp-to-ip-speakers-and-sip-phones)

----

## HTTP API and Building Your Own User Interface

talKKonnect exposes two HTTP interfaces: a JSON telemetry endpoint (`/uistatus`) that publishes the
complete live state of the client, and a control API (`/?command=...`) that executes remote control
commands. Together they let you build any user interface you like, in any language, with no changes to
talKKonnect itself.

[**Read the full HTTP API specification**](./docs/api.md) — it documents every endpoint, field, command
and status code, and includes build guides for the two interfaces people most often want:

* **A framebuffer screen** — a full-screen status display drawn straight to a Linux framebuffer, HDMI
  monitor or SPI TFT panel on the talkkonnect device itself, with far more room than a 4x20 character
  LCD. [See the guide](./docs/api.md#building-a-framebuffer-screen-client)
* **A web UI** — a browser dashboard showing live state and sending control
  commands. [See the guide](./docs/api.md#building-a-web-ui)

Both are small, well-specified projects that an **AI programming agent** (such as Claude Code) can write
almost end to end. The specification includes ready-to-use starting prompts, guidance on iterating with an
agent, and a review checklist covering the mistakes that actually show up in generated
code. [See the guide](./docs/api.md#building-a-ui-with-an-ai-programming-agent)

If you build a UI this way, please share it with us — we would like to feature community-built interfaces.

## History and Features of talKKonnect

<a id="some-interesting-features"></a>

This project was created by [Suvir Kumar](https://www.linkedin.com/in/suvir-kumar-51a1333b) as a fork
of [talkiepi](http://projectable.me/) by Daniel Chote which was, in turn, a fork
of [barnard](https://github.com/layeh/barnard) a text based mumble client. talKKonnect was developed
using [golang](https://golang.org/) and based on [gumble](https://github.com/layeh/gumble) library by Tim
Cooper. [Read the full history and background of talKKonnect](./docs/history.md)

See [a video explanation of the history and reasons for creating talkkonnect](https://youtu.be/nLmHM48SqFs)

----

## Questions & Contributing

We invite interested individuals to provide feedback and improvements to the project.
You can help with creating youtube videos, documentation, programming, testing and/or feedback of your user experience.

To speak to us, connect with a standard mumble client (Android/iPhone/Windows/Linux), or with talkkonnect itself, to our
community server at mumble.talkkonnect.com port 64738. Use any unique username; no password is required. We are standing
by usually on the HAM-CB channel.

Currently we do not have a WIKI, so send feedback to <suvir@talkkonnect.com> or open an Issue on GitHub.

Please visit our [blog](https://www.talkkonnect.com) for updates on the project,
our [github](https://github.com/talkkonnect) for the latest source code,
and our [facebook](https://www.facebook.com/talkkonnect) page for future updates and information.

Thank you all for your kind feedback sent along with some pictures and use cases for talkkonnect.

----

## License

[talKKonnect](https://www.talkkonnect.com) is open source and available under
the [Mozilla Public License 2.0](./LICENSE).

<suvir@talkkonnect.com> Updated 29/07/2026. talkkonnect version 4.18.01 is the latest release as of this writing.


