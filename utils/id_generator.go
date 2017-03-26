package utils

import "time"

func UniqueId() (id uint64) {
	return uint64(time.Now().Unix())
}
