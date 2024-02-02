SET "GOOS=js"
SET "GOARCH=wasm"
SET "CGO_ENABLED=0"
go build -tags editor,OPENGL -o ./bin/kaiju.wasm -ldflags="-s -w" ./src/main.go