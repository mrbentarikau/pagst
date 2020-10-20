#!/bin/bash
VERSION=$(git describe --tags)
echo Building version $VERSION
go build -ldflags "-X github.com/mrbentarikau/pagst/common.VERSION=${VERSION}"