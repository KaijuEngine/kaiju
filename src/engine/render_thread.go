/******************************************************************************/
/* render_thread.go                                                           */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package engine

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"

	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/rendering"
)

const renderThreadFrameQueueDepth = 1

type RenderThread struct {
	host     *Host
	requests chan renderThreadRequest
	initDone chan error
	done     chan error
	stopOnce sync.Once
	stopped  chan struct{}
	threadID atomic.Uint64
}

type renderThreadRequest struct {
	frame *RenderFrame
	call  func(*rendering.GPUDevice) error
	done  chan error
	stop  bool
}

func NewRenderThread(host *Host) *RenderThread {
	return &RenderThread{
		host:     host,
		requests: make(chan renderThreadRequest, renderThreadFrameQueueDepth),
		initDone: make(chan error, 1),
		done:     make(chan error, 1),
		stopped:  make(chan struct{}),
	}
}

func (host *Host) StartRenderThread() error {
	if host.renderThread != nil {
		return nil
	}
	thread := NewRenderThread(host)
	host.renderThread = thread
	if err := thread.Start(); err != nil {
		host.renderThread = nil
		return err
	}
	return nil
}

func (host *Host) RenderThreadActive() bool {
	return host.renderThread != nil
}

func (host *Host) RenderThreadProcessPending() error {
	if host.renderThread == nil {
		host.ProcessPendingRenderResources()
		return nil
	}
	return host.renderThread.ProcessPending()
}

func (host *Host) SubmitRenderFrame(frame RenderFrame) error {
	if host.renderThread == nil {
		host.renderCapturedFrame(frame)
		return nil
	}
	return host.renderThread.Submit(frame)
}

func (host *Host) SubmitRenderFrameAndWait(frame RenderFrame) error {
	if host.renderThread == nil {
		host.renderCapturedFrame(frame)
		return nil
	}
	return host.renderThread.SubmitAndWait(frame)
}

func (r *RenderThread) Start() error {
	go r.loop()
	return <-r.initDone
}

func (r *RenderThread) ProcessPending() error {
	return r.runSync(func(*rendering.GPUDevice) error {
		r.host.ProcessPendingRenderResources()
		return nil
	})
}

func (r *RenderThread) Submit(frame RenderFrame) error {
	return r.enqueue(renderThreadRequest{frame: &frame})
}

func (r *RenderThread) SubmitAndWait(frame RenderFrame) error {
	done := make(chan error, 1)
	if err := r.enqueue(renderThreadRequest{frame: &frame, done: done}); err != nil {
		return err
	}
	return <-done
}

func (r *RenderThread) Run(call func(*rendering.GPUDevice)) {
	if call == nil {
		return
	}
	if r.IsRenderThread() {
		call(r.host.Window.GpuInstance.PrimaryDevice())
		return
	}
	_ = r.runSync(func(device *rendering.GPUDevice) error {
		call(device)
		return nil
	})
}

func (r *RenderThread) IsRenderThread() bool {
	id := currentOSThreadID()
	return id != 0 && id == r.threadID.Load()
}

func (r *RenderThread) Stop() error {
	if r == nil {
		return nil
	}
	var err error
	r.stopOnce.Do(func() {
		done := make(chan error, 1)
		err = r.enqueue(renderThreadRequest{stop: true, done: done})
		if err == nil {
			err = <-done
		}
		<-r.stopped
	})
	return err
}

func (r *RenderThread) runSync(call func(*rendering.GPUDevice) error) error {
	done := make(chan error, 1)
	if err := r.enqueue(renderThreadRequest{call: call, done: done}); err != nil {
		return err
	}
	return <-done
}

func (r *RenderThread) enqueue(request renderThreadRequest) error {
	if r == nil {
		return errors.New("render thread is nil")
	}
	select {
	case <-r.stopped:
		return errors.New("render thread is stopped")
	default:
	}
	if cap(r.requests) > 0 && len(r.requests) == cap(r.requests) {
		region := tracing.NewRegion("RenderThread.QueueBackpressure")
		r.requests <- request
		region.End()
		return nil
	}
	r.requests <- request
	return nil
}

func (r *RenderThread) loop() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer close(r.stopped)
	r.threadID.Store(currentOSThreadID())
	if err := r.host.InitializeRenderer(); err != nil {
		r.initDone <- err
		r.done <- err
		return
	}
	r.initDone <- nil
	for {
		wait := tracing.NewRegion("RenderThread.WaitForFrame")
		request := <-r.requests
		wait.End()
		err := r.handleRequest(request)
		if request.done != nil {
			request.done <- err
		}
		if request.stop {
			r.done <- err
			return
		}
	}
}

func (r *RenderThread) handleRequest(request renderThreadRequest) error {
	if request.stop {
		r.host.TeardownRenderer()
		return nil
	}
	if request.call != nil {
		device := r.host.Window.GpuInstance.PrimaryDevice()
		return request.call(device)
	}
	if request.frame != nil {
		r.host.runBeforeRenderCallbacks()
		r.host.ProcessPendingRenderResources()
		r.host.renderCapturedFrame(*request.frame)
	}
	return nil
}
