# talKKonnect

### A Headless Mumble Client that can function as a Talkie / Intercom / Gateway for Linux or Single Board Computers (IP Radio / Push-to-Talk)

*If you use talkkonnect, please let us know your use case by sending us pictures, and please also STAR the
talkkonnect/talkkonnect repo on github.com!*

---

* [What is talKKonnect?](#what-is-talkkonnect)
* [Installation / Getting Started](#installation--getting-started)
* [Functionality and Configurability](#functionality-and-configurability)
* [Using the Terminal CLI](./docs/terminal-cli.md)
* [Voice Targets (Whisper and Shout)](./docs/voice-targets.md)
* [Multicast to IP Speakers and SIP Phones](./docs/multicast.md)
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
actions, typed commands with Tab completion for the rest, and the same prompt remotely over
SSH. [Read the terminal CLI reference](./docs/terminal-cli.md)

----

## Voice Targets (Whisper and Shout)

A voice target sends your transmitted audio to specific users or a specific channel instead of the
channel you are joined to — whispering and shouting in Mumble terms. It is how one device pages a single
zone rather than everybody, and the `vt` command sets targets up without hand-editing the
XML. [Read about voice targets and the vt command](./docs/voice-targets.md)

----

## Multicast to IP Speakers and SIP Phones

talKKonnect can re-transmit the audio it receives from Mumble to a multicast group as RTP, so hardware IP
PA speakers (CyberData, Algo, Barix, Advanced Network Devices) and SIP desk phones such as Yealink can
hear a Mumble channel without being Mumble clients. This turns a talKKonnect node into a gateway between
a Mumble channel and an existing paging system. [Read about multicast output](./docs/multicast.md)

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


