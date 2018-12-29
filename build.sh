#!/bin/bash

# You might need a 'go get ...' to load the modules

go clean
go build

mkdir release >/dev/null 2>&1

env GOOS=linux GOARCH=arm GOARM=5 go build -o release/growatt_pi2 -ldflags '-s'
env GOOS=linux GOARCH=arm         go build -o release/growatt_pi -ldflags '-s'
env GOOS=linux GOARCH=arm64       go build -o release/growatt_odroid -ldflags '-s'
