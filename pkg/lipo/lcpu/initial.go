package lcpu

func init() {
	for i := range cpus {
		cpuSet[cpus[i].name] = &cpus[i]
		id := id(cpus[i].cpuType, cpus[i].subType)
		cpuIDSet[id] = &cpus[i]
	}
}

func id(c CPUType, s SubCPUType) uint64 {
	s &= ^MaskSubCPUType
	return (uint64(c) << 32) | uint64(s)
}

var (
	cpuSet   = map[string]*cpu{}
	cpuIDSet = map[uint64]*cpu{}
	cpus     = []cpu{
		{cpuType: CPUTypeI386, subType: SubCPUTypeX86All, name: "i386"},
		{cpuType: CPUTypeX86_64, subType: SubCPUTypeX86_64All, name: "x86_64"},
		{cpuType: CPUTypeX86_64, subType: SubCPUTypeX86_64H, name: "x86_64h"},
		// arm
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmAll, name: "arm"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV4T, name: "armv4t"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV6, name: "armv6"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7, name: "armv7"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7F, name: "armv7f"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7S, name: "armv7s"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7K, name: "armv7k"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV6M, name: "armv6m"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7M, name: "armv7m"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV7EM, name: "armv7em"},
		{cpuType: CPUTypeArm, subType: SubCPUTypeArmV8M, name: "armv8m"},
		// arm64
		{cpuType: CPUTypeArm64, subType: SubCPUTypeArm64All, name: "arm64"},
		{cpuType: CPUTypeArm64, subType: SubCPUTypeArm64E, name: "arm64e"},
		{cpuType: CPUTypeArm64, subType: SubCPUTypeArm64V8, name: "arm64v8"},
		// arm64_32
		{cpuType: CPUTypeArm64_32, subType: SubCPUTypeArm64_32V8, name: "arm64_32"},
	}
)
