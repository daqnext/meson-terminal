#!/bin/sh
# Copyright 2020 Daqnext Foundation Ltd.

VERSION="v0.1.1"
COPY_FILES=("config.txt" "*.shoppynext.com_key.key" "*.shoppynext.com_chain.crt")

generate_tar() {
  cp ${COPY_FILES[*]} "./build/$1" && cd build && zip -r "$1.zip" $1 && rm -rf $1 && cd ..|| exit
}

rm -f -R ./build
mkdir build

echo "Compiling Windows x86_64 version"

DIR="meson-${VERSION}-win32" && GOOS=windows GOARCH=386   go build -o "./build/${DIR}/meson.${VERSION}.exe" && generate_tar ${DIR}
DIR="meson-${VERSION}-win64" && GOOS=windows GOARCH=amd64 go build -o "./build/${DIR}/meson.${VERSION}.exe" && generate_tar ${DIR}

echo "Compiling MAC     x86_64 version"
DIR="meson-${VERSION}-darwin-amd64" && GOOS=darwin GOARCH=amd64 go build -o "./build/${DIR}/meson.${VERSION}" && generate_tar ${DIR}

echo "Compiling Linux   x86_64 version"
DIR="meson-${VERSION}-linux-386"   &&  GOOS=linux GOARCH=386   go build -o "./build/${DIR}/meson.${VERSION}" && generate_tar ${DIR}
DIR="meson-${VERSION}-linux-amd64" &&  GOOS=linux GOARCH=amd64 go build -o "./build/${DIR}/meson.${VERSION}" && generate_tar ${DIR}
