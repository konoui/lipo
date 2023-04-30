package sflag

import (
	"fmt"
	"sort"
	"strings"
)

type FlagType int

func (t FlagType) String() string {
	switch t {
	case TypeRequired:
		return "required"
	case TypeOptional:
		return "optional"
	case typeNotDefined:
		return "not-defined"
	}
	return "unknown"
}

const (
	TypeRequired FlagType = iota + 1
	TypeOptional
	typeNotDefined
)

type Group struct {
	Name        string
	types       map[string]FlagType
	flagSet     *FlagSet
	description string
}

func (g *Group) String() string {
	return g.Name
}

func (g *Group) AddRequired(fg FlagGetter) *Group {
	return g.add(fg, TypeRequired)
}

func (g *Group) AddOptional(fg FlagGetter) *Group {
	return g.add(fg, TypeOptional)
}

func (g *Group) add(fg FlagGetter, typ FlagType) *Group {
	if g.types == nil {
		g.types = make(map[string]FlagType)
	}

	flag := fg.Flag()
	g.types[flag.Name] = typ
	return g
}

func (g *Group) AddDescription(s string) *Group {
	g.description = s
	return g
}

func (g *Group) LookupByType(typ FlagType) []*Flag {
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
	for _, group := range groups {
		if !group.flagSet.parsed {
			return nil, fmt.Errorf("must call FlagSet.Parse() before LookupGroup()")
		}
		if err := group.validate(); err == nil {
			found = append(found, group)
		}
	}

	if len(found) == 0 {
		return nil, fmt.Errorf("found no flag group from %v", groups)
	}
	if len(found) > 1 {
		return groups[0], fmt.Errorf("found multiple flag groups: %v", found)
	}
	return found[0], nil
}

// NonGroupFlagNames returns flag name not belonging to the flag group.
func (g *Group) NonGroupFlagNames() []string {
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
	flags := g.LookupByType(TypeRequired)
	for _, flag := range flags {
		if !g.Seen(flag.Name) {
			return fmt.Errorf("required flag in the group is not specified: %s", flag.Name)
		}
	}

	diff := g.NonGroupFlagNames()
	if len(diff) > 0 {
		return fmt.Errorf("undefined flags in the group are specified: %v", diff)
	}

	return nil
}

func UsageFunc(groups ...*Group) func() string {
	return func() string {
		var b strings.Builder
		for _, g := range groups {
			b.WriteString("\n")
			b.WriteString(g.Usage())
		}
		return b.String()
	}
}

func (g *Group) Usage() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s:", g.Name))
	b.WriteString(strings.ReplaceAll(g.description, "\n", "\n  "))
	b.WriteString("\n")
	flags := make([]*Flag, 0, len(g.types))
	for k := range g.types {
		flags = append(flags, g.Lookup(k))
	}

	// sort by flag type
	sort.Slice(flags, func(i, j int) bool {
		iv, jv := g.types[flags[i].Name], g.types[flags[j].Name]
		if iv == jv {
			return flags[i].Name < flags[j].Name
		}
		return iv < jv
	})

	for _, flag := range flags {
		buildFlagUsage(&b, flag, g.types[flag.Name])
	}
	return b.String()
}
