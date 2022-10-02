package sflag

import (
	"errors"
	"fmt"
	"strconv"
)

func FlagValue[T any](p *T, from func(v string) (T, error)) Value {
	f := flagValue[T]{p: p, from: from}
	return &f
}

func FlagValues[T any](p *T, from func(v string) (T, error), cap func() int) Value {
	fvs := flagValues[T]{
		flagValue: flagValue[T]{p: p, from: from},
		cap:       cap,
	}
	return &fvs
}

type flagValue[T any] struct {
	p    *T
	from func(v string) (T, error)
}

type flagValues[T any] struct {
	flagValue[T]
	cap func() int
}

func (fv *flagValue[T]) Set(v string) error {
	if fv.p == nil {
		return errors.New("pointer is empty")
	}

	value, err := fv.from(v)
	if err != nil {
		return err
	}

	*fv.p = value
	return nil
}

func (fv *flagValues[T]) Cap() int {
	return fv.cap()
}

// ---- definitions

func String(p *string) Value {
	return FlagValue(p, func(v string) (string, error) { return v, nil })
}

func Bool(p *bool) Value {
	return FlagValue(p, func(v string) (bool, error) { return strconv.ParseBool(v) })
}

func MultipleFlagString(p *[]string) Value {
	cur := 0
	from := func(v string) ([]string, error) {
		if cur < len(*p) {
			(*p)[cur] = v
		} else {
			*p = append(*p, v)
		}
		cur++
		return *p, nil
	}
	return FlagValue(p, from)
}

func FlexStrings(p *[]string) Value {
	fv := MultipleFlagString(p).(*flagValue[[]string])
	cap := func() int {
		if len(*p) == 0 {
			return 1
		}
		return CapNoLimit
	}
	return FlagValues(p, fv.from, cap)
}

func FixedStrings(p *[]string) Value {
	var cur int
	maxLen := len((*p)[0])
	from := func(v string) ([]string, error) {
		if cur >= maxLen {
			return nil, fmt.Errorf("fill error. cur %d, len %d", cur, maxLen)
		}
		(*p)[cur] = v
		cur++
		return *p, nil
	}
	return FlagValues(p, from, func() int { return maxLen - cur })
}

func MultipleFlagFixedStrings(p *[][]string) Value {
	var idx, cur int
	maxLen := len((*p)[0])
	from := func(v string) ([][]string, error) {
		if cur >= maxLen {
			return nil, fmt.Errorf("fill error. cur %d, len %d", cur, maxLen)
		}
		if len(*p) <= idx {
			*p = append(*p, make([]string, maxLen))
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
	return FlagValues(p, from, cap)
}
