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
	assert.NoError(t, err)
	v.Print()
	v.DelNeedle(1491730516)
	v.Print()
	v.Fragment()
	v.Print()
}