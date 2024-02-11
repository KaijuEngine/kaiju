package host_container

import (
	"kaiju/engine"
	"kaiju/systems/console"
	"runtime"
	"time"
)

type Container struct {
	Host         *engine.Host
	runFunctions []func()
	PrepLock     chan struct{}
}

func (c *Container) RunFunction(f func()) {
	c.runFunctions = append(c.runFunctions, f)
}

func (c *Container) Run(width, height int) error {
	runtime.LockOSThread()
	if err := c.Host.Initialize(width, height); err != nil {
		return err
	}
	c.Host.Window.Renderer.Initialize(c.Host, int32(c.Host.Window.Width()), int32(c.Host.Window.Height()))
	c.Host.FontCache().Init(c.Host.Window.Renderer, c.Host.AssetDatabase(), c.Host)
	lastTime := time.Now()
	c.PrepLock <- struct{}{}
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

func New(name string) *Container {
	host := engine.NewHost(name)
	c := &Container{
		Host:         host,
		runFunctions: []func(){},
		PrepLock:     make(chan struct{}),
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

func (c *Container) Close() {
	c.Host.Close()
}
