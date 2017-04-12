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
	pathMustExists(dir)
	filepath := filepath.Join(dir, strconv.FormatUint(id, 10)+".data")
	v = new(Volume)
	v.ID = id
	v.Path = dir
	v.File, err = os.OpenFile(filepath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("Open file: ", err)
	}
	v.Directory, err = NewLeveldbDirectory(dir)
	if err != nil {
		return nil, fmt.Errorf("Leveldb: ", err)
	}

	var oldCurrentIndex []byte = make([]byte, InitIndexSize)
	_, err = v.File.ReadAt(oldCurrentIndex, 0) // Read old current index from file
	if err != nil && err != io.EOF{
		return nil, err
	}
	oldCurrentIndexNum := binary.BigEndian.Uint64(oldCurrentIndex)
	if oldCurrentIndexNum > InitIndexSize {
		v.setCurrentIndex(oldCurrentIndexNum)
	} else {
		v.setCurrentIndex(InitIndexSize)
	}
	v.Size = MaxVolumeSize
	v.lock = sync.Mutex{}
	return
}

func (v *Volume) GetNeedle(id uint64) (n *Needle, err error) {
	n, err = v.Directory.Get(id)
	if err != nil {
		return
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
	_, err = v.GetNeedle(id)
	if err != nil && err != ErrDeleted{
		return err
	}
	return v.Directory.Del(id)
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
	n.Checksum = utils.Checksum(data)
	now := time.Now()
	n.CreatedAt = now
	n.UpdatedAt = now
	n.File = v.File
	n.FileExt = ext
	needleData, err := NeedleMarshal(n)
	if err != nil {
		return
	}
	v.File.WriteAt(needleData, int64(n.Offset))
	err = v.Directory.New(n)
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
	offset = v.CurrentOffset
	v.setCurrentIndex( v.CurrentOffset + HeaderSize(extsize) + filesize)
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
	v.lock.Lock()
	defer v.lock.Unlock()
	newFilePath := v.File.Name() + ".temp" // filepath.Join(v.Path, strconv.FormatUint(v.ID, 10)+".temp")
	newFile, err := os.OpenFile(newFilePath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return  fmt.Errorf("Open file: ", err)
	}
	iter := v.Directory.Iter()
	var key []byte = make([]byte, 8)
	var exists bool = true
	var currentOffset int64 = int64(InitIndexSize)
	for exists {
		key, exists = iter.Next()
		if !exists {
			break
		}
		id := binary.BigEndian.Uint64(key)
		needle, err := v.GetNeedle(id) // 读取一个needle
		if err != nil {
			fmt.Println("Read needle err: ", err)
			continue
		}
		offset := int64(needle.Offset)
		size := int64(needle.Size) + int64(HeaderSize(uint64(len(needle.FileExt))))
		var needleData []byte = make([]byte, size)
		v.File.ReadAt(needleData, offset)
		_, err = newFile.WriteAt(needleData, currentOffset) // 修改 offset， 并写到新文件中
		fmt.Println("Write: ", currentOffset, needleData)
		if err != nil {
			fmt.Println("Write err: ",err)
			continue
		}
		fmt.Println("Current offset: ", currentOffset)
		needle.Offset = uint64(currentOffset)
		v.Directory.Set(needle.ID, needle) // 更新 needle
		currentOffset += size
	}
	var currentOffsetByte = make([]byte, InitIndexSize)
	binary.BigEndian.PutUint64(currentOffsetByte, uint64(currentOffset))
	newFile.WriteAt(currentOffsetByte, 0)
	v.setCurrentIndex(uint64(currentOffset))
	fullpath:= v.File.Name()
	// 给老文件改名
	delpath := fullpath + ".del"
	fmt.Println("Old change name: ", fullpath, delpath)
	delFile , err := os.OpenFile(delpath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return
	}
	if delFile != nil {
		defer delFile.Close()
	}
	err = os.Rename(fullpath, delpath)
	if err != nil {
		return
	}
	err = os.Rename(newFilePath, fullpath)
	if err != nil {
		os.Rename(delpath, fullpath) // 出错， 把已经改掉的老文件的名字改回去
		return
	}
	v.File = newFile
	v.Path = newFilePath
	fmt.Println("New change name: ", delpath, fullpath)
	err = os.Remove(delpath)
	fmt.Println("Deleted: ", delpath)
	return
}


func (v *Volume) Print() {
	iter := v.Directory.Iter()
	var hasNext bool = true
	for hasNext {
		var key []byte
		key, hasNext = iter.Next()
		if hasNext {
			id := binary.BigEndian.Uint64(key)
			n, err := v.GetNeedle(id)
			if err != nil && err != ErrDeleted  {
				fmt.Println("Err: ", err)
			}
			fmt.Printf("id: %d, offset: %d \n", id, n.Offset)
			var b []byte = make([]byte, n.Size)
			dataOffset := n.Offset + HeaderSize(uint64(len(n.FileExt)))
			v.File.ReadAt(b, int64(dataOffset))
			fmt.Println("data: ", n.FileExt, string(b))
		} else {
			fmt.Println("-------- Finish-------")
		}
	}
	iter.Release()
}

func (v *Volume) CheckCurrentIndex() (same bool) {
	i1 := v.CurrentOffset
	b := make([]byte, InitIndexSize)
	v.File.ReadAt(b, 0)
	fmt.Println(b)
	i2 := binary.BigEndian.Uint64(b)
	return i1 == i2
}

func (v *Volume) ReadHeader(offset int64) (header []byte, err error) {
	var b []byte = make([]byte, NeedleFixSize)
	v.File.ReadAt(b, offset)
	fmt.Println(b)
	n, err := NeedleUnmarshal(b)
	if err != nil {
		return
	}
	header = make([]byte, len(b) + len(n.FileExt))
	v.File.ReadAt(header[:len(b)], offset)
	v.File.ReadAt(header[len(b):], offset + int64(len(b)))
	return
}

// 从完整文件名中获取扩展名
func Ext(filename string) (ext string) {
	index := strings.LastIndex(filename, ".")
	fmt.Println(index)
	if index == -1 {
		return ""
	}
	return strings.TrimSpace(filename[index+1:])
}

func pathExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

func pathMustExists(path string) {
	exists := pathExist(path)
	if !exists {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic("Path Must exists: " +  err.Error())
		}
	}
}

