package lipo

import (
	"debug/macho"
	"errors"
	"fmt"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

const (
	alignBitMax uint32 = 15
	alignBitMin uint32 = 5
)

const (
	noMatchFmt         = "%s specified but fat file: %s does not contain that architecture"
	unsupportedArchFmt = "unsupported architecture: %s"
)

var (
	errNoInput = errors.New("no input files specified")
)

type Lipo struct {
	in        []string
	out       string
	segAligns []*SegAlignInput
	arches    []*ArchInput
}

type SegAlignInput struct {
	Arch     string
	AlignHex string
}

type ArchInput struct {
	Arch string
	Bin  string
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

func WithArch(arches []*ArchInput) Option {
	return func(l *Lipo) {
		l.arches = arches
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

func (l *Lipo) validateOneInput() error {
	num := len(l.in)
	if num == 0 {
		return errNoInput
	} else if num != 1 {
		return fmt.Errorf("only one input file can be specified")
	}
	return nil
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

func validateInputArches(arches []string) error {
	dup := duplicates(arches)
	if dup != "" {
		return fmt.Errorf("architecture %s specified multiple times", dup)
	}

	for _, arch := range arches {
		if !mcpu.IsSupported(arch) {
			return fmt.Errorf("unsupported architecture %s", arch)
		}
	}
	return nil
}

func contains(tg string, l []string) bool {
	for _, s := range l {
		if tg == s {
			return true
		}
	}
	return false
}

func duplicates(l []string) string {
	seen := map[string]bool{}
	for _, v := range l {
		if o, k := seen[v]; o || k {
			return v
		}
		seen[v] = true
	}
	return ""
}

// remove return values `a` does not have
func remove(a []string, b []string) string {
	for _, v := range b {
		if !contains(v, a) {
			return v
		}
	}
	return ""
}
