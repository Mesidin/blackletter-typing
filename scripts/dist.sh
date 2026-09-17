#!/bin/sh
# Build portable binaries for Mac (Intel + Apple Silicon) and Linux (Omarchy).
set -e
cd "$(dirname "$0")/.."
mkdir -p dist
export CGO_ENABLED=0

build() {
	os=$1
	arch=$2
	out=$3
	echo "→ dist/$out"
	GOOS=$os GOARCH=$arch go build -trimpath -ldflags="-s -w" -o "dist/$out" .
}

build darwin amd64 blackletter-mac-intel
build darwin arm64 blackletter-mac-apple
build linux amd64 blackletter-linux

echo
echo "Copy one file onto the other machine and make it executable:"
echo "  Mac (Apple Silicon):  dist/blackletter-mac-apple"
echo "  Mac (Intel):          dist/blackletter-mac-intel"
echo "  Omarchy / Linux:      dist/blackletter-linux"
echo
ls -lh dist
