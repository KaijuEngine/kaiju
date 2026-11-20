/******************************************************************************/
/* integration_testing_non_editor.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

//go:build !editor

package integration_testing

import (
	"errors"
	"os"

	"kaijuengine.com/engine/assets"
)

func (IntegrationGame) ContentDatabase() (assets.Database, error) {
	if _, err := os.Stat("game_content"); err == nil {
		return assets.NewFileDatabase("game_content")
	}
	if _, err := os.Stat("src/game_content"); err == nil {
		return assets.NewFileDatabase("src/game_content")
	}
	return nil, errors.New("game_content database not found")
}
