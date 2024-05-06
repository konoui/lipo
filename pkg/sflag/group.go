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

// lookup returns a Flag if the flag name is registered in the group and the flag set
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
		if selected := selectError(errs); len(selected) > 0 {
			return nil, errors.Join(selected...)
		}
		return nil, errors.Join(errors.New("please check the usage of the command. found no flag group"), errors.Join(errs...))
	}
	if len(found) > 1 {
		return groups[0], fmt.Errorf("found multiple flag groups: %v", found)
	}
	return found[0], nil
}

const fmtUniqueNotFound = "%s: -%s is not specified"
const fmtRequiredNotFound = "%s: -%s is required"
const fmtUndefinedFound = "%s: %v are undefined"

// selectError selects errors that have unique flag
func selectError(errors []error) []error {
	selected := []error{}
	for _, e := range errors {
		if !strings.HasSuffix(e.Error(), "is not specified") {
			selected = append(selected, e)
		}
	}
	return selected
}

func (g *Group) validate() error {
	hasUnique := g.lookup(g.Name) != nil
	if hasUnique && !g.seen(g.Name) {
		return fmt.Errorf(fmtUniqueNotFound, g.Name, g.Name)
	}

	flags := g.lookupByType(TypeRequired)
	for _, flag := range flags {
		if !g.seen(flag.Name) {
			return fmt.Errorf(fmtRequiredNotFound, g.Name, flag.Name)
		}
	}

	diff := g.nonGroupFlagNames()
	if len(diff) > 0 {
		return fmt.Errorf(fmtUndefinedFound, g.Name, diff)
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
