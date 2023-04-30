package sflag

import (
	"fmt"
	"strconv"
)

// Bool presents `-flag“ NOT `-flag true/false`
func (f *FlagSet) Bool(name, usage string) *FlagRef[bool] {
	p := new(bool)
	flag := f.Var(FlagValue(p, strconv.ParseBool), name, usage, WithDenyDuplicate())
	ref := &FlagRef[bool]{flag: flag}
	return ref
}

// String presents `-flag <value>`
func (f *FlagSet) String(name, usage string) *FlagRef[string] {
	p := new(string)
	flag := f.Var(
		FlagValue(p, func(v string) (string, error) { return v, nil }),
		name, usage, WithDenyDuplicate())
	return &FlagRef[string]{flag: flag}
}

// StringFlags presents `-flag <value1> -flag <value2> -flag <value3> -flag ...`
func (f *FlagSet) StringFlags(name, usage string) *FlagRef[[]string] {
	p := new([]string)
	flag := f.Var(newStringFlags(p), name, usage)
	return &FlagRef[[]string]{flag: flag}
}

func newStringFlags(p *[]string) Value {
	cur := 0
	convert := func(v string) ([]string, error) {
		if cur < len(*p) {
			(*p)[cur] = v
		} else {
			*p = append(*p, v)
		}
		cur++
		return *p, nil
	}
	return FlagValue(p, convert)
}

// Strings presents `-flag <value1> <value2> <value3> ...`
func (f *FlagSet) Strings(name, usage string) *FlagRef[[]string] {
	p := new([]string)
	flag := f.Var(newStrings(p), name, usage, WithDenyDuplicate())
	return &FlagRef[[]string]{flag: flag}
}

func newStrings(p *[]string) Value {
	fv := newStringFlags(p).(*flagValue[[]string])
	cap := func() int {
		if len(*p) == 0 {
			return 1
		}
		return CapNoLimit
	}
	return FlagValues(p, fv.convert, cap)
}

// FixedStringFlags presents `-flag <value1> <value2> -flag <value3> <value4> -flag ...`
// This is a reference implementation
func (f *FlagSet) FixedStringFlags(name, usage string) *FlagRef[[][2]string] {
	p := new([][2]string)
	flag := f.Var(newFixedStringFlags(p), name, usage)
	return &FlagRef[[][2]string]{flag: flag}
}

func newFixedStringFlags(p *[][2]string) Value {
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
	return FlagValues(p, convert, cap)
}
