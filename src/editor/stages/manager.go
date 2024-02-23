package stages

import (
	"bytes"
	"kaiju/engine"
	"kaiju/filesystem"
)

type Manager struct {
	host  *engine.Host
	stage string
}

func NewManager(host *engine.Host) Manager {
	return Manager{host: host}
}

func (m *Manager) Save() error {
	if m.stage == "" {
		// TODO:  Show a dialog to get the stage name and block on it
	}
	stream := bytes.NewBuffer(make([]byte, 0))
	all := m.host.Entities()
	var err error = nil
	for i := 0; i < len(all) && err == nil; i++ {
		err = all[i].EditorSerialize(stream)
	}
	if err != nil {
		return err
	}
	return filesystem.WriteFile(m.stage, stream.Bytes())
}
