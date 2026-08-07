# Getting Started with talKKonnect

### Introduction
talKKonnect can be run on a variety of modern architectures. It was originally designed to run on single-board computers
(SBCs) such as the Raspberry Pi and Orange Pi, but it can now also be compiled and run on an x86 Linux PC — it has been
tested on bare metal x86 hardware and also runs as a virtual machine in a Proxmox virtualization environment. This page
has instructions for getting and running talKKonnect directly from the repository.

It is even easier, if you already have hardware such as a Raspberry Pi, to put talKKonnect on your device. Take a look at [our instructions for
using pre-made hardware builds](./hardware-builds.md) if you have one. Note that those ready-made SD card images are
talKKonnect **version 2** and are very old — they are a quick way to preview and test the features of a much earlier
talKKonnect without building anything. We are currently at **version 4**, and to run version 4 you have to build from
source using the instructions on this page.

### Build from Source
talKKonnect has been reliably tested on both Debian-based Linux distributions (especially Raspbian), as well as
Arch-based.

#### The easy way: the tk-build-v1.sh script

On a **fresh install of Debian or Raspberry Pi OS (minimal)** you do not have to do any of the steps below by hand.
The [`scripts/tk-build-v1.sh`](../scripts/tk-build-v1.sh) bash script installs talKKonnect from source for you. It:

* installs all the build and runtime dependencies (`git`, `pkg-config`, `gccgo`, PulseAudio, ALSA, `ffmpeg`, libopus and friends)
* installs the latest Go toolchain for your architecture (64-bit and 32-bit ARM, and 64-bit and 32-bit x86 PC are all handled)
* creates the `talkkonnect` user and adds it to the `audio`, `dialout`, `gpio` and other required groups
* sets up `GOPATH`, `GOBIN` and the `tk` shell alias in your `.bashrc`
* clones the talKKonnect source into `/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect`
* builds the binary to `/home/talkkonnect/bin/talkkonnect` and prints a build report

Run it as root:

```shell
wget https://raw.githubusercontent.com/talkkonnect/talkkonnect/main/scripts/tk-build-v1.sh
sudo bash tk-build-v1.sh
```

If you have already cloned the repository, just run it from the `scripts` directory instead:

```shell
sudo bash scripts/tk-build-v1.sh
```

The whole run is logged to `/var/log/tk-install.log`. When it finishes you still need a configuration file — copy one of
the [sample configs](https://github.com/talkkonnect/talkkonnect/tree/main/sample-configs) to the program directory as
`talkkonnect.xml`, or run [`scripts/tk-post-install.sh`](../scripts/tk-post-install.sh), which fetches a default config
and prompts you for your Mumble server details. See
[Configuration and Running](./running-talkkonnect.md) for what to put in it.

> **Note:** the script overwrites an existing talKKonnect source tree and binary under `/home/talkkonnect`, so back up
> your `talkkonnect.xml` before re-running it on an existing installation.

#### The manual way

If you would rather do it yourself, or you are on a distribution the script does not cover (it is Debian/`apt` based),
build it by hand.

Requirements:
* [Go 1.25.0 or newer](https://go.dev/doc/install)
* **libopus 1.6.1** development headers (`libopus-dev` or built from source)

Debian and Raspbian often ship an older `libopus-dev`. The dependency installer builds and installs libopus 1.6.1 automatically when needed. You can also run it directly:

```shell
sudo make deps-opus
```

In Linux (and maybe macOS) talKKonnect can be used simply by cloning the repository and building:
```shell
git clone https://github.com/talkkonnect/talkkonnect.git
cd talkkonnect
make deps # Will attempt to identify your distribution and install build dependencies
make build # Will build talKKonnect.
```

If you have any trouble with the `make deps` command, you may need to try explicitly targeting your specific distribution:
1. Arch (and Manjaro, EndeavourOS, etc.): `make deps-arch`
2. Debian (and Ubuntu, Raspbian, and so on): `make deps-debian`


