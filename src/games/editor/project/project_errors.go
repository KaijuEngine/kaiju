package project

type ConfigLoadError struct{}

func (e ConfigLoadError) Error() string {
	return "failed to load the project configuration file"
}
