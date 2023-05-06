package sflag

import (
	"errors"
	"fmt"
)

const (
	fmtRequireOneValueAtLeast = "the -%s flag requires one value at least"
	fmtRequireMoreValues      = "the -%s flag requires %d values at least"
)

func (f *FlagSet) Parse(args []string) error {
	if f.parsed {
		return errors.New("Prase() has already been called")
	}
	f.args = args
	newArgs := []string{}
	for {
		isFlag, err := f.parse()
		if err != nil {
			return err
		}
		if !isFlag {
			if len(f.args) == 0 {
				break
			}
			newArgs = append(newArgs, f.consumeArg())
		}
	}
	f.args = newArgs
	f.parsed = true
	return nil
}

// parse returns false when a next argument is not a flag
func (f *FlagSet) parse() (isFlag bool, _ error) {
	if len(f.args) == 0 {
		return false, nil
	}

	flag, ok := f.isFlagName(f.args[0])
	if !ok {
		return false, nil
	}

	name := flag.Name
	defer func() { f.seen[name] = struct{}{} }()

	_, seen := f.seen[name]
	if seen && flag.denyDuplicate {
		return false, fmt.Errorf("duplication: more than one -%s flag specified", flag.Name)
	}

	// update and skip flag name
	f.consumeArg()

	value := flag.value
	// special case, value is not required
	if _, ok := value.(*Value[bool]); ok {
		if err := value.set("true"); err != nil {
			return false, err
		}
		return true, nil
	}

	// limited-cap case, consume num of remaining caps
	cap := value.cap()
	for i := 0; i < cap; i++ {
		if len(f.args) == 0 {
			if cap == 1 {
				return false, fmt.Errorf(fmtRequireOneValueAtLeast, flag.Name)
			}
			return false, fmt.Errorf(fmtRequireMoreValues, flag.Name, cap)
		}

		if _, ok := f.isFlagName(f.args[0]); ok {
			if cap == 1 {
				return false, fmt.Errorf(fmtRequireOneValueAtLeast, flag.Name)
			}
			return false, fmt.Errorf(fmtRequireMoreValues, flag.Name, cap)
		}

		v := f.consumeArg()
		if err := value.set(v); err != nil {
			return false, err
		}
	}

	// check non-limit case after limited-cap case since a transition of limited-cap to non-limit will occur
	cap = value.cap()
	if cap == CapNoLimit {
		for {
			if len(f.args) == 0 {
				return false, nil
			}

			if _, ok := f.isFlagName(f.args[0]); ok {
				return true, nil
			}

			v := f.consumeArg()
			if err := value.set(v); err != nil {
				return false, err
			}
		}
	}
	// limited cap case
	return true, nil

}

// isFlagName returns a Flag and True if `s` is registered.
// `s` must contain hyphen .e.g. `-flag` NOT `flag`
func (f *FlagSet) isFlagName(s string) (*Flag, bool) {
	name := flagName(s)
	if name == "" {
		return nil, false
	}

	flag, exist := f.flags[name]
	if exist {
		return flag, true
	}

	long, exist := f.shortTo[name]
	if !exist {
		return nil, false
	}

	flag, exist = f.flags[long]
	if exist {
		return flag, true
	}
	return nil, false
}

// flagName checks `s` is registered flag name or not.
// if not flag name, return empty string, otherwise it returns flag name without a hyphen
func flagName(s string) string {
	if len(s) < 1 || s[0] != '-' {
		return ""
	}
	name := s[len("-"):]
	return name
}

func (f *FlagSet) consumeArg() (arg string) {
	arg, f.args = f.args[0], f.args[1:]
	return arg
}
