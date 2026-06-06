/******************************************************************************/
/* rendering_helpers.go                                                       */
/******************************************************************************/
/* MIT License, Copyright (c) 2015-present Brent Farris, (John 4:13-14)       */
/******************************************************************************/

package rendering

import (
	"encoding/json"

	"kaijuengine.com/engine/assets"
)

func unmarshallJsonFile(assets assets.Database, file string, to any) error {
	s, err := assets.ReadText(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), to); err != nil {
		return err
	}
	return nil
}
