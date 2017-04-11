package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
	"os"
	"encoding/binary"
)

func TestLeveldbDirectory_leveldb(t *testing.T) {
	db, err := leveldb.OpenFile("directory", nil)
	assert.NoError(t, err)
	key := "keytest"
	value := "vtest"
	err = db.Put([]byte(key), []byte(value), nil)
	assert.NoError(t, err)
	v, err := db.Get([]byte(key), nil)
	fmt.Println(string(v))
	fmt.Println(err)
	assert.NoError(t, err)
	assert.Equal(t, value, string(v))
}

func TestLeveldbDirectory(t *testing.T) {
	os.RemoveAll("/tmp/lvldir")
	d, err := NewLeveldbDirectory("/tmp/lvldir")
	assert.NoError(t, err)
	iter := d.db.NewIterator(nil, nil)
	for iter.Next() {
		t.Log("DELETE:", iter.Key())
		d.db.Delete(iter.Key(), nil)
	}
	now := time.Now()
	var id uint64 = 3
	n := &Needle{
		ID:        id,
		Size:      20,
		Offset:    60,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = d.New(n)
	assert.NoError(t, err)
	newN, err := d.Get(id)
	assert.NoError(t, err)
	t.Log(newN)
	assert.Equal(t, n.ID, newN.ID)
	n.Offset = 70
	d.Set(id, n)
	setedN, err := d.Get(id)
	assert.NoError(t, err)
	t.Log(setedN)
	assert.Equal(t, int(setedN.Offset), 70)
	exists := d.Has(id)
	assert.True(t, exists)
	t.Log(exists)
	err = d.Del(id)
	assert.NoError(t, err)
	newExists := d.Has(id)
	assert.False(t, newExists)
}

func TestLeveldbDirectory_Iter(t *testing.T) {
	v, err  := NewVolume(1, "/tmp/iter")
	assert.NoError(t, err)
	id, err := v.NewFile([]byte("dde"), "dde1")
	assert.NoError(t, err)
	t.Log(id)
	iter := v.Directory.Iter()
	var key []byte = make([]byte, 8)
	var exists bool = true
	for exists {
		key, exists = iter.Next()
		t.Log(id, exists)
		if exists {
			id := binary.BigEndian.Uint64(key)
			t.Log(id)
		}
	}
	iter.Release()
}