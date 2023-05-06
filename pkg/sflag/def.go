package sflag

import (
	"fmt"
	"strconv"
)

// Bool represents `-flag` NOT `-flag true/false`
func (f *FlagSet) Bool(name, usage string, opts ...FlagOpt) *FlagRef[bool] {
	addOpts := append([]FlagOpt{WithDenyDuplicate()}, opts...)
	return Register(f,
		NewValue(new(bool), strconv.ParseBool),
		name, usage, addOpts...)
}

// String represents `-flag <value>`
func (f *FlagSet) String(name, usage string, opts ...FlagOpt) *FlagRef[string] {
	addOpts := append([]FlagOpt{WithDenyDuplicate()}, opts...)
	return Register(f,
		NewValue(new(string), func(v string) (string, error) { return v, nil }),
		name, usage, addOpts...)
}

// StringFlags represents `-flag <value1> -flag <value2> -flag <value3> -flag ...`
func (f *FlagSet) StringFlags(name, usage string, opts ...FlagOpt) *FlagRef[[]string] {
	return Register(f, newStringFlags(), name, usage, opts...)
}

func newStringFlags() *Value[[]string] {
	p := new([]string)
	convert := func(v string) ([]string, error) {
		*p = append(*p, v)
		return *p, nil
	}
	return NewValue(p, convert)
}

// Strings represents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) Strings(name, usage string, opts ...FlagOpt) *FlagRef[[]string] {
	addOpts := append([]FlagOpt{WithDenyDuplicate()}, opts...)
	return Register(f, newStrings(), name, usage, addOpts...)
}

func newStrings() *Value[[]string] {
	fv := newStringFlags()
	p := fv.p
	fv.capper = func() int {
		// require one value at least
		if len(*p) == 0 {
			return 1
		}
		return CapNoLimit
	}
	return fv
}

// FixedStringFlags represents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// This is a reference implementation
func (f *FlagSet) FixedStringFlags(name, usage string, opts ...FlagOpt) *FlagRef[[][2]string] {
	return Register(f, newFixedStringFlags(), name, usage, opts...)
}

func newFixedStringFlags() *Value[[][2]string] {
	p := new([][2]string)

	var idx, cur int
	maxLen := 2
	convert := func(v string) ([][2]string, error) {
		if cur >= maxLen {
			// Note the error will occur when cap() is not called properly to reset the cursor
			return nil, fmt.Errorf("cursor exceeded maximum length: cursor %d, max_len %d", cur, maxLen)
		}
		if len(*p) <= idx {
			*p = append(*p, [2]string{})
		}
		(*p)[idx][cur] = v
		cur++
		return *p, nil
	}
	cap := func() int {
		cap := maxLen - cur
		if cap == 0 {
			cur = 0
			idx++
		}
		return cap
	}
	return NewValues(p, convert, cap)
}
