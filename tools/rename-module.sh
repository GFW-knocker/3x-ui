#!/bin/sh
# Rewrites upstream's Go module path to this fork's. Re-run after every upstream merge:
# files that arrive from upstream still import the old path and will not build until then.
set -eu

OLD=github.com/mhsanaei/3x-ui
NEW=github.com/GFW-knocker/3x-ui

cd "$(git rev-parse --show-toplevel)"

# The /v3 path covers imports, go.mod, -X ldflags and docs; the bare prefix is goimports'
# local-prefixes in .golangci.yml. README credits upstream by URL and is left alone.
files=$(git grep -l -F "$OLD/v3" -- ':!tools/rename-module.sh' || true)
for f in $files; do
    sed -i "s#$OLD/v3#$NEW/v3#g" "$f"
done
sed -i "s#^\(\s*- \)$OLD\$#\1$NEW#" .golangci.yml

# The new path sorts differently inside import blocks, so let gofmt re-sort them.
go_files=$(printf '%s\n' $files | grep '\.go$' || true)
if [ -n "$go_files" ]; then
    gofmt -w $go_files
fi

left=$(git grep -l -F "$OLD/v3" -- ':!tools/rename-module.sh' || true)
if [ -n "$left" ]; then
    echo "rename-module: still referencing $OLD/v3:" >&2
    echo "$left" >&2
    exit 1
fi
echo "rename-module: $(printf '%s\n' $files | grep -c .) files now use $NEW/v3"
