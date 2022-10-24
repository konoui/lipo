package sflag

import (
	"fmt"
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
	Get() any
}

type Values interface {
	Value
	Cap() int
}

const (
	CapNoLimit = -1
)

type FlagSet struct {
	name  string
	flags map[string]*Flag
	args  []string
	seen  map[string]struct{}
	Usage func() string
}

func NewFlagSet(name string) *FlagSet {
	f := &FlagSet{
		name: name,
	}
	f.Usage = f.usage
	return f
}

func (f *FlagSet) Args() []string {
	return f.args
}

func (f *FlagSet) Lookup(name string) *Flag {
	return f.flags[name]
}

func (f *FlagSet) NewGroup(name string) *Group {
	return &Group{Name: name, flagSet: f}
}

func (f *FlagSet) usage() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("usage: %s:\n", f.name))

	for _, flag := range sortFlags(f.flags) {
		buildFlagUsage(&b, flag, typeNotDefined)
	}
	return b.String()
}

func buildFlagUsage(b *strings.Builder, flag *Flag, typ FlagType) {
	if typ == TypeRequired {
		fmt.Fprintf(b, "  -%s  *%s*", flag.Name, typ.String())
	} else {
		fmt.Fprintf(b, "  -%s", flag.Name)
	}

	usage := flag.Usage

	if b.Len() <= 4 { // space, space, '-', 'x'.
		b.WriteString("\t")
	} else {
		b.WriteString("\n    \t")
	}
	b.WriteString(strings.ReplaceAll(usage, "\n", "\n    \t"))
	b.WriteString("\n")
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
