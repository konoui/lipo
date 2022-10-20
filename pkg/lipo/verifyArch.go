package lipo

import "github.com/konoui/lipo/pkg/util"

func (l *Lipo) VerifyArch(arches ...string) (bool, error) {
	gotArches, err := l.Archs()
	if err != nil {
		return false, err
	}

	for _, a := range arches {
		if !util.Contains(gotArches, a) {
			return false, nil
		}
	}
	return true, nil
}
