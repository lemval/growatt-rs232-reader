#!/bin/bash

go clean
go build

env GOOS=linux GOARCH=arm   go build -o growatt_pi
env GOOS=linux GOARCH=arm64 go build -o growatt_odroid
scp growatt_pi pi@10.0.2.72:/tmp
