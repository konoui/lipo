package lipo

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"sync/atomic"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
	"github.com/konoui/lipo/pkg/util"
)

var _ io.ReadCloser = &fatArch{}

// fatArch consist of FatArchHeader and io.Reader for binary
type fatArch struct {
	macho.FatArchHeader
	r io.Reader
	c io.Closer
	// TODO check work fine
	count *int32
}

func (fa *fatArch) Read(p []byte) (int, error) {
	return fa.r.Read(p)
}

func (fa *fatArch) Close() error {
	if fa == nil || fa.c == nil {
		return nil
	}

	if fa.count != nil && *fa.count > 0 {
		atomic.AddInt32(fa.count, -1)
		return nil
	}

	err := fa.c.Close()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}
	return err
}

type fatArches []*fatArch

func (f fatArches) close() error {
	msg := ""
	for _, closer := range f {
		err := closer.Close()
		if err != nil {
			msg += err.Error()
		}
	}
	if msg != "" {
		return fmt.Errorf("close errors: %s", msg)
	}
	return nil
}

func (f fatArches) createFatBinary(p string, perm os.FileMode) (err error) {
	f, err = f.sort()
	if err != nil {
		return err
	}

	if len(f) == 0 {
		return errors.New("error empty fat file due to no inputs")
	}
	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := out.Chmod(perm); cerr != nil && err == nil {
			err = cerr
			return
		}
		if ferr := out.Close(); ferr != nil && err == nil {
			err = ferr
			return
		}
	}()

	return f.outputFatBinary(out)
}

func (f fatArches) outputFatBinary(out io.Writer) error {
	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(f)),
	}

	// sort by offset by asc for effective writing binary data
	sort.Slice(f, func(i, j int) bool {
		return f[i].Offset < f[j].Offset
	})

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("error write fat header: %w", err)
	}

	// write headers
	for _, hdr := range f {
		if err := binary.Write(out, binary.BigEndian, hdr.FatArchHeader); err != nil {
			return fmt.Errorf("error write arch headers: %w", err)
		}
	}

	off := fatHeader.size()
	for _, fatArch := range f {
		if off < fatArch.Offset {
			// write empty data for alignment
			empty := make([]byte, fatArch.Offset-off)
			if _, err := out.Write(empty); err != nil {
				return fmt.Errorf("error alignment: %w", err)
			}
			off = fatArch.Offset
		}

		// write binary data
		if _, err := io.CopyN(out, fatArch.r, int64(fatArch.Size)); err != nil {
			return fmt.Errorf("error write binary data: %w", err)
		}
		off += fatArch.Size
	}

	return nil
}

// Note mock using qsort
var SortFunc = sort.Slice

// https://github.com/apple-oss-distributions/cctools/blob/cctools-973.0.1/misc/lipo.c#L2677
func compare(i, j *fatArch) bool {
	if i.Cpu == j.Cpu {
		return (i.SubCpu & ^mcpu.MaskSubType) < (j.SubCpu & ^mcpu.MaskSubType)
	}

	if i.Cpu == mcpu.TypeArm64 {
		return false
	}
	if j.Cpu == mcpu.TypeArm64 {
		return true
	}

	return i.Align < j.Align
}

func (f fatArches) sort() (fatArches, error) {
	SortFunc(f, func(i, j int) bool {
		return compare(f[i], f[j])
	})

	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(f)),
	}

	// update offset
	offset := int64(fatHeader.size())
	for i := range f {
		offset = align(int64(offset), 1<<int64(f[i].Align))
		if !boundaryOK(offset) {
			return nil, fmt.Errorf("exceeds maximum fat32 size")
		}
		f[i].Offset = uint32(offset)
		offset += int64(f[i].Size)
	}

	return f, nil
}

func (f fatArches) extract(arches ...string) fatArches {
	return util.Filter(f, func(v *fatArch) bool {
		return contains(mcpu.ToString(v.Cpu, v.SubCpu), arches)
	})
}

