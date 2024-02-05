package host_container

import (
	"kaiju/engine"
	"time"
)

type HostContainer struct {
	Host         *engine.Host
	runFunctions []func()
	CloseSignal  chan bool
}

func (c *HostContainer) RunFunction(f func()) {
	c.runFunctions = append(c.runFunctions, f)
}

func (c *HostContainer) Run() {
	lastTime := time.Now()
	for !c.Host.Closing {
		since := time.Since(lastTime)
		deltaTime := since.Seconds()
		lastTime = time.Now()
		c.Host.Update(deltaTime)
		if !c.Host.Closing {
			c.Host.Render()
		}
	}
	c.Host.Teardown()
	c.CloseSignal <- true
}

func New() (*HostContainer, error) {
	host, err := engine.NewHost()
	if err != nil {
		return nil, err
	}
	c := &HostContainer{&host, []func(){}, make(chan bool)}
	host.Window.Renderer.Initialize(&host, int32(host.Window.Width()), int32(host.Window.Height()))
	host.FontCache().Init(host.Window.Renderer, host.AssetDatabase(), &host)
	c.Host.Updater.AddUpdate(func(deltaTime float64) {
		if len(c.runFunctions) > 0 {
			for _, f := range c.runFunctions {
				f()
			}
			c.runFunctions = c.runFunctions[:0]
		}
	})
	return c, nil
}
