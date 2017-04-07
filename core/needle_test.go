package core

import (
	"encoding/binary"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNeedleMarshal(t *testing.T) {
	v, err := NewVolume(1, "")
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, []byte("20"), "test.jpg")
	assert.NoError(t, err)
	//var id uint64 = 3
	//now := time.Now()
	//n := &Needle{
	//	ID: id,
	//	Size: 20,
	//	Offset: 60,
	//	CreatedAt: now,
	//	UpdatedAt: now,
	//}
	data, err := NeedleMarshal(n)
	assert.NoError(t, err)
	newN, err := NeedleUnmarshal(data)
	assert.NoError(t, err)
	t.Log(newN)
	assert.NotNil(t, newN)
	assert.Equal(t, n.ID, newN.ID)
}

func TestNeedle_ReadWrite(t *testing.T) {
	v, err := NewVolume(1, "")
	assert.NoError(t, err)
	n, err := v.NewNeedle(1, []byte("20"), "twrite.jpg")
	assert.NoError(t, err)

	data := []byte("dddbbeifls3fff")
	num, err := n.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), num)

	var readByte []byte = make([]byte, num)
	num, err = n.Read(readByte)
	assert.NoError(t, err)
	t.Log(string(readByte))
	assert.Equal(t, readByte, data)
}

func TestNeedle_MultiReadWrite(t *testing.T) {
	v, err := NewVolume(1, "")
	assert.NoError(t, err)
	for i := 0; i < 10; i++ {
		n, err := v.NewNeedle(uint64(i), []byte("20"), "d.jpg")
		assert.NoError(t, err)
		var data []byte = make([]byte, 4)
		binary.BigEndian.PutUint32(data, uint32(i*i))
		num, err := n.Write(data)
		assert.NoError(t, err)
		assert.Equal(t, len(data), num)
	}
	for i := 0; i < 10; i++ {
		n, err := v.GetNeedle(uint64(i))
		assert.NoError(t, err)
		t.Logf("%+v", n)
		var readByte []byte = make([]byte, 4)
		_, err = n.Read(readByte)
		assert.NoError(t, err)
		origin := binary.BigEndian.Uint32(readByte)
		t.Log(origin)
		assert.Equal(t, origin, uint32(i*i))
	}
}
