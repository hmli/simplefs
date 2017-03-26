package api

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewServer(t *testing.T) {
	s := NewServer(22333, "/tmp/fs")
	t.Logf("%+v", s)
	assert.NotNil(t, s)
}

func TestServer_FileHandler(t *testing.T) {
	s := NewServer(22333, "/tmp/fs")
	//f := "/Users/blacksheep/work/src/simplefs/test.jpg"
	f := "/Users/li/work/src/simplefs/test.jpg"
	file, err := os.OpenFile(f, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		t.Error(err)
	}
	data, err := ioutil.ReadAll(file)
	assert.NoError(t, err)
	id, err := s.newFile(data, "testfile") //TODO bug here
	assert.NoError(t, err)
	needle, err := s.Volume.GetNeedle(id)
	assert.NoError(t, err)
	t.Logf("%+v", needle)
	t.Log(needle.File.Name())
	t.Log(needle.Size)
	data, ext, err := s.getFile(id)
	t.Log(data, ext, err)

}
