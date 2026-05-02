#!/usr/bin/env bash
set -euo pipefail

# scripts/deps/debian.sh
#
# Install build/runtime dependencies for talkkonnect
# on Debian / Ubuntu / Raspbian family systems.
#
# This script:
#   - DOES NOT run dist-upgrade / full-upgrade
#   - Only installs required packages
#   - Is safe to re-run
#   - Requires sudo

say() { printf '%s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

if [[ $EUID -ne 0 ]]; then
  die "Please run with sudo (this script installs system packages)"
fi

say "Installing talkkonnect dependencies (Debian-family)"
say

export DEBIAN_FRONTEND=noninteractive

say "Updating package index..."
apt-get update -y

say "Installing build and runtime dependencies..."
apt-get install -y --no-install-recommends \
  build-essential \
  pkg-config \
  git \
  ca-certificates \
  curl \
  libopenal-dev \
  libopus-dev \
  libasound2-dev \
  ffmpeg \
  mplayer

say
say "Dependency installation complete."
say "You can now run:"
say "  make build"
