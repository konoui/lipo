package sflag

import (
	"fmt"
	"sort"
	"strings"
)

type FlagSet struct {
	name  string
	flags map[string]*Flag
	// map for a short name to a long name
	shortTo map[string]string
	args    []string
	// seen is a structure to handle a duplication error
	seen   map[string]struct{}
	Usage  func() string
	parsed bool
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

// lookup return the name is registered or not
func (f *FlagSet) lookup(name string) *Flag {
	flag, _ := f.isFlagName("-" + name)
	return flag
}

// NewGroup grouping flags.
// `name` is used to output an error when invalid flag combinations are specified.
// `name` should be one of flag names.
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
		fmt.Fprintf(b, "  -%s  *%s*", flag.Name, TypeRequired.String())
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
