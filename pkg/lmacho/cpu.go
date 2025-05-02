package lmacho

import (
	"debug/macho"
	"fmt"
)

type SubCpu = uint32
type Cpu = macho.Cpu

func IsSupportedCpu(v string) bool {
	_, ok := cpuNameSet[v]
	return ok
}

func ToCpu(v string) (cpu Cpu, sub SubCpu, ok bool) {
	cs, ok := cpuNameSet[v]
	if ok {
		return Cpu(cs.t), cs.s, true
	}
	return 0, 0, false
}

func ToCpuString(cpu Cpu, subCpu SubCpu) string {
	maskedSub := (subCpu & ^MaskSubCpuType)
	id := id(uint32(cpu), subCpu)
	cs, ok := cpuIDSet[id]
	if ok {
		return cs.v
	}
	unknown := fmt.Sprintf("unknown(%d,%d)", cpu, maskedSub)
	return unknown
}

func CpuNames() []string {
	cpus := make([]string, len(cpuNames))
	for i, c := range cpuNames {
		cpus[i] = c.v
	}
	return cpus
}

type cpuName struct {
	t uint32
	s uint32
	v string
}

var (
	cpuNameSet = map[string]*cpuName{}
	cpuIDSet   = map[uint64]*cpuName{}
)

func init() {
	for i := range cpuNames {
		cpuNameSet[cpuNames[i].v] = &cpuNames[i]
		id := id(cpuNames[i].t, cpuNames[i].s)
		cpuIDSet[id] = &cpuNames[i]
	}
}

func id(t, s uint32) uint64 {
	s &= ^MaskSubCpuType
	return (uint64(t) << 32) | uint64(s)
}

var cpuNames = []cpuName{
	{t: uint32(TypeI386), s: SubTypeX86All, v: "i386"},
	{t: uint32(TypeX86_64), s: SubTypeX86_64All, v: "x86_64"},
	{t: uint32(TypeX86_64), s: SubTypeX86_64H, v: "x86_64h"},
	// arm
	{t: uint32(TypeArm), s: SubTypeArmAll, v: "arm"},
	{t: uint32(TypeArm), s: SubTypeArmV4T, v: "armv4t"},
	{t: uint32(TypeArm), s: SubTypeArmV6, v: "armv6"},
	{t: uint32(TypeArm), s: SubTypeArmV7, v: "armv7"},
	{t: uint32(TypeArm), s: SubTypeArmV7F, v: "armv7f"},
	{t: uint32(TypeArm), s: SubTypeArmV7S, v: "armv7s"},
	{t: uint32(TypeArm), s: SubTypeArmV7K, v: "armv7k"},
	{t: uint32(TypeArm), s: SubTypeArmV6M, v: "armv6m"},
	{t: uint32(TypeArm), s: SubTypeArmV7M, v: "armv7m"},
	{t: uint32(TypeArm), s: SubTypeArmV7EM, v: "armv7em"},
	{t: uint32(TypeArm), s: SubTypeArmV8M, v: "armv8m"},
	// arm64
	{t: uint32(TypeArm64), s: SubTypeArm64All, v: "arm64"},
	{t: uint32(TypeArm64), s: SubTypeArm64E, v: "arm64e"},
	{t: uint32(TypeArm64), s: SubTypeArm64V8, v: "arm64v8"},
	// arm64_32
	{t: uint32(TypeArm64_32), s: SubTypeArm64_32V8, v: "arm64_32"},
}

// /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach/machine.h
const CPUArch64 = 0x01000000 /* 64 bit ABI */
const cpuArch64_32 = 0x02000000

const (
	// skip
	TypeX86    Cpu = 7
	TypeI386   Cpu = TypeX86
	TypeX86_64 Cpu = TypeI386 | CPUArch64
	// skip
	TypeArm      Cpu = 12
	TypeArm64    Cpu = TypeArm | CPUArch64
	TypeArm64_32 Cpu = TypeArm | cpuArch64_32
	TypePpc      Cpu = 18
	CTypePpc64   Cpu = TypePpc | 64
	// skip
)

const MaskSubCpuType SubCpu = 0xff000000

const (
	SubTypeX86All    SubCpu = 3
	SubTypeX86_64All SubCpu = 3
	SubTypeX86Arch1  SubCpu = 4
	SubTypeX86_64H   SubCpu = 8
)

const (
	SubTypeArmAll SubCpu = 0
	SubTypeArmV4T SubCpu = 5
	SubTypeArmV6  SubCpu = 6
	// skip
	SubTypeArmV7  SubCpu = 9
	SubTypeArmV7F SubCpu = 10
	SubTypeArmV7S SubCpu = 11
	SubTypeArmV7K SubCpu = 12
	// skip
	SubTypeArmV6M  SubCpu = 14
	SubTypeArmV7M  SubCpu = 15
	SubTypeArmV7EM SubCpu = 16
	SubTypeArmV8M  SubCpu = 17
)

const (
	SubTypeArm64_32V8 SubCpu = 1
)

const (
	SubTypeArm64All SubCpu = 0
	SubTypeArm64V8  SubCpu = 1
	SubTypeArm64E   SubCpu = 2
)

func ToCpuValues(c Cpu, s SubCpu) (cpu string, sub string) {
	var v string
	switch c {
	case TypeI386:
		return "CPU_TYPE_I386", "CPU_SUBTYPE_I386_ALL"
	case TypeX86_64:
		v = "CPU_TYPE_X86_64"
		switch s & ^MaskSubCpuType {
		case SubTypeX86All:
			return v, "CPU_SUBTYPE_X86_64_ALL"
		case SubTypeX86_64H:
			return v, "CPU_SUBTYPE_X86_64_H"
		}
	case TypeArm:
		v = "CPU_TYPE_ARM"
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
		v = "CPU_TYPE_ARM64"
		switch s & ^MaskSubCpuType {
		case SubTypeArm64All:
			return v, "CPU_SUBTYPE_ARM64_ALL"
		case SubTypeArm64V8:
			return v, "CPU_SUBTYPE_ARM64_V8"
		case SubTypeArm64E:
			return v, "CPU_SUBTYPE_ARM64E"
		}
	case TypeArm64_32:
		v = "CPU_TYPE_ARM64_32"
		switch s & ^MaskSubCpuType {
		case SubTypeArm64_32V8:
			return v, "CPU_SUBTYPE_ARM64_32_V8"
		}
	}

	if v == "" {
		v = fmt.Sprintf("%d", c)
	}
	return v, fmt.Sprintf("%d", s & ^MaskSubCpuType)
}
