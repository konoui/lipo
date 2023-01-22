package sflag

type FlagOption func(*Flag)

func WithDenyDuplicate() FlagOption {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func WithGroup(g *Group, typ FlagType) FlagOption {
	return func(flag *Flag) {
		g.Add(flag, typ)
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
	P *T
	*Flag
}

// Bool presents `-flag“ NOT `-flag true/false`
func (f *FlagSet) Bool(p *bool, name, usage string, opts ...FlagOption) *Flag {
	opts = append(opts, WithDenyDuplicate())
	return f.Var(Bool(p), name, usage, opts...)
}

// String presents `-flag <value>`
func (f *FlagSet) String(p *string, name, usage string, opts ...FlagOption) *Flag {
	opts = append(opts, WithDenyDuplicate())
	return f.Var(String(p), name, usage, opts...)
}

// StringFlags presents `-flag <value1> -flag <value2> -flag <value3> -flag ...`
func (f *FlagSet) StringFlags(p *[]string, name, usage string, opts ...FlagOption) *Flag {
	return f.Var(StringFlags(p), name, usage, opts...)
}

// Strings presents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) Strings(p *[]string, name, usage string, opts ...FlagOption) *Flag {
	opts = append(opts, WithDenyDuplicate())
	return f.Var(Strings(p), name, usage, opts...)
}

// FixedStringFlags presents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// This is a reference implementation
func (f *FlagSet) FixedStringFlags(p *[][2]string, name, usage string, opts ...FlagOption) *Flag {
	return f.Var(FixedStringFlags(p), name, usage, opts...)
}
