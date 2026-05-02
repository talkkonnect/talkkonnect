#!/usr/bin/env bash
set -euo pipefail

# scripts/deps/install.sh
#
# Purpose:
#   Install build/runtime dependencies for talkkonnect on supported Linux distros.
#
# Behavior:
#   - Detects distro via /etc/os-release (preferred) and dispatches to a distro script.
#   - Does NOT perform a full system upgrade/dist-upgrade.
#   - Prints clear guidance for unsupported distros.

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$(cd -- "${SCRIPT_DIR}/../.." >/dev/null 2>&1 && pwd)"

ARCH_SCRIPT="${SCRIPT_DIR}/arch.sh"
DEBIAN_SCRIPT="${SCRIPT_DIR}/debian.sh"

say() { printf '%s\n' "$*"; }
die() { printf 'ERROR: %s\n' "$*" >&2; exit 1; }

# Basic sanity checks
[[ -x "${ARCH_SCRIPT}" ]] || die "Missing or non-executable: ${ARCH_SCRIPT}"
[[ -x "${DEBIAN_SCRIPT}" ]] || die "Missing or non-executable: ${DEBIAN_SCRIPT}"

OS_RELEASE="/etc/os-release"
[[ -r "${OS_RELEASE}" ]] || die "Cannot read ${OS_RELEASE}; this installer supports Linux systems with os-release."

# shellcheck disable=SC1090
source "${OS_RELEASE}"

ID_LIKE="${ID_LIKE:-}"
OS_ID="${ID:-unknown}"

# Architecture is generally irrelevant for package selection here, but we log it for clarity.
ARCH="$(uname -m || true)"

say "talkkonnect deps installer"
say "  Detected OS:   ${OS_ID}${ID_LIKE:+ (like: ${ID_LIKE})}"
say "  Detected arch: ${ARCH}"
say

is_like() {
  local needle="$1"
  [[ " ${ID_LIKE} " == *" ${needle} "* ]]
}

case "${OS_ID}" in
  arch|manjaro|endeavouros|garuda)
    exec "${ARCH_SCRIPT}"
    ;;
  debian|ubuntu|raspbian|linuxmint|pop)
    exec "${DEBIAN_SCRIPT}"
    ;;
  *)
    # Fall back to ID_LIKE for derivatives.
    if is_like arch; then
      exec "${ARCH_SCRIPT}"
    elif is_like debian || is_like ubuntu; then
      exec "${DEBIAN_SCRIPT}"
    else
      die "Unsupported distro '${OS_ID}'. Supported: Arch-family and Debian-family.
Try:
  - Arch:   make deps-arch
  - Debian: make deps-debian
Or install dependencies manually (see scripts/deps/* and README)."
    fi
    ;;
esac
