package sflag

import (
	"errors"
)

type FlagGetter interface {
	Flag() *Flag
}

type FlagRef[T any] struct {
	flag *Flag
}

type FlagOption func(*Flag)

type Flag struct {
	Name          string
	ShortName     string
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

type flagValue[T any] struct {
	p       *T
	convert func(v string) (T, error)
}

type flagValues[T any] struct {
	flagValue[T]
	cap func() int
}

var (
	_ FlagGetter = &FlagRef[any]{}
	_ Value      = &flagValue[any]{}
	_ Values     = &flagValues[any]{}
)

const (
	CapNoLimit = -1
)

// Get returns a typed flag value
func (fr *FlagRef[T]) Get() T {
	return fr.flag.Value.Get().(T)
}

// Flag returns a Flag var to access flag name and usage.
// To access a flag value, use Get() instead of Flag.Value
func (fr *FlagRef[T]) Flag() *Flag {
	return fr.flag
}

// WithDenyDuplicate Parse() returns an error if encountering duplicated flags are specified.
// This is used for a custom flag definition as pre-defined flags(Bool/String etc) are enabled by default.
func WithDenyDuplicate() FlagOption {
	return func(flag *Flag) {
		flag.denyDuplicate = true
	}
}

func WithShortName(short string) FlagOption {
	return func(flag *Flag) {
		flag.ShortName = short
	}
}

// Register registers Value with name and usage as a Flag
// This is used to define a custom flag type.
func Register[T any](f *FlagSet, v Value, name, usage string, opts ...FlagOption) *FlagRef[T] {
	if _, ok := v.Get().(T); !ok {
		panic("Value.Get().[T] does not match Register[T]/FlagRef[T]")
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
	flag := &Flag{Name: name, Usage: usage, Value: v}
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

	return &FlagRef[T]{flag: flag}
}

// FlagValue converts (T and convert method) to Value of interface to define a custom flag type.
func FlagValue[T any](p *T, convert func(v string) (T, error)) Value {
	f := flagValue[T]{p: p, convert: convert}
	return &f
}

// FlagValue converts (T, convert and cap methods) to Value of interface to define a custom flag type.
func FlagValues[T any](p *T, convert func(v string) (T, error), cap func() int) Values {
	fvs := flagValues[T]{
		flagValue: flagValue[T]{p: p, convert: convert},
		cap:       cap,
	}
	return &fvs
}

// Set converts string value to T and stores it
func (fv *flagValue[T]) Set(v string) error {
	if fv.p == nil {
		return errors.New("pointer is empty")
	}

	value, err := fv.convert(v)
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

// Cap returns number of availabilities to store values
func (fv *flagValues[T]) Cap() int {
	return fv.cap()
}
