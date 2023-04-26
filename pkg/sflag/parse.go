package sflag

import (
	"fmt"
)

const (
	fmtRequireOneValue        = "the -%s flag requires one value"
	fmtRequireOneValueAtLeast = fmtRequireOneValue + " at least"
	fmtRequireMoreValues      = "the -%s flag requires %d values at least"
)

func (f *FlagSet) Parse(args []string) error {
	f.args = args
	newArgs := []string{}
	for {
		ok, err := f.parse()
		if err != nil {
			return err
		}
		if !ok {
			if len(f.args) == 0 {
				break
			}
			newArgs = append(newArgs, f.consumeArg())
		}
	}
	f.args = newArgs
	return nil
}

func (f *FlagSet) parse() (bool, error) {
	if len(f.args) == 0 {
		return false, nil
	}

	name := flagName(f.args[0])
	if name == "" {
		return false, nil
	}

	flag, exist := f.flags[name]
	if !exist {
		return false, nil
	}
	defer func() { f.seen[name] = struct{}{} }()

	_, seen := f.seen[name]
	if seen && flag.denyDuplicate {
		return false, fmt.Errorf("duplication: more than one -%s flag specified", flag.Name)
	}

	// update and skip flag name
	f.consumeArg()

	value := flag.Value
	// special case, value is not required
	if _, ok := value.(*flagValue[bool]); ok {
		if err := value.Set("true"); err != nil {
			return false, err
		}
		return true, nil
	}

	values, isValues := value.(Values)
	if !isValues {
		if len(f.args) == 0 {
			return false, fmt.Errorf(fmtRequireOneValue, flag.Name)
		}

		if f.isNextArgFlag() {
			return false, fmt.Errorf(fmtRequireOneValue, flag.Name)
		}

		v := f.consumeArg()
		if err := value.Set(v); err != nil {
			return false, err
		}
		return true, nil
	}

	// limited-cap case, consume num of remaining caps
	cap := values.Cap()
	for i := 0; i < cap; i++ {
		if len(f.args) == 0 {
			if cap == 1 {
				return false, fmt.Errorf(fmtRequireOneValueAtLeast, flag.Name)
			}
			return false, fmt.Errorf(fmtRequireMoreValues, flag.Name, cap)
		}

		if f.isNextArgFlag() {
			if cap == 1 {
				return false, fmt.Errorf(fmtRequireOneValueAtLeast, flag.Name)
			}
			return false, fmt.Errorf(fmtRequireMoreValues, flag.Name, cap)
		}

		v := f.consumeArg()
		if err := values.Set(v); err != nil {
			return false, err
		}
	}

	// check non-limit case after limited-cap case since a transition of limited-cap to non-limit will occur
	cap = values.Cap()
	if cap == CapNoLimit {
		for {
			if len(f.args) == 0 {
				return false, nil
			}

			if f.isNextArgFlag() {
				return true, nil
			}

			v := f.consumeArg()
			if err := values.Set(v); err != nil {
				return false, err
			}
		}
	}
	// limited cap case
	return true, nil

}

// flagName checks `s` is registered flag name or not.
// if not flag name, return empty string, otherwise it returns flag name without hyphen
func flagName(s string) string {
	if len(s) < 2 || s[0] != '-' {
		return ""
	}
	name := s[len("-"):]
	return name
}

func (f *FlagSet) consumeArg() (arg string) {
	arg, f.args = f.args[0], f.args[1:]
	return arg
}

// isNextArgFlag return true if next arg is registered flag name
func (f *FlagSet) isNextArgFlag() bool {
	if len(f.args) == 0 {
		return false
	}
	nextArg := f.args[0]
	_, isFlag := f.flags[flagName(nextArg)]
	return isFlag
}
