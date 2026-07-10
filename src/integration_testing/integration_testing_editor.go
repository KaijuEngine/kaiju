//go:build editor

/******************************************************************************/
/* integration_testing_editor.go                                              */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package integration_testing

import (
	"kaijuengine.com/editor/editor_embedded_content"
	"kaijuengine.com/engine/assets"
)

func (IntegrationGame) ContentDatabase() (assets.Database, error) {
	// TODO:  Only do this if it is the editor, otherwise use standard content
	return &editor_embedded_content.EditorContent{}, nil
}
