package manager

import (
	"github.com/hmli/simplefs/core"
)

type Manager struct {
	volumes map[uint64]*core.Volume
}

func (m *Manager) NewVolume(id uint64, dir string) (volume *core.Volume, err error) {
	volume, err = core.NewVolume(id, dir)
	if err != nil {
		return
	}
	_, exists := m.volumes[id]
	if exists {
		return nil, ErrVidRepeat
	}
	m.volumes[id] = volume
	return
}

func (m *Manager) GetVolume(id uint64) (volume *core.Volume, err error) {
	volume, exists := m.volumes[id]
	if !exists {
		return nil, ErrNoVolume
	}
	return
}


