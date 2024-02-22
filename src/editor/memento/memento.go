package memento

type Memento interface {
	Do()
	Undo()
	Delete()
}
