#!/usr/bin/env bash
set -euo pipefail

# scripts/deps/opus.sh
#
# Build and install libopus from the official Xiph release tarball when the
# system package is older than the required minimum version.
#
# Usage:
#   sudo ./scripts/deps/opus.sh
#
# Environment overrides:
#   OPUS_VERSION   - target version (default: 1.6.1)
#   OPUS_PREFIX    - install prefix (default: /usr/local)
#   OPUS_MIN_VERSION - skip install when pkg-config reports >= this (default: 1.6.1)

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"

OPUS_VERSION="${OPUS_VERSION:-1.6.1}"
OPUS_MIN_VERSION="${OPUS_MIN_VERSION:-1.6.1}"
OPUS_PREFIX="${OPUS_PREFIX:-/usr/local}"
OPUS_SHA256="${OPUS_SHA256:-6ffcb593207be92584df15b32466ed64bbec99109f007c82205f0194572411a1}"
OPUS_URL="https://downloads.xiph.org/releases/opus/opus-${OPUS_VERSION}.tar.gz"

say() { printf '%s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

version_ge() {
  local installed="$1"
  local required="$2"
  if [[ "$(printf '%s\n' "$required" "$installed" | sort -V | head -n1)" == "$required" ]]; then
    return 0
  fi
  return 1
}

installed_version() {
  if pkg-config --exists opus 2>/dev/null; then
    pkg-config --modversion opus
    return 0
  fi
  echo "0"
}

if [[ $EUID -ne 0 ]]; then
  die "Please run with sudo (this script installs libopus to ${OPUS_PREFIX})"
fi

current="$(installed_version)"
if version_ge "$current" "$OPUS_MIN_VERSION"; then
  say "libopus ${current} already satisfies minimum ${OPUS_MIN_VERSION}; skipping source install."
  exit 0
fi

say "Installing libopus ${OPUS_VERSION} (current: ${current}, required: >= ${OPUS_MIN_VERSION})"

export DEBIAN_FRONTEND=noninteractive
if command -v apt-get >/dev/null 2>&1; then
  apt-get update -y
  apt-get install -y --no-install-recommends \
    build-essential pkg-config curl ca-certificates
fi

workdir="$(mktemp -d)"
trap 'rm -rf "$workdir"' EXIT

tarball="${workdir}/opus-${OPUS_VERSION}.tar.gz"
say "Downloading ${OPUS_URL} ..."
curl -fsSL -o "$tarball" "$OPUS_URL"

actual_sha256="$(sha256sum "$tarball" | awk '{print $1}')"
if [[ "$actual_sha256" != "$OPUS_SHA256" ]]; then
  die "SHA256 mismatch for opus-${OPUS_VERSION}.tar.gz (got ${actual_sha256})"
fi

tar -xf "$tarball" -C "$workdir"
srcdir="${workdir}/opus-${OPUS_VERSION}"
[[ -d "$srcdir" ]] || die "Expected source directory not found: ${srcdir}"

pushd "$srcdir" >/dev/null
./configure --prefix="${OPUS_PREFIX}" --disable-doc
make -j"$(nproc 2>/dev/null || echo 2)"
make install
popd >/dev/null

if command -v ldconfig >/dev/null 2>&1; then
  ldconfig
fi

# Prefer /usr/local libopus at runtime when a distro copy is also present.
if [[ -d "${OPUS_PREFIX}/lib" ]]; then
  conf_file="/etc/ld.so.conf.d/talkkonnect-opus-local.conf"
  if [[ ! -f "$conf_file" ]] || ! grep -qx "${OPUS_PREFIX}/lib" "$conf_file" 2>/dev/null; then
    echo "${OPUS_PREFIX}/lib" > "$conf_file"
    ldconfig
  fi
fi

export PKG_CONFIG_PATH="${OPUS_PREFIX}/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
installed="$(PKG_CONFIG_PATH="${PKG_CONFIG_PATH}" pkg-config --modversion opus 2>/dev/null || true)"
if ! version_ge "${installed:-0}" "$OPUS_MIN_VERSION"; then
  die "libopus install finished but pkg-config reports '${installed:-none}' (need >= ${OPUS_MIN_VERSION}). Check PKG_CONFIG_PATH includes ${OPUS_PREFIX}/lib/pkgconfig"
fi

say "libopus ${installed} installed to ${OPUS_PREFIX}"
