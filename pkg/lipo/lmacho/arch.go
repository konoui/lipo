package lmacho

import (
	"debug/macho"
	"fmt"
	"sort"
)

const (
	alignBitMax uint32 = 15
	alignBitMin uint32 = 5
)

func SegmentAlignBit(f *macho.File) uint32 {
	cur := alignBitMax
	for _, l := range f.Loads {
		if s, ok := l.(*macho.Segment); ok {
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
		segAlign = segAlign << 1
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

// Note mock using qsort
var SortFunc = sort.Slice

// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#L2677
func compare(i, j FatArch) bool {
	if i.Cpu == j.Cpu {
		return (i.SubCpu & ^MaskSubCpuType) < (j.SubCpu & ^MaskSubCpuType)
	}

	if i.Cpu == CpuTypeArm64 {
		return false
	}
	if j.Cpu == CpuTypeArm64 {
		return true
	}

	return i.Align < j.Align
}

func (f *FatFile) allSortedArches() ([]FatArch, error) {
	arches := f.AllArches()
	SortFunc(arches, func(i, j int) bool {
		return compare(arches[i], arches[j])
	})

	// update offset
	offset := uint64(f.fatHeaderSize() + f.fatArchHeaderSize()*uint64(len(arches)))
	for i := range arches {
		offset = align(uint64(offset), 1<<arches[i].Align)
		if f.Magic == macho.MagicFat && !boundaryOK(offset) {
			return nil, fmt.Errorf("exceeds maximum 32 bit size")
		}
		arches[i].Offset = offset
		offset += arches[i].Size
	}

	return arches, nil
}

func align(offset, v uint64) uint64 {
	return (offset + v - 1) / v * v
}

func boundaryOK(s uint64) (ok bool) {
	return s < 1<<32
}
