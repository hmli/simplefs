package core

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"testing"
	"time"
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
	d, err := NewLeveldbDirectory("/tmp/lvldir")
	assert.NoError(t, err)
	iter := d.db.NewIterator(nil, nil)
	for iter.Next() {
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
	err = d.Set(n)
	assert.NoError(t, err)
	newN, err := d.Get(id)
	assert.NoError(t, err)
	t.Log(newN)
	assert.Equal(t, n.ID, newN.ID)
	exists := d.Has(id)
	assert.True(t, exists)
	id, has := d.Next()
	t.Log(id, has)
	assert.True(t, has)
	id2, has2 := d.Next()
	t.Log(id2, has2)
	assert.False(t, has2)
	err = d.Del(id)
	assert.NoError(t, err)
	newExists := d.Has(id)
	assert.False(t, newExists)
}
