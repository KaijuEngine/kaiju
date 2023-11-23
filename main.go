package main

import (
	"kaiju/engine"
	"time"
)

func main() {
	lastTime := time.Now()
	host := engine.NewHost()
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
	}
}
