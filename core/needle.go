package core

import (
	"encoding/binary"
	"io"
	"os"
	"time"
)

// 64 * 3 + 8 + 32 + 64*2 = 360 bits = 45 bytes for an needle header
var NeedleFixSize uint64 = 45 // 不包括 len(Filename)

// Needle in Haystack
type Needle struct {
	ID        uint64 // 唯一ID， 64
	Size      uint64 // size of BODY
	Offset    uint64 // points to start of header
	Flag      uint8  // 0: normal, 1: deleted
	File      *os.File
	FileExt   string // 本以为不需要, 但接口中返回时还需要这个字段
	Checksum  uint32 // TODO checksum
	rOffset   uint64 // 用在 Read() 函数里的
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (n *Needle) IsDeleted() bool {
	return n.Flag == 1
}

func (n *Needle) Read(b []byte) (num int, err error) {
	start := n.Offset + HeaderSize(uint64(len(n.FileExt))) + n.rOffset
	end := start + n.Size
	length := end - start
	if uint64(len(b)) > length {
		b = b[:length]
	}
	num, err = n.File.ReadAt(b, int64(start))
	n.rOffset += uint64(num)
	if n.rOffset > n.Size {
		err = io.EOF
	}
	return
}

func (n *Needle) Write(b []byte) (num int, err error) {
	start := n.Offset + HeaderSize(uint64(len(n.FileExt))) + n.rOffset
	end := start + n.Size
	length := end - start
	if uint64(len(b)) > length { // 这个needle 预分配的空间不足以写入
		return 0, ErrSmallNeedle
	} else {
		num, err = n.File.WriteAt(b, int64(start))
		n.rOffset += uint64(num)
	}
	return
}

// NeedleMarshal: Needle struct -> bytes
func NeedleMarshal(n *Needle) (data []byte, err error) {
	if n == nil {
		err = ErrNilNeedle
		return
	}
	data = make([]byte, HeaderSize(uint64(len(n.FileExt))))
	binary.BigEndian.PutUint64(data[0:8], n.ID)
	binary.BigEndian.PutUint64(data[8:16], n.Size)
	binary.BigEndian.PutUint64(data[16:24], n.Offset)
	binary.BigEndian.PutUint32(data[24:28], n.Checksum)
	binary.BigEndian.PutUint64(data[28:36], uint64(n.CreatedAt.Unix()))
	binary.BigEndian.PutUint64(data[36:44], uint64(n.UpdatedAt.Unix()))
	data[44] = byte(n.Flag)
	copy(data[45:], []byte(n.FileExt))
	//data[45:] = []byte(n.FileExt)
	return
}

// NeedleUnmarshal: bytes -> needle struct
func NeedleUnmarshal(b []byte) (n *Needle, err error) {
	if len(b) < 45 {
		return nil, ErrWrongLen
	}
	n = new(Needle)
	n.ID = binary.BigEndian.Uint64(b[0:8])
	n.Size = binary.BigEndian.Uint64(b[8:16])
	n.Offset = binary.BigEndian.Uint64(b[16:24])
	n.Checksum = binary.BigEndian.Uint32(b[24:28])
	n.CreatedAt = time.Unix(int64(binary.BigEndian.Uint64(b[28:36])), 0)
	n.UpdatedAt = time.Unix(int64(binary.BigEndian.Uint64(b[36:44])), 0)
	n.Flag = uint8(b[44])
	n.FileExt = string(b[45:])
	return
}


// HeaderSize Needle 中除了正文外的额外信息的大小
func HeaderSize(extsize uint64) (size uint64) {
	return extsize + NeedleFixSize
}
