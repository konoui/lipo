package lipo

func (l *Lipo) Archs() ([]string, error) {
	if err := validateOneInput(l.in); err != nil {
		return nil, err
	}

	bin := l.in[0]

	r, err := inspectFile(bin)
	if err != nil {
		return nil, err
	}
	return r.arches, nil
}
