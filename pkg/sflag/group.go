package sflag

import (
	"fmt"
	"strings"
)

const (
	TypeRequire = iota
	TypeOption
)

type Group struct {
	Name        string
	types       map[string]int
	flagSet     *FlagSet
	description string
}

func (g *Group) Add(flag *Flag, typ int) {
	if g.types == nil {
		g.types = make(map[string]int)
	}
	g.types[flag.Name] = typ
}

func (g *Group) AddDescription(s string) *Group {
	g.description = s
	return g
}

func (g *Group) LookupByType(typ int) []*Flag {
	flags := []*Flag{}
	for name, ft := range g.types {
		if ft == typ {
			flags = append(flags, g.Lookup(name))
		}
	}
	return flags
}

func (g *Group) Lookup(name string) *Flag {
	_, ok := g.types[name]
	if ok {
		return g.flagSet.Lookup(name)
	}
	return nil
}

func (g *Group) Seen(name string) bool {
	_, o := g.flagSet.seen[name]
	_, k := g.types[name]
	return o && k
}

func LookupGroup(groups ...*Group) (*Group, error) {
	found := []*Group{}
	errNames := []string{}
	for _, group := range groups {
		if err := group.validate(); err == nil {
			found = append(found, group)
			errNames = append(errNames, group.Name)
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("found no flag group")
	}
	if len(found) > 1 {
		return nil, fmt.Errorf("found multiple flag groups: %v", errNames)
	}
	return found[0], nil
}

// Diff returns non group flag with specified
func (g *Group) Diff() []string {
	diff := []string{}
	for name := range g.flagSet.seen {
		_, exist := g.types[name]
		if !exist {
			diff = append(diff, name)
		}
	}

	return diff
}

func (g *Group) validate() error {
	flags := g.LookupByType(TypeRequire)
	for _, flag := range flags {
		if !g.Seen(flag.Name) {
			return fmt.Errorf("required flag -%s is not specified", flag.Name)
		}
	}

	diff := g.Diff()
	if len(diff) > 0 {
		return fmt.Errorf("non group flag is specified: %v", diff)
	}

	return nil
}

func UsageFunc(groups ...*Group) func() string {
	return func() string {
		var b strings.Builder
		for _, g := range groups {
			b.WriteString(g.Usage())
			b.WriteString("\n")
		}
		return b.String()
	}
}

func (g *Group) Usage() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s:", g.Name))
	b.WriteString(strings.ReplaceAll(g.description, "\n", "\n  "))
	b.WriteString("\n")
	flags := make(map[string]*Flag)
	for k := range g.types {
		flags[k] = g.Lookup(k)
	}

	buildFlagsUsage(&b, flags)
	return b.String()
}
