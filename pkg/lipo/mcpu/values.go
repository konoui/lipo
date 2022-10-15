package mcpu

import (
	"debug/macho"
	"fmt"
)

func StringValues(c macho.Cpu, s uint32) (string, string) {
	switch c {
	case TypeI386:
		return "CPU_TYPE_I386", "CPU_SUBTYPE_I386_ALL"
	case TypeX86_64:
		v := "CPU_TYPE_X86_64"
		switch s & ^MaskSubType {
		case SubTypeX86All:
			return v, "CPU_SUBTYPE_X86_64_ALL"
		case SubTypeX86_64H:
			return v, "CPU_SUBTYPE_X86_64_H"
		}
	case TypeArm:
		v := "CPU_TYPE_ARM"
		switch s {
		case SubTypeArmV4T:
			return v, "CPU_SUBTYPE_ARM_V4T"
		case SubTypeArmV6:
			return v, "CPU_SUBTYPE_ARM_V6"
		case SubTypeArmV6M:
			return v, "CPU_SUBTYPE_ARM_V6M"
		case SubTypeArmV7:
			return v, "CPU_SUBTYPE_ARM_V7"
		case SubTypeArmV7F:
			return v, "CPU_SUBTYPE_ARM_V7F"
		case SubTypeArmV7S:
			return v, "CPU_SUBTYPE_ARM_V7S"
		case SubTypeArmV7K:
			return v, "CPU_SUBTYPE_ARM_V7K"
		case SubTypeArmV7M:
			return v, "CPU_SUBTYPE_ARM_V7M"
		case SubTypeArmV7EM:
			return v, "CPU_SUBTYPE_ARM_V7EM"
		case SubTypeArmV8M:
			return v, "CPU_SUBTYPE_ARM_V8M"
		case SubTypeArmAll:
			return v, "CPU_SUBTYPE_ARM_ALL"
		}
	case TypeArm64:
		v := "CPU_TYPE_ARM64"
		switch s & ^MaskSubType {
		case SubTypeArm64All:
			return v, "CPU_SUBTYPE_ARM64_ALL"
		case SubTypeArm64V8:
			return v, "CPU_SUBTYPE_ARM64_V8"
		case SubTypeArm64E:
			return v, "CPU_SUBTYPE_ARM64E"
		}
	}

	return fmt.Sprintf("%d", c), fmt.Sprintf("%d", s & ^MaskSubType)
}
