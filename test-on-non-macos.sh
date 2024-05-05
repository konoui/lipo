#!/bin/sh
# https://github.com/actions/runner-images/issues/7253

set -eu

FAT_BIN=./bin/fat
ARM64_BIN=./bin/arm64
AMD64_BIN=./bin/amd64
EXTRACT_BIN=./bin/extract
THIB_BIN=./bin/thin
LIPO_BIN=./bin/lipo

echo "====building lipo binary"
make
echo "====building base binaries"
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/amd64 example/main.go
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o bin/arm64 example/main.go

echo "====creating fat"
$LIPO_BIN -output $FAT_BIN -create $ARM64_BIN $AMD64_BIN
$LIPO_BIN $FAT_BIN -verify_arch x86_64 arm64

echo "====extracting arm64"
$LIPO_BIN -extract arm64 -output $EXTRACT_BIN $FAT_BIN
$LIPO_BIN $EXTRACT_BIN -verify_arch arm64

echo "====thin"
$LIPO_BIN -thin arm64 -output $THIB_BIN $FAT_BIN
$LIPO_BIN $THIB_BIN -verify_arch arm64

echo "===replace TODO"

echo all test pass!
