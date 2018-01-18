package common

import (
	"sync"
)

type HasID interface {
	ID() int64
}
type DBCache struct {
	m sync.RWMutex
	v bool
	d map[int64]HasID
}

func (dc DBCache) Get(id int64) HasID {
	dc.m.RLock()
	defer dc.m.RUnlock()

	return dc.d[id]
}

func (dc *DBCache) Init(data []HasID) {
	if dc == nil {
		return
	}
	dc.m.Lock()
	defer dc.m.Unlock()

	dc.v = true
	dc.d = make(map[int64]HasID, len(data))
	for i := range data {
		dc.d[data[i].ID()] = data[i]
	}
}

func (dc *DBCache) Clear() {
	if dc == nil {
		return
	}
	dc.m.Lock()
	defer dc.m.Unlock()

	dc.v = false
	dc.d = make(map[int64]HasID)
}

func (dc DBCache) Enabled() bool {
	dc.m.RLock()
	defer dc.m.RUnlock()
	return dc.v
}

type DBSliceCache struct {
	m sync.RWMutex
	v bool
	d map[int64][]HasID
}

func (dc DBSliceCache) Get(id int64) []HasID {
	dc.m.RLock()
	defer dc.m.RUnlock()

	return dc.d[id]
}


func (dc *DBSliceCache) Init(data map[int64][]HasID) {
	if dc == nil {
		return
	}
	dc.m.Lock()
	defer dc.m.Unlock()

	dc.v = true
	dc.d = make(map[int64][]HasID, len(data))
	for k := range data {
		dc.d[k] = data[k]
	}
}

func (dc *DBSliceCache) Clear() {
	if dc == nil {
		return
	}
	dc.m.Lock()
	defer dc.m.Unlock()

	dc.v = false
	dc.d = make(map[int64][]HasID)
}

func (dc DBSliceCache) Enabled() bool {
	dc.m.RLock()
	defer dc.m.RUnlock()

	return dc.v
}
