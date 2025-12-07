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

package editor

import (
	"kaiju/engine"
	"log/slog"
	"os"
)

// autoTestState tracks the state of the automated integration test
type autoTestState struct {
	frameCount    int
	testStep      int
	entityCreated bool
}

var autoTest autoTestState

// runAutoTest executes automated integration tests when --autotest flag is set
func (ed *Editor) runAutoTest() {
	// Wait for initialization (project loaded, workspaces ready)
	autoTest.frameCount++

	// Wait 60 frames (~1 second) for everything to initialize
	if autoTest.frameCount < 60 {
		return
	}

	// Execute test sequence
	switch autoTest.testStep {
	case 0:
		slog.Info("AutoTest: Starting automated integration test")
		slog.Info("AutoTest: Step 1 - Creating new entity")
		ed.CreateNewEntity()
		autoTest.entityCreated = true
		autoTest.testStep++

	case 1:
		// Wait a few frames for entity creation to complete
		if autoTest.frameCount > 65 {
			slog.Info("AutoTest: Step 2 - Testing undo operation")
			ed.history.Undo()
			autoTest.testStep++
		}

	case 2:
		// Wait a few frames after undo
		if autoTest.frameCount > 70 {
			slog.Info("AutoTest: Step 3 - Testing redo operation")
			ed.history.Redo()
			autoTest.testStep++
		}

	case 3:
		// Wait a few frames after redo
		if autoTest.frameCount > 75 {
			slog.Info("AutoTest: Step 4 - Switching to content workspace")
			ed.ContentWorkspaceSelected()
			autoTest.testStep++
		}

	case 4:
		// Wait a few frames after workspace switch
		if autoTest.frameCount > 80 {
			slog.Info("AutoTest: Step 5 - Switching to shading workspace")
			ed.ShadingWorkspaceSelected()
			autoTest.testStep++
		}

	case 5:
		// Wait a few frames after workspace switch
		if autoTest.frameCount > 85 {
			slog.Info("AutoTest: Step 6 - Switching back to stage workspace")
			ed.StageWorkspaceSelected()
			autoTest.testStep++
		}

	case 6:
		// Wait a few frames to ensure stability
		if autoTest.frameCount > 90 {
			slog.Info("AutoTest: All tests completed successfully!")
			slog.Info("AutoTest: Exiting with success code")
			os.Exit(0)
		}
	}
}

// initAutoTest checks if auto-test mode is enabled and sets it up
func (ed *Editor) initAutoTest() bool {
	if engine.LaunchParams.AutoTest {
		slog.Info("AutoTest mode enabled - will run automated integration tests")
		return true
	}
	return false
}
