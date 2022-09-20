package lipo

func (l *Lipo) Create() error {
	inputs, err := newInputs(l.in...)
	if err != nil {
		return err
	}
	out, err := newOutput(l.out, inputs)
	if err != nil {
		return err
	}
	return out.create()
}
