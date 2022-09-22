package lipo

import (
	"debug/macho"
	"fmt"
)

func isSupportedArch(v string) bool {
	for _, cs := range cpuNames {
		if cs.v == v {
			return true
		}
	}
	return false
}

func cpuString(cpu macho.Cpu, subCpu uint32) string {
	maskedSub := (subCpu & ^CpuSubMask)
	for _, cs := range cpuNames {
		if cs.t == uint32(cpu) && cs.s == maskedSub {
			return cs.v
		}
	}

	unknown := fmt.Sprintf("unknown cpu: %d subCpu: %d", cpu, maskedSub)
	return unknown
}

type cpuName struct {
	t uint32
	s uint32
	v string
}

var cpuNames = []cpuName{
	{t: uint32(CpuX86), s: CpuSubX86All, v: "i386"},
	{t: uint32(CpuI386), s: CpuSubX86All, v: "i386"},
	{t: uint32(CpuX86_64), s: CpuSubX86_64All, v: "x86_64"},
	{t: uint32(CpuX86_64), s: CpuSubX86_64H, v: "x86_64h"},
	// arm
	{t: uint32(CpuArm), s: CPUSubArmAll, v: "arm"},
	{t: uint32(CpuArm), s: CpuSubArmV4T, v: "armv4t"},
	{t: uint32(CpuArm), s: CpuSubArmV6, v: "armv6"},
	{t: uint32(CpuArm), s: CpuSubArmV7, v: "armv7"},
	{t: uint32(CpuArm), s: CpuSubArmV7F, v: "armv7f"},
	{t: uint32(CpuArm), s: CpuSubArmV7S, v: "armv7s"},
	{t: uint32(CpuArm), s: CpuSubArmV7K, v: "armv7k"},
	{t: uint32(CpuArm), s: CpuSubArmV6M, v: "armv6m"},
	{t: uint32(CpuArm), s: CpuSubArmV7M, v: "armv7m"},
	{t: uint32(CpuArm), s: CpuSubArmV7EM, v: "armv7em"},
	{t: uint32(CpuArm), s: CpuSubArmV8M, v: "armv8m"},
	// arm64
	{t: uint32(CpuArm64), s: CpuSubArm64All, v: "arm64"},
	{t: uint32(CpuArm64), s: CpuSubArm64E, v: "arm64e"},
	{t: uint32(CpuArm64), s: CpuSubArm64V8, v: "arm64v8"},
	// arm64_32
	{t: uint32(CpuArm64_32), s: CpuSubArm64_32All, v: "arm64_32"},
}

const cpuArch64 = 0x01000000
const cpuArch64_32 = 0x02000000

const (
	// skip
	CpuX86    macho.Cpu = 7
	CpuI386   macho.Cpu = CpuX86
	CpuX86_64 macho.Cpu = CpuI386 | cpuArch64
	// skip
	CpuArm      macho.Cpu = 12
	CpuArm64    macho.Cpu = CpuArm | cpuArch64
	CpuArm64_32 macho.Cpu = CpuArm | cpuArch64_32
	CpuPpc      macho.Cpu = 18
	CpuPpc64    macho.Cpu = CpuPpc | 64
	// skip
)

const CpuSubMask uint32 = 0xff000000

const (
	CpuSubX86All    uint32 = 3
	CpuSubX86_64All uint32 = 3
	CpuSubX86Arch1  uint32 = 4
	CpuSubX86_64H   uint32 = 8
)

const (
	CPUSubArmAll uint32 = 0
	CpuSubArmV4T uint32 = 5
	CpuSubArmV6  uint32 = 6
	// skip
	CpuSubArmV7  uint32 = 9
	CpuSubArmV7F uint32 = 10
	CpuSubArmV7S uint32 = 11
	CpuSubArmV7K uint32 = 12
	// skip
	CpuSubArmV6M  uint32 = 14
	CpuSubArmV7M  uint32 = 15
	CpuSubArmV7EM uint32 = 16
	CpuSubArmV8M  uint32 = 17
)

const (
	CpuSubArm64_32All uint32 = 0
)

const (
	CpuSubArm64All uint32 = 0
	CpuSubArm64V8  uint32 = 1
	CpuSubArm64E   uint32 = 2
)
