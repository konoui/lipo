package lcpu

import (
	"debug/macho"
	"fmt"
)

type CPUType = macho.Cpu

type SubCPUType uint32

func Cpu(v string) (cpu CPUType, sub SubCPUType, ok bool) {
	cs, ok := cpuSet[v]
	if ok {
		return cs.cpuType, cs.subType, true
	}
	return 0, 0, false
}

func CpuString(cpu CPUType, sub SubCPUType) string {
	maskedSub := (sub & ^MaskSubCPUType)
	id := id(cpu, sub)
	cs, ok := cpuIDSet[id]
	if ok {
		return cs.name
	}
	unknown := fmt.Sprintf("unknown(%d,%d)", cpu, maskedSub)
	return unknown
}

func CPUSubCPUString(c CPUType, s SubCPUType) (cpu string, sub string) {
	return toCPUStrings(c, s)
}

type cpu struct {
	cpuType CPUType
	subType SubCPUType
	name    string
}

func toCPUStrings(c CPUType, s SubCPUType) (cpu string, sub string) {
	var v string
	switch c {
	case CPUTypeI386:
		return "CPU_TYPE_I386", "CPU_SUBTYPE_I386_ALL"
	case CPUTypeX86_64:
		v = "CPU_TYPE_X86_64"
		switch s & ^MaskSubCPUType {
		case SubCPUTypeX86All:
			return v, "CPU_SUBTYPE_X86_64_ALL"
		case SubCPUTypeX86_64H:
			return v, "CPU_SUBTYPE_X86_64_H"
		}
	case CPUTypeArm:
		v = "CPU_TYPE_ARM"
		switch s {
		case SubCPUTypeArmV4T:
			return v, "CPU_SUBTYPE_ARM_V4T"
		case SubCPUTypeArmV6:
			return v, "CPU_SUBTYPE_ARM_V6"
		case SubCPUTypeArmV6M:
			return v, "CPU_SUBTYPE_ARM_V6M"
		case SubCPUTypeArmV7:
			return v, "CPU_SUBTYPE_ARM_V7"
		case SubCPUTypeArmV7F:
			return v, "CPU_SUBTYPE_ARM_V7F"
		case SubCPUTypeArmV7S:
			return v, "CPU_SUBTYPE_ARM_V7S"
		case SubCPUTypeArmV7K:
			return v, "CPU_SUBTYPE_ARM_V7K"
		case SubCPUTypeArmV7M:
			return v, "CPU_SUBTYPE_ARM_V7M"
		case SubCPUTypeArmV7EM:
			return v, "CPU_SUBTYPE_ARM_V7EM"
		case SubCPUTypeArmV8M:
			return v, "CPU_SUBTYPE_ARM_V8M"
		case SubCPUTypeArmAll:
			return v, "CPU_SUBTYPE_ARM_ALL"
		}
	case CPUTypeArm64:
		v = "CPU_TYPE_ARM64"
		switch s & ^MaskSubCPUType {
		case SubCPUTypeArm64All:
			return v, "CPU_SUBTYPE_ARM64_ALL"
		case SubCPUTypeArm64V8:
			return v, "CPU_SUBTYPE_ARM64_V8"
		case SubCPUTypeArm64E:
			return v, "CPU_SUBTYPE_ARM64E"
		}
	case CPUTypeArm64_32:
		v = "CPU_TYPE_ARM64_32"
		switch s & ^MaskSubCPUType {
		case SubCPUTypeArm64_32V8:
			return v, "CPU_SUBTYPE_ARM64_32_V8"
		}
	}

	if v == "" {
		v = fmt.Sprintf("%d", c)
	}
	return v, fmt.Sprintf("%d", s & ^MaskSubCPUType)
}
