package lipo

func (l *Lipo) VerifyArch(arches ...string) (bool, error) {
	gotArches, err := l.Archs()
	if err != nil {
		return false, err
	}

	for _, a := range arches {
		if !contain(a, gotArches) {
			return false, nil
		}
	}
	return true, nil
}
