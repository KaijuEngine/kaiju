package editor_settings

type Settings struct {
	Snapping SnapSettings
}

type SnapSettings struct {
	TranslateEnabled   bool
	RotationEnabled    bool
	ScaleEnabled       bool
	TranslateIncrement float32
	RotateIncrement    float32
	ScaleIncrement     float32
}
