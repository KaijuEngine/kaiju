package editor_workspace

type Workspace interface {
	Open()
	Update(deltaTime float64)
	Close()
}
