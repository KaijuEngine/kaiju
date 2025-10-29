package editor_settings

import "fmt"

type AppDataMissingError struct {
	Err error
}

func (e AppDataMissingError) Error() string {
	return fmt.Sprintf("failed to get the editor application data folder: %v", e.Err)
}

type WriteError struct {
	Err      error
	onEncode bool
}

func (e WriteError) Error() string {
	if e.onEncode {
		return fmt.Sprintf("failed to encode the settings file: %v", e.Err)
	} else {
		return fmt.Sprintf("failed to create the settings file: %v", e.Err)
	}
}

type ReadError struct {
	Err      error
	onDecode bool
}

func (e ReadError) Error() string {
	if e.onDecode {
		return fmt.Sprintf("failed to decode the settings file: %v", e.Err)
	} else {
		return fmt.Sprintf("failed to load the settings file: %v", e.Err)
	}
}
