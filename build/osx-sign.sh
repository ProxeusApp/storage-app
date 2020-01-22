#!/bin/bash
set -Eeuxo pipefail

WORKDIR="$(dirname $(pwd)/$0)"
APP="$1"

[ -n "$APP" ] || ( echo "Please specify .app file to sign" && exit 1 )

rm -rf /tmp/Proxeus.app
cp -R "$WORKDIR/Proxeus.app" /tmp/
cp -R "$APP" /tmp/Proxeus.app/Contents/MacOS/Proxeus.app
cd /tmp

codesign -s XJY74CAGS5 --deep -f -vv /tmp/Proxeus.app/Contents/MacOS/Proxeus.app
spctl --assess -v /tmp/Proxeus.app/Contents/MacOS/Proxeus.app

codesign -s XJY74CAGS5 --deep -f -vv /tmp/Proxeus.app
spctl --assess -v /tmp/Proxeus.app

unzip "$WORKDIR/empty.dmg.zip"
hdiutil eject /Volumes/Proxeus/ || true
hdiutil attach empty.dmg
cd -

mv /tmp/Proxeus.app /Volumes/Proxeus/
hdiutil eject /Volumes/Proxeus/
rm -f Proxeus.dmg
hdiutil convert -format UDZO -o Proxeus.dmg /tmp/empty.dmg
rm -f /tmp/empty.dmg
