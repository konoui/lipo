package sflag

import (
	"fmt"
)

type FlagOption func(*Flag)

func WithDenyDuplicate() FlagOption {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func (f *FlagSet) Var(v Value, name, usage string, opts ...FlagOption) {
	if name == "" {
		fmt.Fprintf(f.out, "Warning: skip register due to empty flag name\n")
		return
	}

	if f.flags == nil {
		f.flags = make(map[string]*Flag)
	}

	if f.seen == nil {
		f.seen = make(map[string]struct{})
	}

	_, exists := f.flags[name]
	if exists {
		fmt.Fprintf(f.out, "Warning: duplicate flag name %s\n", name)
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
func (f *FlagSet) Bool(p *bool, name, usage string) {
	f.Var(Bool(p), name, usage, WithDenyDuplicate())
}

// String presents -flag <value>
func (f *FlagSet) String(p *string, name, usage string) {
	f.Var(String(p), name, usage, WithDenyDuplicate())
}

// MultipleFlagString presents `-flag <value1> -flag <value2> -flag <value3>`
func (f *FlagSet) MultipleFlagString(p *[]string, name, usage string) {
	f.Var(MultipleFlagString(p), name, usage)
}

// FlexStrings presents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) FlexStrings(p *[]string, name, usage string) {
	f.Var(FlexStrings(p), name, usage, WithDenyDuplicate())
}

// FixedStrings presents `-flag <value1> <value2>` number of values are specified by initialize of a variable
// e.g. s := []string{make([]string, 2)}
// This is a reference implementation
func (f *FlagSet) FixedStrings(p *[]string, name, usage string) {
	f.Var(FixedStrings(p), name, usage, WithDenyDuplicate())
}

// MultipleFlagFixedStrings presents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// e.g. s := [][]string{make([]string, 2)}
// This is a reference implementation
func (f *FlagSet) MultipleFlagFixedStrings(p *[][]string, name, usage string) {
	f.Var(MultipleFlagFixedStrings(p), name, usage)
}
