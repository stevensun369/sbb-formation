#!/usr/bin/env sh
set -eu

if [ ! -f .env ]; then
  echo "missing .env; copy .env.example to .env and set TOKEN" >&2
  exit 1
fi

# Load only the build inputs needed by this script. TOKEN is embedded into
# each binary by the linker and is not required when the binary runs.
set -a
. ./.env
set +a

: "${TOKEN:?TOKEN must be set in .env}"

mkdir -p bin
build() {
  GOOS="$1" GOARCH="$2" suffix="$3"
  output="bin/sbb_formation-${suffix}"
  [ "$GOOS" = "windows" ] && output="${output}.exe"
  CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
    -trimpath \
    -ldflags "-s -w -X main.buildToken=${TOKEN}" \
    -o "$output" .
}

build linux amd64 linux-amd64
build windows amd64 windows-amd64
build darwin amd64 macos-amd64
build darwin arm64 macos-arm64

echo "Built binaries in bin/"
