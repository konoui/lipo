package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/konoui/lipo/pkg/ar"
	"github.com/konoui/lipo/pkg/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

type FatFile struct {
	lmacho.FatHeader
	Arches []Arch
	io.Closer
}

type Archive struct {
	io.Closer
	Arches       []Arch
	size         uint64
	name         string
	updatedAlign uint32
	sr           *io.SectionReader
}

func (a *Archive) CPUString() string {
	return a.Arches[0].CPUString()
}

func (a *Archive) CPU() lmacho.Cpu {
	return a.Arches[0].CPU()
}

func (a *Archive) SubCPU() lmacho.SubCpu {
	return a.Arches[0].SubCPU()
}

// https://github.com/apple-oss-distributions/cctools/blob/cctools-1021.4/misc/lipo.c#L1474-L1496
func (a *Archive) Align() uint32 {
	if a.CPU()&lmacho.CPUArch64 > 0 {
		return lmacho.AlignBitMin64
	}
	return lmacho.AlignBitMin32
}

func (a *Archive) Size() uint64 {
	return a.size
}

func (a *Archive) Type() macho.Type {
	// FIXME
	return 9999
}

func (a *Archive) Name() string {
	return a.name
}

func (a *Archive) UpdateAlign(alignBit uint32) {
	a.updatedAlign = alignBit
}

func (a *Archive) Read(b []byte) (int, error) {
	return a.sr.Read(b)
}

func (a *Archive) ReadAt(b []byte, off int64) (int, error) {
	return a.sr.ReadAt(b, off)
}

type Arch interface {
	lmacho.Object
	io.Closer
	Name() string
	UpdateAlign(uint32)
}

var _ Arch = &arch{}

type arch struct {
	lmacho.Object
	io.Closer
	name         string
	updatedAlign uint32
}

func (a *arch) Name() string {
	return a.name
}

func (a *arch) Align() uint32 {
	return a.updatedAlign
}

func (a *arch) UpdateAlign(alignBit uint32) {
	a.updatedAlign = alignBit
}

type nopCloser struct{}

func (*nopCloser) Close() error {
	return nil
}

func close[T io.Closer](arches ...T) {
	for _, a := range arches {
		a.Close()
	}
}

func OpenFatFile(p string) (*FatFile, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	ff, err := lmacho.NewFatFile(f)
	if err != nil {
		return nil, err
	}

	arches := make([]Arch, len(ff.Arches))
	for i := range ff.Arches {
		arches[i] = &arch{
			Object:       ff.Arches[i],
			name:         p,
			updatedAlign: ff.Arches[i].Align(),
			Closer:       &nopCloser{},
		}
	}

	return &FatFile{
		Arches:    arches,
		FatHeader: ff.FatHeader,
		Closer:    f,
	}, nil
}

func OpenArches(inputs []*ArchInput) ([]Arch, error) {
	arches := make([]Arch, 0, len(inputs))
	for _, input := range inputs {
		f, err := os.Open(input.Bin)
		if err != nil {
			return nil, err
		}
		stats, err := f.Stat()
		if err != nil {
			return nil, err
		}

		typ, err := inspect(input.Bin)
		if err != nil {
			return nil, err
		}

		switch typ {
		case inspectThin:
			sr := io.NewSectionReader(f, 0, stats.Size())
			obj, err := lmacho.NewArch(sr)
			if err != nil {
				fe := &lmacho.FormatError{}
				if errors.As(err, &fe) {
					fmt.Println("aaa", err)
					return nil, fmt.Errorf("can't figure out the architecture type of: %s", input.Bin)
				}
				return nil, err
			}
			if input.Arch != "" {
				if obj.CPUString() != input.Arch {
					return nil, fmt.Errorf("specified architecture: %s for input file: %s does not match the file's architecture", input.Arch, input.Bin)
				}
			}
			arches = append(arches, &arch{
				Object:       obj,
				name:         input.Bin,
				updatedAlign: obj.Align(),
				Closer:       f,
			})
		case inspectArchive:
			archive, err := OpenArchive(input.Bin)
			if err != nil {
				return nil, err
			}
			arches = append(arches, archive)
		case inspectFat:
			fat, err := OpenFatFile(input.Bin)
			if err != nil {
				return nil, err
			}
			arches = append(arches, fat.Arches...)
		default:
			return nil, fmt.Errorf("can't figure out the architecture type of: %s", input.Bin)
		}

	}

	dup := util.Duplicates(arches, func(a Arch) string {
		return a.CPUString()
	})
	if dup != nil {
		return nil, fmt.Errorf("the inputs have the same architectures (%s)", *dup)
	}

	return arches, nil
}

func OpenArchive(p string) (*Archive, error) {
	ra, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	info, err := ra.Stat()
	if err != nil {
		return nil, err
	}

	size := info.Size()

	files, err := ar.NewArchive(ra)
	if err != nil {
		return nil, err
	}

	arches := make([]Arch, 0, len(files))
	for _, f := range files {
		if strings.HasPrefix(f.Name, ar.PrefixSymdef) {
			continue
		}

		m, err := lmacho.NewArch(f.SectionReader)
		if err != nil {
			typ, _ := inspect(p)
			if typ == inspectFat {
				return nil, &lmacho.FormatError{Err: fmt.Errorf("archive member %s(%s) is a fat file (not allowed in an archive", ra.Name(), f.Name)}
			}
			return nil, &lmacho.FormatError{Err: fmt.Errorf("archive member %s(%s) is not macho file: %w", ra.Name(), f.Name, err)}
		}

		arches = append(arches, &arch{
			Object:       m,
			name:         f.Name,
			updatedAlign: m.Align(),
			Closer:       &nopCloser{},
		})
	}

	if len(arches) == 0 {
		return nil, &lmacho.FormatError{Err: errors.New(("no object in the archive"))}
	}

	cpuString := arches[0].CPUString()
	for _, arch := range arches {
		if cpuString != arch.CPUString() {
			return nil, fmt.Errorf("archive member %s(%s) cputype (%d) and cpusubtype (%d) does not match previous archive members cputype (%d) and cpusubtype (%d) (all members must match)", p, arch.Name(), arch.CPU(), arch.SubCPU(), arches[0].CPU(), arches[0].SubCPU())
		}
	}

	sr := io.NewSectionReader(ra, 0, size)

	return &Archive{
		Arches: arches,
		Closer: ra,
		size:   uint64(size),
		name:   p,
		sr:     sr,
	}, nil
}
