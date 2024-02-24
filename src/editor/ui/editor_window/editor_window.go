package editor_window

import "kaiju/host_container"

type EditorWindow interface {
	Tag() string
	Container() *host_container.Container
}
