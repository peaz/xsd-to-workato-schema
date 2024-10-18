#!/bin/bash

# Get the latest version from Git tag
#version=$(git describe --abbrev=0 --tags)
version="0.0.1"

directory="releases/$version"
if [ ! -d "$directory" ]; then
    mkdir -p "$directory"
    echo "Directory created: $directory"
else
    echo "Directory already exists: $directory"
fi

env GOOS=linux GOARCH=amd64 go build -o="$directory/xsd2wkt-$version-linux" $PWD/src/xsd2wkt/main.go
env GOOS=darwin GOARCH=amd64 go build -o="$directory/xsd2wkt-$version-osx" $PWD/src/xsd2wkt/main.go
env GOOS=windows GOARCH=amd64 go build -o="$directory/xsd2wkt-$version-win.exe" $PWD/src/xsd2wkt/main.go