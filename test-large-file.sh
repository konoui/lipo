#!/bin/bash

set -eu

ARM64_FILE="./bin/arm64"
AMD64_FILE="./bin/amd64"
ARM64_LARGE_FILE=$ARM64_FILE.large
AMD64_LARGE_FILE=$AMD64_FILE.large

if [ ! -f $ARM64_FILE ]; then
    echo creating ARM64_FILE
    CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o $ARM64_FILE example/main.go
fi

if [ ! -f $AMD64_FILE ]; then
    echo creating AMD64_FILE
    CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $AMD64_FILE example/main.go
fi

if [ ! -f large-data ]; then
    echo creating large data
    mkfile 2G large-data
fi 

if [ ! -f $ARM64_LARGE_FILE ]; then
    echo creating large file
    cp $ARM64_FILE $ARM64_LARGE_FILE
    cat large-data >> $ARM64_LARGE_FILE
fi

if [ ! -f $AMD64_LARGE_FILE ]; then
    echo creating large file
    cp $AMD64_FILE $AMD64_LARGE_FILE
    cat large-data >> $AMD64_LARGE_FILE
fi

echo compiling my lipo
make

echo test: creating fat file without fat64 will fail
./bin/lipo -create -output large-fat $ARM64_LARGE_FILE $AMD64_LARGE_FILE 2>&1 | grep -q "exceed"
echo pass!

echo test: creating fat file with fat64 will success
./bin/lipo -create -output large-fat $ARM64_LARGE_FILE $AMD64_LARGE_FILE -fat64
./bin/lipo -detailed_info large-fat | grep -q "0xcafebabf"
echo pass!

echo all test pass!
