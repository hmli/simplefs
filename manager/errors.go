package manager

import "github.com/kataras/iris/errors"

var (
	ErrVidRepeat = errors.New("Volume id repeated")
	ErrNoVolume = errors.New("No this volume")
)
