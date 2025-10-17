package editor_workspace

type Workspace interface {
	Open()
	Close()
	Focus()
	Blur()
	Update(deltaTime float64)
}
