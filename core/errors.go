package core

import "errors"

var (
	ErrNilNeedle   = errors.New("Nil needle")
	ErrWrongLen    = errors.New("Wrong length of needle bytes ")
	ErrLeakSpace   = errors.New("Volume leak of space")
	ErrDeleted     = errors.New("Needle is deleted")
	ErrSmallNeedle = errors.New("Needle's size less than data size")
)
