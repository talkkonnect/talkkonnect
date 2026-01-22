#!/usr/bin/env bash
set -euo pipefail

# Arch / Manjaro / EndeavourOS
# Installs build-time deps + common runtime tools.

sudo pacman -Syu --needed --noconfirm \
  base-devel git go pkgconf \
  openal opus alsa-lib \
  ffmpeg mplayer \
  ca-certificates curl
