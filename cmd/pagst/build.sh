#!/bin/sh
#VERSION=$(git describe --tags)
VERSION=$(git describe --tags | awk -F "-" '{print $1}')

currentDate=$(date +%s)
targetFile="../../stdcommands/exchange/currency_codes.go"

if [ -f $targetFile ]; then
	fileDate=$(date -r $targetFile +%s)
else
	echo currency_codes.go does not exist
	fileDate=86400
fi

dateDifference=$(($currentDate - $fileDate))

if [ $dateDifference -ge 86400 ]; then
	echo Genereting Currency Exchange money symbols
	go generate ../../stdcommands/exchange	
else
	echo Skipping Currency Exchange sync
fi

echo Building version $VERSION
go build -ldflags "-X github.com/mrbentarikau/pagst/common.VERSION=${VERSION}"
