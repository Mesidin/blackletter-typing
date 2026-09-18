#!/bin/sh
# Install Blackletter binary and desktop entry for Omarchy / Linux main app menu.
set -e
cd "$(dirname "$0")/.."

mkdir -p ~/.local/bin
mkdir -p ~/.local/share/applications

if [ -f dist/blackletter-linux ]; then
    cp dist/blackletter-linux ~/.local/bin/blackletter
else
    go build -o ~/.local/bin/blackletter .
fi

chmod +x ~/.local/bin/blackletter
cp assets/blackletter.desktop ~/.local/share/applications/

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database ~/.local/share/applications/
fi

echo "Installed Blackletter to ~/.local/bin/blackletter"
echo "Added desktop shortcut to ~/.local/share/applications/blackletter.desktop"
echo "Blackletter should now appear in your Omarchy / Linux main apps menu!"
