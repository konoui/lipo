package lmacho

import (
	"debug/macho"
	"fmt"
	"sort"
)

const (
	AlignBitMax uint32 = 15
	AlignBitMin uint32 = 5
)

func SegmentAlignBit(f *macho.File) uint32 {
	cur := AlignBitMax
	for _, l := range f.Loads {
		if s, ok := l.(*macho.Segment); ok {
			align := GuessAlignBit(s.Addr, AlignBitMin, AlignBitMax)
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

func SortBy(arches []FatArch) ([]FatArch, error) {
	SortFunc(arches, func(i, j int) bool {
		return compare(arches[i], arches[j])
	})

	// update offset
	offset := int64(fatHeaderSize + fatArchHeaderSize*uint32(len(arches)))
	for i := range arches {
		offset = align(int64(offset), 1<<int64(arches[i].Align))
		if !boundaryOK(offset) {
			return nil, fmt.Errorf("exceeds maximum fat32 size")
		}
		arches[i].Offset = uint32(offset)
		offset += int64(arches[i].Size)
	}

	return arches, nil
}

func align(offset, v int64) int64 {
	return (offset + v - 1) / v * v
}

func boundaryOK(s int64) (ok bool) {
	return s < 1<<32
}