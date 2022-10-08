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

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

const (
	alignBitMax uint32 = 15
	alignBitMin uint32 = 5
)

const (
	noMatchFmt = "%s <arch_file> specified but fat file: %s does not contain that architecture"
)

type Lipo struct {
	in        []string
	out       string
	segAligns []*SegAlignInput
}

type SegAlignInput struct {
	Arch     string
	AlignHex string
}

type Option func(l *Lipo)

func WithInputs(in ...string) Option {
	return func(l *Lipo) {
		l.in = in
	}
}

func WithOutput(out string) Option {
	return func(l *Lipo) {
		l.out = out
	}
}

func WithSegAlign(aligns []*SegAlignInput) Option {
	return func(l *Lipo) {
		l.segAligns = aligns
	}
}

func New(opts ...Option) *Lipo {
	l := &Lipo{}
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		opt(l)
	}
	return l
}

// see /Library/Developer/CommandLineTools/SDKs/MacOSX.sdk/usr/include/mach-o/fat.h
type fatHeader struct {
	magic uint32
	narch uint32
}

func (h *fatHeader) size() uint32 {
	// sizeof(fatHeader) = uint32 * 2
	sizeofFatHdr := uint32(4 * 2)
	// sizeof(macho.FatArchHeader) = uint32 * 5
	sizeofFatArchHdr := uint32(4 * 5)
	size := sizeofFatHdr + sizeofFatArchHdr*h.narch
	return size
}

var _ io.ReadCloser = &fatArch{}

// fatArch consist of FatArchHeader and io.Reader for binary
type fatArch struct {
	macho.FatArchHeader
	r io.Reader
	c io.Closer
}

func (fa *fatArch) Read(p []byte) (int, error) {
	return fa.r.Read(p)
}

func (fa *fatArch) Close() error {
	if fa == nil || fa.c == nil {
		return nil
	}

	err := fa.c.Close()
	if errors.Is(err, os.ErrClosed) {
		return nil
	}
	return err
}

func close(closers []*fatArch) error {
	msg := ""
	for _, closer := range closers {
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

func sortByArch(fatArches []*fatArch) ([]*fatArch, error) {
	SortFunc(fatArches, func(i, j int) bool {
		return compare(fatArches[i], fatArches[j])
	})

	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(fatArches)),
	}

	// update offset
	offset := int64(fatHeader.size())
	for i := range fatArches {
		offset = align(int64(offset), 1<<int64(fatArches[i].Align))
		if !boundaryOK(offset) {
			return nil, fmt.Errorf("exceeds maximum fat32 size")
		}
		fatArches[i].Offset = uint32(offset)
		offset += int64(fatArches[i].Size)
	}

	return fatArches, nil
}

func updateAlignBit(fatArches []*fatArch, segAligns []*SegAlignInput) error {
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
		for idx := range fatArches {
			if mcpu.ToString(fatArches[idx].Cpu, fatArches[idx].SubCpu) == a.Arch {
				fatArches[idx].Align = alignBit
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("-segalign <arch_type> %s not found", a.Arch)
		}
	}

	_, err := sortByArch(fatArches)
	return err
}

func segmentAlignBit(f *macho.File) uint32 {
	cur := alignBitMax
	for _, l := range f.Loads {
		if s, ok := l.(*macho.Segment); ok {
			align := guessAlignBit(s.Addr, alignBitMin, alignBitMax)
			if align < cur {
				cur = align
			}
		}
	}
	return cur
}

func guessAlignBit(addr uint64, min, max uint32) uint32 {
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

func align(offset, v int64) int64 {
	return (offset + v - 1) / v * v
}

func boundaryOK(s int64) (ok bool) {
	return s < 1<<32
}

func outputFatBinary(p string, perm os.FileMode, fatArches []*fatArch) (err error) {
	if len(fatArches) == 0 {
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

	return createFatBinary(out, fatArches)
}

func createFatBinary(out io.Writer, fatArches []*fatArch) error {
	fatHeader := &fatHeader{
		magic: macho.MagicFat,
		narch: uint32(len(fatArches)),
	}

	// sort by offset by asc for effective writing binary data
	sort.Slice(fatArches, func(i, j int) bool {
		return fatArches[i].Offset < fatArches[j].Offset
	})

	// write header
	// see https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/debug/macho/fat.go;l=45
	if err := binary.Write(out, binary.BigEndian, fatHeader); err != nil {
		return fmt.Errorf("error write fat header: %w", err)
	}

	// write headers
	for _, hdr := range fatArches {
		if err := binary.Write(out, binary.BigEndian, hdr.FatArchHeader); err != nil {
			return fmt.Errorf("error write arch headers: %w", err)
		}
	}

	off := fatHeader.size()
	for _, fatArch := range fatArches {
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

func contain(tg string, l []string) bool {
	for _, s := range l {
		if tg == s {
			return true
		}
	}
	return false
}
