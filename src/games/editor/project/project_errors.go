package project

import "fmt"

type ConfigLoadError struct {
	Err error
}

func (e ConfigLoadError) Error() string {
	return fmt.Sprintf("failed to load the project configuration file: %v", e.Err)
}
