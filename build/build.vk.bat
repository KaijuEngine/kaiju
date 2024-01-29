go test -timeout 30s -v ./...
SET "CGO_LDFLAGS=-lgdi32 -lXInput"
if %errorlevel% equ 0 (
	echo Tests passed, compiling code...
	go build -tags editor -o ./bin/kaiju.exe -ldflags="-s -w" main.go
) else (
	echo Tests failed, skipping code compile
)