package host_container

import (
	"kaiju/engine"
	"kaiju/systems/console"
	"runtime"
	"time"
)

type HostContainer struct {
	Host         *engine.Host
	runFunctions []func()
	PrepLock     chan bool
}

func (c *HostContainer) RunFunction(f func()) {
	c.runFunctions = append(c.runFunctions, f)
}

func (c *HostContainer) Run() error {
	runtime.LockOSThread()
	if err := c.Host.Initialize(); err != nil {
		return err
	}
	c.Host.Window.Renderer.Initialize(c.Host, int32(c.Host.Window.Width()), int32(c.Host.Window.Height()))
	c.Host.FontCache().Init(c.Host.Window.Renderer, c.Host.AssetDatabase(), c.Host)
	lastTime := time.Now()
	c.PrepLock <- true
	for !c.Host.Closing {
		c.Host.WaitForFrameRate()
		since := time.Since(lastTime)
		deltaTime := since.Seconds()
		lastTime = time.Now()
		c.Host.Update(deltaTime)
		if !c.Host.Closing {
			c.Host.Render()
		}
	}
	console.UnlinkHost(c.Host)
	c.Host.Teardown()
	runtime.UnlockOSThread()
	return nil
}

func New(name string) *HostContainer {
	host := engine.NewHost(name)
	c := &HostContainer{
		Host:         host,
		runFunctions: []func(){},
		PrepLock:     make(chan bool),
	}
	c.Host.Updater.AddUpdate(func(deltaTime float64) {
		if len(c.runFunctions) > 0 {
			for _, f := range c.runFunctions {
				f()
			}
			c.runFunctions = c.runFunctions[:0]
		}
	})
	return c
}
