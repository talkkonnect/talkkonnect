# gopus (talkkonnect fork)

Go bindings for the [Opus](https://www.opus-codec.org/) audio codec.

This fork links against system **libopus** via `pkg-config` on all platforms
(amd64, 386, ARM, etc.). Requires libopus **1.6.1** or newer at build and
runtime.

## Requirements

- cgo
- libopus development headers (`libopus-dev` or built from source via `scripts/deps/opus.sh`)
- pkg-config

## License

Public domain (Unlicense). See upstream layeh/gopus.
