package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewVolume(t *testing.T) {
	v, err := NewVolume(1, "")
	assert.NoError(t, err)
	t.Logf("%+v", v)
}

func TestVolume_NewNeedle(t *testing.T) {
	v, err := NewVolume(1, "")
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, []byte("20"), "test.jpg")
	assert.NoError(t, err)
	t.Logf("%+v", n)
}

func TestExt(t *testing.T) {
	f1 := "test.jpg"
	t.Log(Ext(f1))
	f2 := "test2."
	t.Log(Ext(f2))
	f3 := ".ttt"
	t.Log(Ext(f3))
	f4 := "ddd.fff.ggg.hhh"
	t.Log(Ext(f4))
}

func TestVolume_Fragment(t *testing.T) {
	v, err := NewVolume(1, "/tmp/fs")
	id, err := v.NewFile([]byte("fff"), "f0")
	assert.NoError(t, err)
	v.Print()
	t.Log(id)
	v.DelNeedle(id)
	v.Print()
	fl, _ := v.File.Stat()
	sizeOld := fl.Size()
	err = v.Fragment()
	if err != nil {
		t.Log(err)
	}
	assert.NoError(t, err)
	v.Print()
	//has := v.Directory.Has(id)
	//assert.True(t, has)
	fl, _ = v.File.Stat()
	sizeNew := fl.Size()
	t.Log(sizeOld, sizeNew)
}

func TestVolume_PrintFile(t *testing.T) {
	v, err := NewVolume(1, "/tmp/fs")
	assert.NoError(t, err)
	//v.NewNeedle(1, []byte("20"), "test.jpg")
	header, err := v.ReadHeader(8)
	assert.NoError(t, err)
	t.Log(header, len(header))
}