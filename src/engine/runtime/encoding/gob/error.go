/******************************************************************************/
/* error.go                                                                   */
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

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gob

import "fmt"

// Errors in decoding and encoding are handled using panic and recover.
// Panics caused by user error (that is, everything except run-time panics
// such as "index out of bounds" errors) do not leave the file that caused
// them, but are instead turned into plain error returns. Encoding and
// decoding functions and methods that do not return an error either use
// panic to report an error or are guaranteed error-free.

// A gobError is used to distinguish errors (panics) generated in this package.
type gobError struct {
	err error
}

// errorf is like error_ but takes Printf-style arguments to construct an error.
// It always prefixes the message with "gob: ".
func errorf(format string, args ...any) {
	error_(fmt.Errorf("gob: "+format, args...))
}

// error_ wraps the argument error and uses it as the argument to panic.
func error_(err error) {
	panic(gobError{err})
}

// catchError is meant to be used as a deferred function to turn a panic(gobError) into a
// plain error. It overwrites the error return of the function that deferred its call.
func catchError(err *error) {
	if e := recover(); e != nil {
		ge, ok := e.(gobError)
		if !ok {
			panic(e)
		}
		*err = ge.err
	}
}
