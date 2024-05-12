package lmacho

import (
	"errors"
)

var (
	ErrThin = errors.New("the file is thin file, not fat")
)
