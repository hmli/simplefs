package utils

import "github.com/klauspost/crc32"

func Checksum(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}
