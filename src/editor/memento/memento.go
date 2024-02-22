package memento

type Memento interface {
	Redo()
	Undo()
	Delete()
	Exit()
}
