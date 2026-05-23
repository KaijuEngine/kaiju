//go:build !editor

package integration_testing

import (
	"errors"

	"kaijuengine.com/engine/assets"
)

func (IntegrationGame) ContentDatabase() (assets.Database, error) {
	return nil, errors.New("not implemented")
}
