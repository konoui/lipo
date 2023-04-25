package sflag

import (
	"fmt"
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
		if len(f.args) < 1 {
			return false, fmt.Errorf("-%s flag: one value is required", flag.Name)
		}

		v := f.consumeArg()
		if err := value.Set(v); err != nil {
			return false, err
		}
		return true, nil
	}

	// limited cap case, consume num of remaining caps
	cap := values.Cap()
	for i := 0; i < cap; i++ {
		if len(f.args) == 0 {
			return false, fmt.Errorf("-%s flag: more values are required", flag.Name)
		}

		nextArg := f.args[0]
		_, isName := f.flags[flagName(nextArg)]
		if isName {
			return false, fmt.Errorf("-%s flag: more values are required", flag.Name)
		}

		v := f.consumeArg()
		if err := values.Set(v); err != nil {
			return false, err
		}
	}

	// check no limit case after limited case since transition of limit cap to no limit will occur
	cap = values.Cap()
	if cap == CapNoLimit {
		for {
			if len(f.args) == 0 {
				return false, nil
			}

			nextArg := f.args[0]
			_, isName := f.flags[flagName(nextArg)]
			if isName {
				return true, nil
			}

			v := f.consumeArg()
			if err := values.Set(v); err != nil {
				return false, err
			}
		}
	}
	// cap is limited
	return true, nil

}

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
