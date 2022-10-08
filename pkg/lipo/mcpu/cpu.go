package mcpu

import (
	"debug/macho"
	"fmt"
)

func IsSupported(v string) bool {
	_, ok := cpuNameSet[v]
	return ok
}

func ToCpu(v string) (cpu macho.Cpu, sub uint32, ok bool) {
	cs, ok := cpuNameSet[v]
	if ok {
		return macho.Cpu(cs.t), cs.s, true
	}
	return 0, 0, false
}

func ToString(cpu macho.Cpu, subCpu uint32) string {
	maskedSub := (subCpu & ^MaskSubType)
	id := id(uint32(cpu), subCpu)
	cs, ok := cpuIDSet[id]
	if ok {
		return cs.v
	}
	unknown := fmt.Sprintf("unknown(%d,%d)", cpu, maskedSub)
	return unknown
}

func CpuNames() []string {
	cpus := make([]string, 0, len(cpuNames))
	for _, c := range cpuNames {
		cpus = append(cpus, c.v)
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
	s = s & ^MaskSubType
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
	{t: uint32(TypeArm64_32), s: SubTypeArm64_32All, v: "arm64_32"},
}

// /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach/machine.h
const cpuArch64 = 0x01000000
const cpuArch64_32 = 0x02000000

const (
	// skip
	TypeX86    macho.Cpu = 7
	TypeI386   macho.Cpu = TypeX86
	TypeX86_64 macho.Cpu = TypeI386 | cpuArch64
	// skip
	TypeArm      macho.Cpu = 12
	TypeArm64    macho.Cpu = TypeArm | cpuArch64
	TypeArm64_32 macho.Cpu = TypeArm | cpuArch64_32
	TypePpc      macho.Cpu = 18
	TypePpc64    macho.Cpu = TypePpc | 64
	// skip
)

const MaskSubType uint32 = 0xff000000

const (
	SubTypeX86All    uint32 = 3
	SubTypeX86_64All uint32 = 3
	SubTypeX86Arch1  uint32 = 4
	SubTypeX86_64H   uint32 = 8
)

const (
	SubTypeArmAll uint32 = 0
	SubTypeArmV4T uint32 = 5
	SubTypeArmV6  uint32 = 6
	// skip
	SubTypeArmV7  uint32 = 9
	SubTypeArmV7F uint32 = 10
	SubTypeArmV7S uint32 = 11
	SubTypeArmV7K uint32 = 12
	// skip
	SubTypeArmV6M  uint32 = 14
	SubTypeArmV7M  uint32 = 15
	SubTypeArmV7EM uint32 = 16
	SubTypeArmV8M  uint32 = 17
)

const (
	SubTypeArm64_32All uint32 = 0
)

const (
	SubTypeArm64All uint32 = 0
	SubTypeArm64V8  uint32 = 1
	SubTypeArm64E   uint32 = 2
)
