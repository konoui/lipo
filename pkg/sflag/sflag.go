package sflag

import (
	"errors"
	"fmt"
	"strconv"
)

type Flag struct {
	Name          string
	Usage         string
	Value         Value
	denyDuplicate bool
}

type Value interface {
	Set(string) error
	Get() any
}

type Values interface {
	Value
	Cap() int
}

const (
	CapNoLimit = -1
)

type flagValue[T any] struct {
	p    *T
	from func(v string) (T, error)
}

type flagValues[T any] struct {
	flagValue[T]
	cap func() int
}

func FlagValue[T any](p *T, from func(v string) (T, error)) Value {
	f := flagValue[T]{p: p, from: from}
	return &f
}

func FlagValues[T any](p *T, from func(v string) (T, error), cap func() int) Values {
	fvs := flagValues[T]{
		flagValue: flagValue[T]{p: p, from: from},
		cap:       cap,
	}
	return &fvs
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

// Get returns a flag value NOT a pointer of value
func (fv *flagValue[T]) Get() any {
	return *fv.p
}

func (fv *flagValues[T]) Cap() int {
	return fv.cap()
}

// ---- definitions

func String(p *string) Value {
	return FlagValue(p, func(v string) (string, error) { return v, nil })
}

func Bool(p *bool) Value {
	return FlagValue(p, strconv.ParseBool)
}

func StringFlags(p *[]string) Value {
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

func Strings(p *[]string) Value {
	fv := StringFlags(p).(*flagValue[[]string])
	cap := func() int {
		if len(*p) == 0 {
			return 1
		}
		return CapNoLimit
	}
	return FlagValues(p, fv.from, cap)
}

func FixedStringFlags(p *[][2]string) Value {
	var idx, cur int
	maxLen := 2
	from := func(v string) ([][2]string, error) {
		if cur >= maxLen {
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
	return FlagValues(p, from, cap)
}
