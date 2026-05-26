/******************************************************************************/
/* host_container.go                                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package host_container

import (
	"log/slog"
	"runtime"
	"strconv"
	"strings"
	"weak"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/assets"
	"kaijuengine.com/engine/systems/logging"
	"kaijuengine.com/klib"
	"kaijuengine.com/platform/chrono"
	"kaijuengine.com/platform/profiler/tracing"
)

type Container struct {
	Host         *engine.Host
	runFunctions []func()
	PrepLock     chan struct{}
}

func (c *Container) RunFunction(f func()) {
	c.runFunctions = append(c.runFunctions, f)
}

func (c *Container) Run(width, height, x, y int, platformState any) error {
	runtime.LockOSThread()
	if err := c.Host.Initialize(width, height, x, y, platformState); err != nil {
		slog.Error("Failed to initialize the host", "error", err)
		c.Host.Close()
		return err
	}
	if err := c.Host.InitializeRenderer(); err != nil {
		slog.Error("Failed to initialize the renderer", "error", err)
		c.Host.Close()
		return err
	}
	if err := c.Host.InitializeAudio(); err != nil {
		slog.Error("Failed to initialize audio", "error", err)
		//return err
	}
	clock := chrono.HighResolutionTimer{}
	clock.Start()
	// Do one clean update and render before opening the prep lock
	c.Host.Update(0)
	c.Host.Render()
	c.PrepLock <- struct{}{}
	traceRegionName := strings.Builder{}
	for !c.Host.Closing {
		traceRegionName.Reset()
		traceRegionName.WriteString("Frame: ")
		traceRegionName.WriteString(strconv.FormatUint(c.Host.Frame(), 10))
		r := tracing.NewRegion(traceRegionName.String())
		c.Host.WaitForFrameRate()
		deltaTime := clock.Stop()
		clock.Start()
		c.Host.Update(deltaTime)
		if !c.Host.Closing {
			c.Host.Render()
		}
		r.End()
	}
	c.Host.Teardown()
	runtime.UnlockOSThread()
	return nil
}

func New(name string, logStream *logging.LogStream, adb assets.Database) *Container {
	defer tracing.NewRegion("host_container.New").End()
	host := engine.NewHost(name, logStream, adb)
	c := &Container{
		Host:         host,
		runFunctions: []func(){},
		PrepLock:     make(chan struct{}),
	}
	cw := weak.Make(c)
	c.Host.Updater.AddUpdate(func(deltaTime float64) {
		defer tracing.NewRegion("engine.Host.runFunctions").End()
		cc := cw.Value()
		if cc == nil {
			return
		}
		if len(cc.runFunctions) > 0 {
			for _, f := range cc.runFunctions {
				f()
			}
			cc.runFunctions = klib.WipeSlice(cc.runFunctions)
		}
	})
	return c
}

func (c *Container) Close() {
	defer tracing.NewRegion("Container.Close").End()
	c.Host.Close()
}
