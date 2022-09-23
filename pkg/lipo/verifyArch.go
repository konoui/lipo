package lipo

func (l *Lipo) VerifyArch(arch string) (bool, error) {
	arches, err := l.archs()
	if err != nil {
		return false, err
	}

	for _, a := range arches {
		if a == arch {
			return true, nil
		}
	}
	return false, nil
}
