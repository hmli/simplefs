package core

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"io/ioutil"
	"github.com/hmli/simplefs/utils"
)

const (
	TruncateSize  uint64 = 1 << 30            //1GB
	MaxVolumeSize uint64 = 128 * TruncateSize // 128GB
	InitIndexSize uint64 = 8
	DefaultDir    string = "/tmp/fs"
)

// TODO Compression
// TODO Several volumes make up a Store
type Volume struct {
	ID            uint64
	File          *os.File
	Directory     Directory
	Size          uint64
	Path          string
	CurrentOffset uint64 // 从文件中恢复的时候， 第一个byte就是currentoffset. 每次写入文件也要更新此offset。
	lock          sync.Mutex
}

func NewVolume(id uint64, dir string) (v *Volume, err error) {
	if dir == "" {
		dir = DefaultDir
	}
	// TODO check dir exists. Create one if not exists
	path := filepath.Join(dir, strconv.FormatUint(id, 10)+".data")
	v = new(Volume)
	v.ID = id
	v.File, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Open file: ", err)
	}
	v.Directory, err = NewLeveldbDirectory(dir)
	if err != nil {
		return nil, fmt.Errorf("Leveldb: ", err)
	}

	var oldCurrentIndex []byte = make([]byte, InitIndexSize)
	_, err = v.File.ReadAt(oldCurrentIndex, 0) // Read old current index from file
	if err != nil {
		if err != io.EOF {
			return nil, err
		}

	}
	oldCurrentIndexNum := binary.BigEndian.Uint64(oldCurrentIndex)
	if oldCurrentIndexNum > InitIndexSize {
		v.setCurrentIndex(oldCurrentIndexNum)
	} else {
		v.setCurrentIndex(InitIndexSize)
	}
	v.Size = MaxVolumeSize
	v.Path = path
	v.lock = sync.Mutex{}
	return
}

func (v *Volume) GetNeedle(id uint64) (n *Needle, err error) {
	n, err = v.Directory.Get(id)
	if err != nil {
		return
	}
	if n.IsDeleted() {
		return nil, ErrDeleted
	}
	n.File = v.File
	return
}

func (v *Volume) GetFile(id uint64) (data []byte, ext string, err error) {
	needle, err := v.GetNeedle(id)
	if err != nil {
		return data, ext, fmt.Errorf("Get needle ", err)
	}
	ext = needle.FileExt
	data, err = ioutil.ReadAll(needle)
	if err != nil {
		return nil, "", err
	}
	checksum := utils.Checksum(data)
	if checksum != needle.Checksum {
		return nil, "", ErrWrongCheckSum
	}
	return
}


func (v *Volume) DelNeedle(id uint64) (err error) {
	v.lock.Lock()
	defer v.lock.Unlock()
	// 1. Get & check exists
	// 2. Free space
	// 3. Del key in directory
	n, err := v.GetNeedle(id)
	if err != nil {
		return
	}
	if !n.IsDeleted() {
		n.UpdatedAt = time.Now()
		n.Flag = 1
		v.Directory.Del(id)
	}
	return
}

// NewNeedle allocate a new needle.
func (v *Volume) NewNeedle(id uint64, data []byte, filename string) (n *Needle, err error) {
	size := uint64(len(data))
	v.lock.Lock()
	defer v.lock.Unlock()
	// 1. alloc space
	// 2. set needle's header
	// 3. make needle
	ext := Ext(filename)
	offset, err := v.allocSpace(size, uint64(len(ext)))
	if err != nil {
		return nil, err
	}
	n = new(Needle)
	n.ID = id
	n.Size = size
	n.Offset = offset
	n.Flag = 0
	n.Checksum = utils.Checksum(data)
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	n.File = v.File
	n.FileExt = ext
	err = v.Directory.Set(n)
	if err != nil {
		err = fmt.Errorf("Leveldb: ", err)
	}
	return n, err
}

func (v *Volume) NewFile(data []byte, filename string) (id uint64, err error) {
	id = utils.UniqueId()
	//needle, err := v.NewNeedle(id, uint64(len(data)), filename)
	needle, err := v.NewNeedle(id, data, filename)

	if err != nil {
		return id, fmt.Errorf("New needle : ", err)
	}
	_, err = needle.Write(data)
	if err != nil {
		return
	}
	return id, err
}

func (v *Volume) currentOffset() (offset uint64, err error) {
	var i []byte = make([]byte, 8)
	_, err = v.File.ReadAt(i, 0)
	if err != nil {
		return 0, err
	}
	offset = binary.BigEndian.Uint64(i)
	return
}

func (v *Volume) allocSpace(filesize uint64, extsize uint64) (offset uint64, err error) {
	remainingSize := v.RemainingSpace()
	totalSize := filesize + HeaderSize(extsize)
	if remainingSize < totalSize {
		return v.CurrentOffset, ErrLeakSpace
	}
	offset = v.CurrentOffset + HeaderSize(extsize)
	v.setCurrentIndex(offset + filesize)
	return
}

func (v *Volume) setCurrentIndex(currentOffset uint64) (err error) {
	v.CurrentOffset = currentOffset
	var offsetByte []byte = make([]byte, 8) // starts with 1 byte storing current index
	binary.BigEndian.PutUint64(offsetByte, v.CurrentOffset)
	_, err = v.File.WriteAt(offsetByte, 0)
	return
}

func (v *Volume) RemainingSpace() (size uint64) {
	return v.Size - v.CurrentOffset
}

func (v *Volume) Fragment() (err error) {
	return
}

// 从完整文件名中获取扩展名
func Ext(filename string) (ext string) {
	index := strings.LastIndex(filename, ".")
	if index == len(filename)-1 {
		return ""
	}
	return strings.TrimSpace(filename[index+1:])
}
