package rendering

import (
	"encoding/json"
	"kaiju/engine/assets"
)

func unmarshallJsonFile(assets *assets.Database, file string, to any) error {
	s, err := assets.ReadText(file)
	if err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(s), to); err != nil {
		return err
	}
	return nil
}
