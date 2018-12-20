#!/bin/bash

export GOOS=linux; export GOARCH=arm;   go build -o i3fstatus_${GOOS}_${GOARCH} main.go
export GOOS=linux; export GOARCH=amd64; go build -o i3fstatus_${GOOS}_${GOARCH} main.go
export GOOS=linux; export GOARCH=386;   go build -o i3fstatus_${GOOS}_${GOARCH} main.go
