package sflag

import (
	"errors"
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

func (g *Group) lookupByType(typ FlagType) []*Flag {
	keys, i := make([]string, len(g.types)), 0
	for k := range g.types {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		if g.Name == keys[i] {
			return true
		}
		if g.Name == keys[j] {
			return false
		}
		return keys[i] > keys[j]
	})

	flags := []*Flag{}
	for _, name := range keys {
		ft := g.types[name]
		if ft == typ {
			flags = append(flags, g.lookup(name))
		}
	}
	return flags
}

func (g *Group) lookup(name string) *Flag {
	_, ok := g.types[name]
	if ok {
		return g.flagSet.lookup(name)
	}
	return nil
}

func (g *Group) seen(name string) bool {
	_, o := g.flagSet.seen[name]
	_, k := g.types[name]
	return o && k
}

func LookupGroup(groups ...*Group) (*Group, error) {
	found := []*Group{}
	errs := []error{}
	for _, group := range groups {
		if !group.flagSet.parsed {
			return nil, fmt.Errorf("must call FlagSet.Parse() before LookupGroup()")
		}
		if err := group.validate(); err == nil {
			found = append(found, group)
		} else {
			errs = append(errs, err)
		}
	}

	if len(found) == 0 {
		return nil, errors.Join(errors.New("found no flag group"), errors.Join(errs...))
	}
	if len(found) > 1 {
		return groups[0], fmt.Errorf("found multiple flag groups: %v", found)
	}
	return found[0], nil
}

// nonGroupFlagNames returns flag name not belonging to the flag group.
func (g *Group) nonGroupFlagNames() []string {
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
	flags := g.lookupByType(TypeRequired)
	for _, flag := range flags {
		if !g.seen(flag.Name) {
			return fmt.Errorf("a required flag %s in the group %s is not specified", flag.Name, g.Name)
		}
	}

	diff := g.nonGroupFlagNames()
	if len(diff) > 0 {
		return fmt.Errorf("undefined flags %v in the group %s are specified", diff, g.Name)
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
		flags = append(flags, g.lookup(k))
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
