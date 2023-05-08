package sflag

import (
	"errors"
)

type FlagGetter interface {
	Flag() *Flag
}

// FlagRef is a wrapper to access the Flag and the flag value.
type FlagRef[T any] struct {
	flag *Flag
	v    *Value[T]
}

type FlagOpt func(*Flag)

type Flag struct {
	Name          string
	ShortName     string
	Usage         string
	value         value
	denyDuplicate bool
}

// Value please use NewValue(s) function instead of the definition
type Value[T any] struct {
	p         *T
	converter func(v string) (T, error)
	capper    func() int
}

// value is an internal interface which is used by the Parse() of FlagSet to set values
type value interface {
	set(string) error
	cap() int
}

var (
	_ FlagGetter = &FlagRef[any]{}
	_ value      = &Value[any]{}
)

const (
	CapNoLimit = -1
)

// Get returns a typed flag value
func (fr *FlagRef[T]) Get() T {
	return fr.v.get()
}

// Flag returns the Flag to access the flag name and the usage.
func (fr *FlagRef[T]) Flag() *Flag {
	return fr.flag
}

// WithDenyDuplicate Parse() returns an error if encountering duplicated flags are specified.
// This is used for a custom flag definition. Pre-defined flags(Bool/String etc) are enabled by default.
func WithDenyDuplicate() FlagOpt {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func WithShortName(short string) FlagOpt {
	return func(flag *Flag) {
		flag.ShortName = short
	}
}

// Register registers a Value with a Name and an Usage as a Flag
// This is used to define a custom flag type.
func Register[T any](f *FlagSet, v *Value[T], name, usage string, opts ...FlagOpt) *FlagRef[T] {
	if v.p == nil {
		panic("the value pointer is nil")
	}

	if name == "" {
		panic("the flag name is empty string")
	}

	if f.flags == nil {
		f.flags = make(map[string]*Flag)
	}

	if f.seen == nil {
		f.seen = make(map[string]struct{})
	}

	if f.shortTo == nil {
		f.shortTo = make(map[string]string)
	}

	_, exists := f.flags[name]
	if exists {
		panic("the flag name is duplicate in the registration process")
	}
	flag := &Flag{Name: name, Usage: usage, value: v}
	f.flags[name] = flag

	for _, opt := range opts {
		if opt != nil {
			opt(flag)
		}
	}

	if sname := flag.ShortName; sname != "" {
		_, exists := f.shortTo[sname]
		if exists {
			panic("the short flag name is duplicate in the registration process")
		}
		f.shortTo[sname] = name
	}

	return &FlagRef[T]{flag: flag, v: v}
}

// NewValue is used for a single value definition. e.g.) bool, string, int.
// When implementing `-flag value1 -flag value2 -flag value3`, it can be also useful.
func NewValue[T any](p *T, converter func(v string) (T, error)) *Value[T] {
	fv := Value[T]{p: p, converter: converter, capper: func() int { return 1 }}
	return &fv
}

// NewValues is used for a slice definition.
// e.g. -flag value1 value2 value3
func NewValues[T any](p *T, converter func(v string) (T, error), capper func() int) *Value[T] {
	fv := Value[T]{p: p, converter: converter, capper: capper}
	return &fv
}

// get returns a flag value NOT a pointer of value
func (fv *Value[T]) get() T {
	return *fv.p
}

// set converts string value to T and stores it
func (fv *Value[T]) set(v string) error {
	if fv.p == nil {
		return errors.New("the value pointer is nil")
	}

	value, err := fv.converter(v)
	if err != nil {
		return err
	}

	*fv.p = value
	return nil
}

// cap returns number of capabilities to store values
// When it returns CapNoLimit, Parse() will consume number of available arguments to set values.
func (fv *Value[T]) cap() int {
	return fv.capper()
}
