package sflag

type FlagOption func(*Flag)

func WithDenyDuplicate() FlagOption {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func WithGroup(g *Group, typ int) FlagOption {
	return func(flag *Flag) {
		g.Add(flag, typ)
	}
}

func (f *FlagSet) Var(v Value, name, usage string, opts ...FlagOption) {
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
}

// Bool presents -flag NOT -flag true/false
func (f *FlagSet) Bool(p *bool, name, usage string, opts ...FlagOption) {
	opts = append(opts, WithDenyDuplicate())
	f.Var(Bool(p), name, usage, opts...)
}

// String presents -flag <value>
func (f *FlagSet) String(p *string, name, usage string, opts ...FlagOption) {
	opts = append(opts, WithDenyDuplicate())
	f.Var(String(p), name, usage, opts...)
}

// MultipleFlagString presents `-flag <value1> -flag <value2> -flag <value3>`
func (f *FlagSet) MultipleFlagString(p *[]string, name, usage string, opts ...FlagOption) {
	f.Var(MultipleFlagString(p), name, usage, opts...)
}

// FlexStrings presents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) FlexStrings(p *[]string, name, usage string, opts ...FlagOption) {
	opts = append(opts, WithDenyDuplicate())
	f.Var(FlexStrings(p), name, usage, opts...)
}

// FixedStrings presents `-flag <value1> <value2>` number of values are specified by initialize of a variable
// e.g. s := []string{make([]string, 2)}
// This is a reference implementation
func (f *FlagSet) FixedStrings(p *[]string, name, usage string, opts ...FlagOption) {
	opts = append(opts, WithDenyDuplicate())
	f.Var(FixedStrings(p), name, usage, opts...)
}

// MultipleFlagFixedStrings presents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// This is a reference implementation
func (f *FlagSet) MultipleFlagFixedStrings(p *[][2]string, name, usage string, opts ...FlagOption) {
	f.Var(MultipleFlagFixedStrings(p), name, usage, opts...)
}
