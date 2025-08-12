/******************************************************************************/
/* assertions.go                                                              */
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

package debug

import (
	"errors"
	"kaiju/build"
	"log/slog"
	"runtime"
)

var NotImplementedError = errors.New("not implemented")

func Log(msg string, args ...any) {
	if build.Debug {
		slog.Debug(msg, args...)
	}
}

func Assert(res bool, msg string) {
	if build.Debug {
		if !res {
			panic(msg)
		}
	} else {
		slog.Error(msg)
	}
}

func Halt(msg string) {
	if !build.Shipping {
		slog.Error(msg)
		runtime.Breakpoint()
	}
}

func Ensure(res bool) {
	if !build.Shipping {
		if !res {
			runtime.Breakpoint()
		}
	}
}

func EnsureMsg(res bool, msg string) {
	if !build.Shipping {
		if !res {
			slog.Error(msg)
			runtime.Breakpoint()
		}
	}
}

func EnsureNotError(err error) {
	if !build.Shipping {
		if err != nil {
			EnsureMsg(false, err.Error())
		}
	}
}

func EnsureNotNil(target any) {
	if !build.Shipping {
		EnsureMsg(target != nil, "the target was expected not to be nil, but was")
	}
}

func ThrowNotImplemented(todo string) { EnsureMsg(false, todo) }
