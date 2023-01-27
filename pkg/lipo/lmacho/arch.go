package lmacho

import (
	"debug/macho"
)

const (
	alignBitMax   uint32 = 15
	alignBitMin32 uint32 = 2
	alignBitMin64 uint32 = 3
)

func SegmentAlignBit(f *macho.File) uint32 {
	cur := alignBitMax
	for _, l := range f.Loads {
		if s, ok := l.(*macho.Segment); ok {
			alignBitMin := alignBitMin64
			if s.Cmd == macho.LoadCmdSegment {
				alignBitMin = alignBitMin32
			}
			align := GuessAlignBit(s.Addr, alignBitMin, alignBitMax)
			if align < cur {
				cur = align
			}
		}
	}
	return cur
}

func GuessAlignBit(addr uint64, min, max uint32) uint32 {
	segAlign := uint64(1)
	align := uint32(0)
	if addr == 0 {
		return max
	}
	for {
		segAlign <<= 1
		align++
		if (segAlign & addr) != 0 {
			break
		}
	}

	if align < min {
		return min
	}
	if max < align {
		return max
	}
	return align
}

// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#L2677
func CmpArchFunc(i, j FatArch) int {
	if i.Cpu == j.Cpu {
		return int((i.SubCpu & ^MaskSubCpuType)) - int((j.SubCpu & ^MaskSubCpuType))
	}

	if i.Cpu == CpuTypeArm64 {
		return 1
	}
	if j.Cpu == CpuTypeArm64 {
		return -1
	}

	return int(i.Align) - int(j.Align)
}

func align(offset, v uint64) uint64 {
	return (offset + v - 1) / v * v
}

func boundaryOK(s uint64) (ok bool) {
	return s < 1<<32
}
