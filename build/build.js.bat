SET "GOOS=js"
SET "GOARCH=wasm"
SET "CGO_ENABLED=0"
go build -tags editor -o ./bin/kaiju.wasm -ldflags="-s -w" main.go