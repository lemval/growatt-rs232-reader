#!/bin/bash

# You might need a 'go get ...' to load the modules

go clean
go build

env GOOS=linux GOARCH=arm   go build -o growatt_pi -ldflags '-s'
env GOOS=linux GOARCH=arm64 go build -o growatt_odroid -ldflags '-s'
