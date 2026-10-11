#!/bin/sh
set -e
# Maps a Docker platform (TARGETARCH, TARGETVARIANT) to its Xray archive and the GOARCH the
# panel names binaries by; Xray alone calls arm "arm32" (internal/xray/process.go).
case "$1/${2:-}" in
    amd64/*) XRAY_ZIP="64" GO_ARCH="amd64" ;;
    arm64/*) XRAY_ZIP="arm64-v8a" GO_ARCH="arm64" ;;
    arm/v7 | arm/) XRAY_ZIP="arm32-v7a" GO_ARCH="arm" GO_ARM="7" ;;
    arm/v6) XRAY_ZIP="arm32-v6" GO_ARCH="arm" GO_ARM="6" ;;
    386/*) XRAY_ZIP="32" GO_ARCH="386" ;;
    *)
        echo "DockerInit: unsupported platform $1/${2:-}" >&2
        exit 1
        ;;
esac
XRAY_NAME="$GO_ARCH"
if [ "$GO_ARCH" = "arm" ]; then
    XRAY_NAME="arm32"
fi
mkdir -p build/bin
cd build/bin
curl -sfLRO "https://github.com/GFW-knocker/Xray-core/releases/download/v1.26.10-mahsa-r2/Xray-linux-${XRAY_ZIP}.zip"
unzip "Xray-linux-${XRAY_ZIP}.zip"
rm -f "Xray-linux-${XRAY_ZIP}.zip" geoip.dat geosite.dat
mv xray "xray-linux-${XRAY_NAME}"
# mtg-multi is built from the commit pinned in tools/build-mtg-multi.sh; an explicit GOARM
# keeps the arm/v6 image off Go's ARMv7 default, which that CPU cannot run.
GOOS=linux GOARCH="$GO_ARCH" GOARM="${GO_ARM:-}" sh ../../tools/build-mtg-multi.sh "mtg-linux-${GO_ARCH}"
curl -sfLRO https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geoip.dat
curl -sfLRO https://github.com/Loyalsoldier/v2ray-rules-dat/releases/latest/download/geosite.dat
curl -sfLRo geoip_IR.dat https://github.com/chocolate4u/Iran-v2ray-rules/releases/latest/download/geoip.dat
curl -sfLRo geosite_IR.dat https://github.com/chocolate4u/Iran-v2ray-rules/releases/latest/download/geosite.dat
curl -sfLRo geoip_RU.dat https://github.com/runetfreedom/russia-v2ray-rules-dat/releases/latest/download/geoip.dat
curl -sfLRo geosite_RU.dat https://github.com/runetfreedom/russia-v2ray-rules-dat/releases/latest/download/geosite.dat
cd ../../
