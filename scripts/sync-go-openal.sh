#!/bin/sh
# Sync talkkonnect's vendored go-openal copy to the standalone repository.
set -e
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/third_party/go-openal"
DST="$(cd "$ROOT/.." && pwd)/go-openal"

if [ ! -d "$SRC" ]; then
	echo "missing vendored source: $SRC" >&2
	exit 1
fi
if [ ! -d "$DST/.git" ]; then
	echo "clone the upstream repo first:" >&2
	echo "  git clone https://github.com/talkkonnect/go-openal.git $DST" >&2
	exit 1
fi

rsync -av --delete --exclude='.git' "$SRC/" "$DST/"
echo "Synced $SRC -> $DST"
echo "Review, commit, tag, and push from $DST"
