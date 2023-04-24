package lmacho

import (
	"debug/macho"
	"fmt"
)

type SubCpu = uint32

func IsSupportedCpu(v string) bool {
	_, ok := cpuNameSet[v]
	return ok
}

func ToCpu(v string) (cpu macho.Cpu, sub SubCpu, ok bool) {
	cs, ok := cpuNameSet[v]
	if ok {
		return macho.Cpu(cs.t), cs.s, true
	}
	return 0, 0, false
}

func ToCpuString(cpu macho.Cpu, subCpu SubCpu) string {
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
	{t: uint32(CpuTypeI386), s: SubCpuTypeX86All, v: "i386"},
	{t: uint32(CpuTypeX86_64), s: SubCpuTypeX86_64All, v: "x86_64"},
	{t: uint32(CpuTypeX86_64), s: SubCpuTypeX86_64H, v: "x86_64h"},
	// arm
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmAll, v: "arm"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV4T, v: "armv4t"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV6, v: "armv6"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7, v: "armv7"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7F, v: "armv7f"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7S, v: "armv7s"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7K, v: "armv7k"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV6M, v: "armv6m"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7M, v: "armv7m"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV7EM, v: "armv7em"},
	{t: uint32(CpuTypeArm), s: SubCpuTypeArmV8M, v: "armv8m"},
	// arm64
	{t: uint32(CpuTypeArm64), s: SubCpuTypeArm64All, v: "arm64"},
	{t: uint32(CpuTypeArm64), s: SubCpuTypeArm64E, v: "arm64e"},
	{t: uint32(CpuTypeArm64), s: SubCpuTypeArm64V8, v: "arm64v8"},
	// arm64_32
	{t: uint32(CpuTypeArm64_32), s: SubCpuTypeArm64_32V8, v: "arm64_32"},
}

// /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach/machine.h
const cpuArch64 = 0x01000000
const cpuArch64_32 = 0x02000000

const (
	// skip
	CpuTypeX86    macho.Cpu = 7
	CpuTypeI386   macho.Cpu = CpuTypeX86
	CpuTypeX86_64 macho.Cpu = CpuTypeI386 | cpuArch64
	// skip
	CpuTypeArm      macho.Cpu = 12
	CpuTypeArm64    macho.Cpu = CpuTypeArm | cpuArch64
	CpuTypeArm64_32 macho.Cpu = CpuTypeArm | cpuArch64_32
	CpuTypePpc      macho.Cpu = 18
	CpuTypePpc64    macho.Cpu = CpuTypePpc | 64
	// skip
)

const MaskSubCpuType SubCpu = 0xff000000

const (
	SubCpuTypeX86All    SubCpu = 3
	SubCpuTypeX86_64All SubCpu = 3
	SubCpuTypeX86Arch1  SubCpu = 4
	SubCpuTypeX86_64H   SubCpu = 8
)

const (
	SubCpuTypeArmAll SubCpu = 0
	SubCpuTypeArmV4T SubCpu = 5
	SubCpuTypeArmV6  SubCpu = 6
	// skip
	SubCpuTypeArmV7  SubCpu = 9
	SubCpuTypeArmV7F SubCpu = 10
	SubCpuTypeArmV7S SubCpu = 11
	SubCpuTypeArmV7K SubCpu = 12
	// skip
	SubCpuTypeArmV6M  SubCpu = 14
	SubCpuTypeArmV7M  SubCpu = 15
	SubCpuTypeArmV7EM SubCpu = 16
	SubCpuTypeArmV8M  SubCpu = 17
)

const (
	SubCpuTypeArm64_32V8 SubCpu = 1
)

const (
	SubCpuTypeArm64All SubCpu = 0
	SubCpuTypeArm64V8  SubCpu = 1
	SubCpuTypeArm64E   SubCpu = 2
)

func ToCpuValues(c macho.Cpu, s SubCpu) (cpu string, sub string) {
	var v string
	switch c {
	case CpuTypeI386:
		return "CPU_TYPE_I386", "CPU_SUBTYPE_I386_ALL"
	case CpuTypeX86_64:
		v = "CPU_TYPE_X86_64"
		switch s & ^MaskSubCpuType {
		case SubCpuTypeX86All:
			return v, "CPU_SUBTYPE_X86_64_ALL"
		case SubCpuTypeX86_64H:
			return v, "CPU_SUBTYPE_X86_64_H"
		}
	case CpuTypeArm:
		v = "CPU_TYPE_ARM"
		switch s {
		case SubCpuTypeArmV4T:
			return v, "CPU_SUBTYPE_ARM_V4T"
		case SubCpuTypeArmV6:
			return v, "CPU_SUBTYPE_ARM_V6"
		case SubCpuTypeArmV6M:
			return v, "CPU_SUBTYPE_ARM_V6M"
		case SubCpuTypeArmV7:
			return v, "CPU_SUBTYPE_ARM_V7"
		case SubCpuTypeArmV7F:
			return v, "CPU_SUBTYPE_ARM_V7F"
		case SubCpuTypeArmV7S:
			return v, "CPU_SUBTYPE_ARM_V7S"
		case SubCpuTypeArmV7K:
			return v, "CPU_SUBTYPE_ARM_V7K"
		case SubCpuTypeArmV7M:
			return v, "CPU_SUBTYPE_ARM_V7M"
		case SubCpuTypeArmV7EM:
			return v, "CPU_SUBTYPE_ARM_V7EM"
		case SubCpuTypeArmV8M:
			return v, "CPU_SUBTYPE_ARM_V8M"
		case SubCpuTypeArmAll:
			return v, "CPU_SUBTYPE_ARM_ALL"
		}
	case CpuTypeArm64:
		v = "CPU_TYPE_ARM64"
		switch s & ^MaskSubCpuType {
		case SubCpuTypeArm64All:
			return v, "CPU_SUBTYPE_ARM64_ALL"
		case SubCpuTypeArm64V8:
			return v, "CPU_SUBTYPE_ARM64_V8"
		case SubCpuTypeArm64E:
			return v, "CPU_SUBTYPE_ARM64E"
		}
	case CpuTypeArm64_32:
		v = "CPU_TYPE_ARM64_32"
		switch s & ^MaskSubCpuType {
		case SubCpuTypeArm64_32V8:
			return v, "CPU_SUBTYPE_ARM64_32_V8"
		}
	}

	if v == "" {
		v = fmt.Sprintf("%d", c)
	}
	return v, fmt.Sprintf("%d", s & ^MaskSubCpuType)
}
