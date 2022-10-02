package sflag

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type Flag struct {
	Name          string
	Usage         string
	Value         Value
	denyDuplicate bool
}

type Value interface {
	Set(string) error
}

type Values interface {
	Value
	Cap() int
}

const (
	CapNoLimit = -1
)

type FlagSet struct {
	Usage func()
	name  string
	flags map[string]*Flag
	args  []string
	out   io.Writer
	seen  map[string]struct{}
}

type Option func(f *FlagSet)

func WithOut(w io.Writer) Option {
	return func(f *FlagSet) {
		f.out = w
	}
}

func NewFlagSet(name string, opts ...Option) *FlagSet {
	f := &FlagSet{
		name: name,
		out:  os.Stderr,
	}

	f.Usage = f.printUsage
	for _, opt := range opts {
		if opt != nil {
			opt(f)
		}
	}
	return f
}

func (f *FlagSet) Args() []string {
	return f.args
}

func (f *FlagSet) Out() io.Writer {
	return f.out
}

func (f *FlagSet) Lookup(name string) *Flag {
	return f.flags[name]
}

func (f *FlagSet) printUsage() {
	fmt.Fprintf(f.Out(), "usage: %s:\n", f.name)

	for _, flag := range sortFlags(f.flags) {
		var b strings.Builder
		fmt.Fprintf(&b, "  -%s", flag.Name) // Two spaces before -; see next two comments.
		name, usage := flag.Name, flag.Usage
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		// Boolean flags of one ASCII letter are so common we
		// treat them specially, putting their usage on the same line.
		if b.Len() <= 4 { // space, space, '-', 'x'.
			b.WriteString("\t")
		} else {
			// Four spaces before the tab triggers good alignment
			// for both 4- and 8-space tab stops.
			b.WriteString("\n    \t")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))
		fmt.Fprint(f.Out(), b.String(), "\n")
	}
}

func sortFlags(flags map[string]*Flag) []*Flag {
	result := make([]*Flag, len(flags))
	i := 0
	for _, f := range flags {
		result[i] = f
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
