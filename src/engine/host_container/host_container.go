/******************************************************************************/
/* host_container.go                                                          */
/******************************************************************************/
/*                           This file is part of:                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.org                           */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright, blessing, biblical verse, notice and                  */
/* this permission notice shall be included in all copies or                  */
/* substantial portions of the Software.                                      */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package host_container

import (
	"kaiju/engine"
	"kaiju/engine/systems/console"
	"kaiju/engine/systems/logging"
	"kaiju/platform/profiler/tracing"
	"log/slog"
	"runtime"
	"strconv"
	"strings"
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

func (c *Container) Run(width, height, x, y int) error {
	runtime.LockOSThread()
	if err := c.Host.Initialize(width, height, x, y); err != nil {
		slog.Error("Failed to initialize the host", "error", err)
		return err
	}
	if err := c.Host.InitializeRenderer(); err != nil {
		slog.Error("Failed to initialize the renderer", "error", err)
		return err
	}
	if err := c.Host.InitializeAudio(); err != nil {
		slog.Error("Failed to initialize audio", "error", err)
		return err
	}
	lastTime := time.Now()
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
		since := time.Since(lastTime)
		deltaTime := since.Seconds()
		lastTime = time.Now()
		c.Host.Update(deltaTime)
		if !c.Host.Closing {
			c.Host.Render()
		}
		r.End()
	}
	console.UnlinkHost(c.Host)
	c.Host.Teardown()
	runtime.UnlockOSThread()
	return nil
}

func New(name string, logStream *logging.LogStream) *Container {
	defer tracing.NewRegion("host_container.New").End()
	host := engine.NewHost(name, logStream)
	c := &Container{
		Host:         host,
		runFunctions: []func(){},
		PrepLock:     make(chan struct{}),
	}
	c.Host.Updater.AddUpdate(func(deltaTime float64) {
		defer tracing.NewRegion("engine.Host.runFunctions").End()
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
	defer tracing.NewRegion("Container.Close").End()
	c.Host.Close()
}
