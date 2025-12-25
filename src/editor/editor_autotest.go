/******************************************************************************/
/* editor_autotest.go                                                         */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/* Copyright (c) 2025-present Raphaël Côté.                                   */
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
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package editor

import (
	"fmt"
	"kaiju/build"
	"kaiju/engine"
	"log/slog"
	"os"
)

type autoTestStep struct {
	label string
	call  func()
	wait  int
}

// autoTestState tracks the state of the automated integration test
type autoTestState struct {
	frameCount    int
	testStep      int
	nextStepFrame int
	entityCreated bool
	steps         []autoTestStep
}

var autoTest autoTestState

// initAutoTest checks if auto-test mode is enabled and sets it up
func (ed *Editor) initAutoTest() bool {
	if build.Debug && engine.LaunchParams.AutoTest {
		autoTest.steps = []autoTestStep{
			{label: "Starting tests", call: func() {}, wait: 60},
			{label: "Creating new entity", call: ed.CreateNewEntity, wait: 5},
			{label: "Testing undo operation", call: ed.history.Undo, wait: 5},
			{label: "Testing redo operation", call: ed.history.Redo, wait: 5},
			{label: "Switching to content workspace", call: ed.ContentWorkspaceSelected, wait: 5},
			{label: "Switching back to stage workspace", call: ed.StageWorkspaceSelected, wait: 5},
		}
		slog.Info("AutoTest mode enabled - will run automated integration tests")
		return true
	}
	return false
}

// runAutoTest executes automated integration tests when --autotest flag is set
func (ed *Editor) runAutoTest(deltaTime float64) {
	autoTest.frameCount++
	if autoTest.frameCount < autoTest.nextStepFrame {
		return
	}
	t := autoTest.steps[autoTest.testStep]
	slog.Info(fmt.Sprintf("AutoTest: Step %d - %s", autoTest.testStep, t.label))
	t.call()
	autoTest.nextStepFrame = t.wait
	slog.Info(fmt.Sprintf("AutoTest: Step %d - Complete", autoTest.testStep))
	autoTest.testStep++
	if autoTest.testStep >= len(autoTest.steps) {
		slog.Info("AutoTest: All tests completed successfully!")
		slog.Info("AutoTest: Exiting with success code")
		os.Exit(0)
	}
}
