package lipo

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

func extract[T lmacho.Object](objects []T, cpuStrings ...string) []T {
	m := util.ExistsMap(cpuStrings, func(v string) string {
		return v
	})
	return util.Filter(objects, func(o T) bool {
		_, ok := m[o.CPUString()]
		return ok
	})
}

func extractFamily[T lmacho.Object](objects []T, cpuStrings ...string) []T {
	m := util.ExistsMap(cpuStrings, func(v string) lmacho.Cpu {
		cpu, _, _ := lmacho.ToCpu(v)
		return cpu
	})
	return util.Filter(objects, func(o T) bool {
		_, ok := m[o.CPU()]
		return ok
	})
}

func remove[T lmacho.Object](objects []T, cpuStrings ...string) []T {
	m := util.ExistsMap(cpuStrings, func(v string) string {
		return v
	})
	return util.Filter(objects, func(o T) bool {
		_, ok := m[o.CPUString()]
		return !ok
	})
}

func replace[T lmacho.Object](objects []T, with []T) []T {
	cpuStrings := cpuStrings(with)
	new := remove(objects, cpuStrings...)
	return append(new, with...)
}

func cpuStrings[T lmacho.Object](objects []T) (cpuStrings []string) {
	ret := make([]string, len(objects))
	for i := 0; i < len(objects); i++ {
		ret[i] = objects[i].CPUString()
	}
	return ret
}

func contains[T lmacho.Object](objects []T, in ...T) bool {
	cpuStrings := cpuStrings(in)
	return len(extract(objects, cpuStrings...)) == len(in)
}

func updateAlignBit(arches []Arch, segAligns []*SegAlignInput) error {
	if len(segAligns) == 0 {
		return nil
	}

	dup := util.Duplicates(segAligns, func(k *SegAlignInput) string { return k.Arch })
	if dup != nil {
		return fmt.Errorf("segalign %s specified multiple times", *dup)
	}

	// make a map to lookup a fatArch early
	fam := make(map[string]Arch)
	for i := range arches {
		fam[arches[i].CPUString()] = arches[i]
	}

	for _, a := range segAligns {
		origHex := a.AlignHex
		if strings.HasPrefix(a.AlignHex, "0x") || strings.HasPrefix(a.AlignHex, "0X") {
			a.AlignHex = a.AlignHex[2:]
		}
		align, err := strconv.ParseInt(a.AlignHex, 16, 64)
		if err != nil {
			return fmt.Errorf("segalign %s not a proper hexadecimal number: %w", origHex, err)
		}

		if align == 0 || (align != 1 && (align%2) != 0) {
			return fmt.Errorf("segalign %s (hex) must be a non-zero power of two", a.AlignHex)
		}

		// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#LL74C42-L74C47
		maxSectAlign, _ := strconv.ParseInt("8000", 16, 64) // 0x8000 =  2^15
		if align > maxSectAlign {
			return fmt.Errorf("segalign %s (hex) must equal to or less than %x (hex)", a.AlignHex, maxSectAlign)
		}

		arch, found := fam[a.Arch]
		if !found {
			return fmt.Errorf("segalign %s specified but resulting fat file does not contain that architecture", a.Arch)
		}

		// update align bit
		alignBit := uint32(math.Log2(float64(align)))
		arch.UpdateAlign(alignBit)
	}

	return nil
}

type inspectType int

const (
	inspectFat inspectType = iota + 1
	inspectThin
	inspectArchive
	inspectUnknown
)

// inspect return object if the file is ar or macho(thin) object
func inspect(p string) (Arch, inspectType, error) {
	// handle general errors
	f, err := os.Open(p)
	if err != nil {
		return nil, inspectUnknown, err
	}
	defer f.Close()

	inspectedErrs := []error{}
	ff, err := OpenFatFile(p)
	if err == nil {
		defer ff.Close()
		return nil, inspectFat, nil
	}

	if errors.Is(err, lmacho.ErrThin) {
		a, err := OpenArches([]*ArchInput{{Bin: p}})
		if err != nil {
			return nil, inspectUnknown, err // unexpected error
		}
		defer close(a...)
		return a[0], inspectThin, nil
	}

	inspectedErrs = append(inspectedErrs, err)

	objs, err := OpenArchiveArches(p)
	if err == nil {
		defer close(objs...)
		return objs[0], inspectArchive, nil
	}
	if strings.HasPrefix(err.Error(), "archive member") {
		return nil, inspectUnknown, err
	}

	inspectedErrs = append(inspectedErrs, err)

	return nil, inspectUnknown, errors.Join(fmt.Errorf("can't figure out the architecture type of: %s", p), errors.Join(inspectedErrs...))
}
