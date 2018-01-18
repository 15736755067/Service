package common

import (
	"testing"
)

type FS struct {
	Id     int64
	Values []int64
}

func (f FS) ID() int64 {
	return f.Id
}
func (f FS) VALUES() []int64 {
	return f.Values
}

type HasVALUES interface {
	VALUES() []int64
}

func TestDBCacheInit(t *testing.T) {
	afs := make([]FS, 10)
	for i, _ := range afs {
		afs[i].Id = int64(i)
		afs[i].Values = []int64{afs[i].Id + 1, afs[i].Id + 2, afs[i].Id + 3}
	}
	bfs := make([]HasID, 10)
	for i, _ := range bfs {
		bfs[i] = afs[i]
	}
	var dc DBCache
	dc.Init(bfs)
	t.Log(dc.Get(5).(HasVALUES).VALUES())
	afs[5].Values[0] = int64(100)
	if dc.Get(5).(HasVALUES).VALUES()[0] != 6 {
		t.Log(dc.Get(5).(HasVALUES).VALUES())
		t.Fail()
	}
	v5 := dc.Get(5).(FS)
	v5.Values[0] = -1
	y5 := dc.Get(5).(FS)
	y5.Values[0] = -2
	if v5.Values[0] == y5.Values[0] {
		t.Fail()
	}
}