func (f fatArches) remove(arches ...string) fatArches {
	return util.Filter(f, func(v *fatArch) bool {
		return !contains(mcpu.ToString(v.Cpu, v.SubCpu), arches)
	})
}

func (f fatArches) contains(in fatArches) bool {
	arches := util.Map(in, func(v *fatArch) string {
		return mcpu.ToString(v.Cpu, v.SubCpu)
	})
	return len(f.extract(arches...)) == len(in)
}

func (f fatArches) replace(with fatArches) fatArches {
	arches := util.Map(with, func(v *fatArch) string {
		return mcpu.ToString(v.Cpu, v.SubCpu)
	})
	new := f.remove(arches...)
	return append(new, with...)
}

func (f fatArches) arches() []string {
	return util.Map(f, func(v *fatArch) string {
		return mcpu.ToString(v.Cpu, v.SubCpu)
	})
}

func (f fatArches) updateAlignBit(segAligns []*SegAlignInput) error {
	if len(segAligns) == 0 {
		return nil
	}

	seen := map[string]bool{}
	for _, a := range segAligns {
		align, err := strconv.ParseInt(a.AlignHex, 16, 64)
		if err != nil {
			return err
		}
		if (align % 2) != 0 {
			return fmt.Errorf("argument to -segalign <arch_type> %s (hex) must be a non-zero power of two", a.AlignHex)
		}

		if o, k := seen[a.Arch]; o || k {
			return fmt.Errorf("-segalign %s <value> specified multiple times", a.Arch)
		}
		seen[a.Arch] = true

		alignBit := uint32(math.Log2(float64(align)))
		found := false
		for idx := range f {
			if mcpu.ToString(f[idx].Cpu, f[idx].SubCpu) == a.Arch {
				f[idx].Align = alignBit
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("-segalign <arch_type> %s not found", a.Arch)
		}
	}

	_, err := f.sort()
	return err
}

// fatArchesFromFatBin gathers fatArches from fat binary header if `cond` returns true
func fatArchesFromFatBin(path string) (fatArches, error) {
	fat, err := macho.OpenFat(path)
	if err != nil {
		return nil, err
	}
	defer fat.Close()

	if len(fat.Arches) < 1 {
		return nil, errors.New("number of arches must be greater than 1")
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	fatArches := fatArches{}
	count := int32(0)
	for _, hdr := range fat.Arches {
		fatArches = append(fatArches, &fatArch{
			FatArchHeader: hdr.FatArchHeader,
			r:             io.NewSectionReader(f, int64(hdr.Offset), int64(hdr.Size)),
			c:             f,
			count:         &count,
		})
		count++
	}

	return fatArches, nil
}

func fatArchesFromCreateInputs(inputs []*createInput) (fatArches, error) {
	fatHdr := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(inputs)),
	}

	fatArches := make(fatArches, 0, len(inputs))

	offset := int64(fatHdr.size())
	for _, in := range inputs {
		offset = align(offset, 1<<in.align)

		// validate addressing boundary since size and offset of fat32 are uint32
		if !(boundaryOK(offset) && boundaryOK(in.size)) {
			return nil, fmt.Errorf("exceeds maximum fat32 size at %s", in.path)
		}

		hdrOffset := uint32(offset)
		hdrSize := uint32(in.size)
		hdr := macho.FatArchHeader{
			Cpu:    in.hdr.Cpu,
			SubCpu: in.hdr.SubCpu,
			Offset: hdrOffset,
			Size:   hdrSize,
			Align:  in.align,
		}

		offset += int64(hdr.Size)

		f, err := os.Open(in.path)
		if err != nil {
			return nil, err
		}
		fatArches = append(fatArches, &fatArch{
			FatArchHeader: hdr,
			r:             f,
			c:             f,
		})
	}

	return fatArches, nil
}
