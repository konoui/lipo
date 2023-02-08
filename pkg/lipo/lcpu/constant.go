package lcpu

// /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach/machine.h
const cpuArch64 = 0x01000000
const cpuArch64_32 = 0x02000000

const (
	// skip
	CPUTypeX86    CPUType = 7
	CPUTypeI386   CPUType = CPUTypeX86
	CPUTypeX86_64 CPUType = CPUTypeI386 | cpuArch64
	// skip
	CPUTypeArm      CPUType = 12
	CPUTypeArm64    CPUType = CPUTypeArm | cpuArch64
	CPUTypeArm64_32 CPUType = CPUTypeArm | cpuArch64_32
	CPUTypePpc      CPUType = 18
	CPUTypePpc64    CPUType = CPUTypePpc | 64
	// skip
)

const MaskSubCPUType SubCPUType = 0xff000000

const (
	SubCPUTypeX86All    SubCPUType = 3
	SubCPUTypeX86_64All SubCPUType = 3
	SubCPUTypeX86Arch1  SubCPUType = 4
	SubCPUTypeX86_64H   SubCPUType = 8
)

const (
	SubCPUTypeArmAll SubCPUType = 0
	SubCPUTypeArmV4T SubCPUType = 5
	SubCPUTypeArmV6  SubCPUType = 6
	// skip
	SubCPUTypeArmV7  SubCPUType = 9
	SubCPUTypeArmV7F SubCPUType = 10
	SubCPUTypeArmV7S SubCPUType = 11
	SubCPUTypeArmV7K SubCPUType = 12
	// skip
	SubCPUTypeArmV6M  SubCPUType = 14
	SubCPUTypeArmV7M  SubCPUType = 15
	SubCPUTypeArmV7EM SubCPUType = 16
	SubCPUTypeArmV8M  SubCPUType = 17
)

const (
	SubCPUTypeArm64_32V8 SubCPUType = 1
)

const (
	SubCPUTypeArm64All SubCPUType = 0
	SubCPUTypeArm64V8  SubCPUType = 1
	SubCPUTypeArm64E   SubCPUType = 2
)
