# Getting Started with talKKonnect

### Introduction
talKKonnect can be run on a variety of modern architectures. This page has instructions for getting and running
talKKonnect directly from the repository.

It is even easier, if you already have hardware such as a Raspberry Pi, to put talKKonnect on your device. Take a look at [our instructions for
using pre-made hardware builds](./hardware-builds.md) if you have one.

### Build from Source
talKKonnect has been reliably tested on both Debian-based Linux distributions (especially Raspbian), as well as
Arch-based.

Requirements:
* [Go 1.25.0 or newer](https://go.dev/doc/install)

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


