#!/bin/bash

export CGO_LDFLAGS="-lX11"
go test -timeout 30s -v ./...
if [ $? -eq 0 ]; then
	echo "Tests passed, compiling code..."
	go build -tags editor -o ./bin/kaiju -ldflags="-s -w" ./src/main.go
else
	echo "Tests failed, skipping code compile"
fi