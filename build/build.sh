#!/bin/bash

export CGO_LDFLAGS="-lX11 -lGL"
go test -timeout 30s -v ./...
if [ $? -eq 0 ]; then
	echo "Tests passed, compiling code..."
	go build -tags editor,OPENGL -o ./bin/kaiju -ldflags="-s -w" main.go
else
	echo "Tests failed, skipping code compile"
fi