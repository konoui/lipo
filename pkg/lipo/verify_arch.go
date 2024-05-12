package lipo

import "github.com/konoui/lipo/pkg/util"

func (l *Lipo) VerifyArch(arches ...string) (bool, error) {
	gotArches, err := l.Archs()
	if err != nil {
		return false, err
	}

	m := util.ExistsMap(gotArches, func(a string) string { return a })
	for _, a := range arches {
		if _, ok := m[a]; !ok {
			return false, nil
		}
	}
	return true, nil
}
