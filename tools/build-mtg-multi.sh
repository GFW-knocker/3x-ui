#!/bin/sh
# Builds the mtg-multi MTProto sidecar from source at a pinned commit.
# Usage: build-mtg-multi.sh <output-binary>; cross-compiles via GOOS/GOARCH/GOARM.
set -eu

MTG_MULTI_REPO="https://github.com/GFW-knocker/mtg-multi.git"
MTG_MULTI_COMMIT="9ba241ba8adf7ea4d8d12055407a7ac5b81b5633"

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <output-binary>" >&2
    exit 2
fi
case $1 in
    /*) out=$1 ;;
    *) out="$(pwd)/$1" ;;
esac

src=$(mktemp -d)
trap 'rm -rf "$src"' EXIT

git -C "$src" init -q
git -C "$src" fetch -q --depth 1 "$MTG_MULTI_REPO" "$MTG_MULTI_COMMIT"
git -C "$src" checkout -q FETCH_HEAD
# A fetch by hash must land on exactly that commit; anything else is not the reviewed source.
head=$(git -C "$src" rev-parse HEAD)
if [ "$head" != "$MTG_MULTI_COMMIT" ]; then
    echo "mtg-multi: fetched $head, expected $MTG_MULTI_COMMIT" >&2
    exit 1
fi

short=$(echo "$MTG_MULTI_COMMIT" | cut -c1-8)
(cd "$src" && CGO_ENABLED=0 go build -trimpath -mod=readonly \
    -ldflags "-s -w -X main.version=$short" -o "$out" .)
echo "mtg-multi $short built: $out"
