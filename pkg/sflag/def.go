package sflag

type FlagOption func(*Flag)

func WithDenyDuplicate() FlagOption {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func (f *FlagSet) Var(v Value, name, usage string, opts ...FlagOption) *Flag {
	if name == "" {
		panic("empty flag name is registered")
	}

	if f.flags == nil {
		f.flags = make(map[string]*Flag)
	}

	if f.seen == nil {
		f.seen = make(map[string]struct{})
	}

	_, exists := f.flags[name]
	if exists {
		panic("duplicate flag name is registered")
	}

	flag := &Flag{Name: name, Usage: usage, Value: v}
	f.flags[name] = flag

	for _, opt := range opts {
		if opt != nil {
			opt(flag)
		}
	}
	return flag
}

type FlagRef[T any] struct {
	Value func() T
	Flag  func() *Flag
}

// Bool presents `-flag“ NOT `-flag true/false`
func (f *FlagSet) Bool(name, usage string) FlagRef[bool] {
	p := new(bool)
	flag := f.Var(Bool(p), name, usage, WithDenyDuplicate())
	ref := FlagRef[bool]{
		Flag:  func() *Flag { return flag },
		Value: func() bool { return flag.Value.Get().(bool) },
	}
	return ref
}

// String presents `-flag <value>`
func (f *FlagSet) String(name, usage string) FlagRef[string] {
	p := new(string)
	flag := f.Var(String(p), name, usage, WithDenyDuplicate())
	ref := FlagRef[string]{
		Flag:  func() *Flag { return flag },
		Value: func() string { return flag.Value.Get().(string) },
	}
	return ref
}

// StringFlags presents `-flag <value1> -flag <value2> -flag <value3> -flag ...`
func (f *FlagSet) StringFlags(name, usage string) FlagRef[[]string] {
	p := new([]string)
	flag := f.Var(StringFlags(p), name, usage)
	ref := FlagRef[[]string]{
		Flag:  func() *Flag { return flag },
		Value: func() []string { return flag.Value.Get().([]string) },
	}
	return ref
}

// Strings presents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) Strings(name, usage string) FlagRef[[]string] {
	p := new([]string)
	flag := f.Var(Strings(p), name, usage, WithDenyDuplicate())
	ref := FlagRef[[]string]{
		Flag:  func() *Flag { return flag },
		Value: func() []string { return flag.Value.Get().([]string) },
	}
	return ref
}

// FixedStringFlags presents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// This is a reference implementation
func (f *FlagSet) FixedStringFlags(name, usage string) FlagRef[[][2]string] {
	p := new([][2]string)
	flag := f.Var(FixedStringFlags(p), name, usage)
	ref := FlagRef[[][2]string]{
		Flag:  func() *Flag { return flag },
		Value: func() [][2]string { return flag.Value.Get().([][2]string) },
	}
	return ref
}
