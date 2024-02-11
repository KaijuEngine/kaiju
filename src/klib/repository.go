/*****************************************************************************/
/* repository.go                                                             */
/*****************************************************************************/
/*                           This file is part of:                           */
/*                                KAIJU ENGINE                               */
/*                          https://kaijuengine.org                          */
/*****************************************************************************/
/* Copyright (c) 2022-present Kaiju Engine contributors (CONTRIBUTORS.md)    */
/* Copyright (c) 2015-2022 Brent Farris.                                     */
/*                                                                           */
/* May all those that this source may reach be blessed by the LORD and find  */
/* peace and joy in life.                                                    */
/* "Everyone who drinks of this water will be thirsty again; but whoever     */
/* drinks of the water that I will give him shall never thirst;" -Jesus      */
/*                                                                           */
/* Permission is hereby granted, free of charge, to any person obtaining     */
/* a copy of this software and associated documentation files (the           */
/* "Software"), to deal in the Software without restriction, including       */
/* without limitation the rights to use, copy, modify, merge, publish,       */
/* distribute, sublicense, and/or sell copies of the Software, and to        */
/* permit persons to whom the Software is furnished to do so, subject to     */
/* the following conditions:                                                 */
/*                                                                           */
/* The above copyright, blessing, notice and this permission notice shall    */
/* be included in all copies or substantial portions of the Software.        */
/*                                                                           */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,           */
/* EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF        */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.    */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY      */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,      */
/* TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE         */
/* SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                    */
/*****************************************************************************/

package klib

import (
	"fmt"
	"runtime"
)

var complaints = make(map[int]bool)

func NotYetImplemented(issueId int) {
	if _, ok := complaints[issueId]; ok {
		return
	}
	complaints[issueId] = true
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	fmt.Printf("%s:%d %s is not yet implemented. If you are interested in contributing to the project by implementing this function, please see https://github.com/KaijuEngine/kaiju/issues/%d\n", frame.File, frame.Line, frame.Function, issueId)
}
