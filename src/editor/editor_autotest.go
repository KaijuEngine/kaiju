/******************************************************************************/
/* ui_workspace.go                                                            */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/* + Copyright (c) 2025-present Raphaël Côté                                  */
/******************************************************************************/

package editor

import (
	"fmt"
	"log/slog"
	"os"

	"kaijuengine.com/build"
	"kaijuengine.com/engine"
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
			{label: "Switching to content workspace", call: func() { ed.WorkspaceSelected("content") }, wait: 5},
			{label: "Switching back to stage workspace", call: func() { ed.WorkspaceSelected("stage") }, wait: 5},
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
