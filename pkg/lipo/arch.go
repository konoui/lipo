package lipo

import (
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
	Arches []Arch
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
	arches := make([]Arch, len(inputs))
	for i, input := range inputs {
		f, err := os.Open(input.Bin)
		if err != nil {
			return nil, err
		}
		stats, err := f.Stat()
		if err != nil {
			return nil, err
		}

		sr := io.NewSectionReader(f, 0, stats.Size())
		obj, err := lmacho.NewArch(sr)
		if err != nil {
			fe := &lmacho.FormatError{}
			if errors.As(err, &fe) {
				return nil, fmt.Errorf("can't figure out the architecture type of: %s", input.Bin)
			}
			return nil, err
		}

		if input.Arch != "" {
			if obj.CPUString() != input.Arch {
				return nil, fmt.Errorf("specified architecture: %s for input file: %s does not match the file's architecture", input.Arch, input.Bin)
			}
		}

		arches[i] = &arch{
			Object:       obj,
			name:         input.Bin,
			updatedAlign: obj.Align(),
			Closer:       f,
		}
	}

	dup := util.Duplicates(arches, func(a Arch) string {
		return a.CPUString()
	})
	if dup != nil {
		return nil, fmt.Errorf("duplicate architecture: %s", *dup)
	}

	return arches, nil
}

func OpenArchiveArches(p string) (*Archive, error) {
	ra, err := os.Open(p)
	if err != nil {
		return nil, err
	}

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

	return &Archive{
		Arches: arches,
		Closer: ra,
	}, nil
}
