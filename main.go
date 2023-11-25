package main

import (
	"kaiju/bootstrap"
	"kaiju/editor/project"
	"kaiju/engine"
	"time"
)

func main() {
	lastTime := time.Now()
	host := engine.NewHost()
	bootstrap.Main(&host)
	for !host.Closing {
		deltaTime := time.Since(lastTime).Seconds()
		lastTime = time.Now()
		host.Update(deltaTime)
	}
}
