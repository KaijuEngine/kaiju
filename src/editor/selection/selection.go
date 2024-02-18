package selection

import "kaiju/engine"

type Selection struct {
	entities []*engine.Entity
}

func New() Selection {
	return Selection{
		entities: make([]*engine.Entity, 0),
	}
}

func (s *Selection) Entities() []*engine.Entity {
	return s.entities
}

func (s *Selection) IsEmpty() bool { return len(s.entities) == 0 }
