package lmacho

import (
	"debug/macho"
	"fmt"

	"github.com/konoui/go-qsort"
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
func CmpArchFunc[T Object](i, j T) int {
	if i.CPU() == j.CPU() {
		return int((i.SubCPU() & ^MaskSubCpuType)) - int((j.SubCPU() & ^MaskSubCpuType))
	}

	if i.CPU() == TypeArm64 {
		return 1
	}
	if j.CPU() == TypeArm64 {
		return -1
	}

	return int(i.Align()) - int(j.Align())
}

// sortArches sorts and update offset by `arches`
func sortAndUpdateArches(arches []*FatArch, magic uint32) error {
	qsort.Slice(arches, CmpArchFunc)

	// update offset
	offset := FatHeaderSize() + FatArchHeaderSize(magic)*uint64(len(arches))
	for i := range arches {
		offset = align(offset, 1<<arches[i].Align())
		arches[i].faHdr.Offset = offset
		offset += arches[i].Size()
		if magic == macho.MagicFat && !boundary32OK(offset) {
			return fmt.Errorf("exceeds maximum 32 bit size. please handle it as fat64")
		}
	}

	return nil
}

func hasDuplicatesErr[T Object](arches []T) error {
	seenArches := make(map[uint64]bool, len(arches))
	for _, a := range arches {
		seenArch := (uint64(a.CPU()) << 32) | uint64(a.SubCPU())
		if o, k := seenArches[seenArch]; o || k {
			return fmt.Errorf("duplicate architecture %s", a.CPUString())
		}
		seenArches[seenArch] = true
	}

	return nil
}

func validateHideARM64Objects[T Object](objects []T, hideARM64 bool) error {
	if !hideARM64 {
		return nil
	}
	for _, obj := range objects {
		if obj.Type() == macho.TypeObj {
			return fmt.Errorf("hideARM64 specified but type is not MH_EXECUTE")
		}
	}
	return nil
}

func checkMaxAlignBit[T Object](arches []T) error {
	for _, a := range arches {
		if a.Align() > alignBitMax {
			return fmt.Errorf("align (2^%d) too large of fat file (cputype (%d) cpusubtype (%d)) (maximum 2^%d)", a.Align(), a.CPU(), a.SubCPU()^MaskSubCpuType, alignBitMax)
		}

	}
	return nil
}

func align(offset, v uint64) uint64 {
	return (offset + v - 1) / v * v
}

func boundary32OK(s uint64) (ok bool) {
	return s < 1<<32
}
