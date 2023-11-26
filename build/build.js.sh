#!/bin/bash

export GOOS="js"
export GOARCH="wasm"
export CGO_ENABLED=0
go build -tags editor -o ./bin/kaiju.wasm -ldflags="-s -w" main.go