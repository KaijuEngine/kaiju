/******************************************************************************/
/* integration_testing_non_editor.go                                          */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

//go:build !editor

package integration_testing

import (
	"errors"

	"kaijuengine.com/engine/assets"
)

func (IntegrationGame) ContentDatabase() (assets.Database, error) {
	return nil, errors.New("not implemented")
}
