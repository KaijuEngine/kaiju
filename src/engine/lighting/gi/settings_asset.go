package gi

import "encoding/json"

const SettingsAssetKey = "globalIlluminationSettings"

func MarshalSettings(settings Settings) ([]byte, error) {
	if err := settings.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(settings)
}

func ReadSettings(reader AssetReader) (Settings, error) {
	data, err := reader.Read(SettingsAssetKey)
	if err != nil {
		return Settings{}, err
	}
	settings := Settings{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return Settings{}, err
	}
	return settings, settings.Validate()
}
