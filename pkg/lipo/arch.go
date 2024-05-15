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
	c      func() error
}

func (ff *FatFile) Close() error {
	return ff.c()
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
	c            func() error
	name         string
	updatedAlign uint32
}

func (a *arch) Name() string {
	return a.name
}

func (a *arch) Close() error {
	return a.c()
}

func (a *arch) Align() uint32 {
	return a.updatedAlign
}

func (a *arch) UpdateAlign(alignBit uint32) {
	a.updatedAlign = alignBit
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
		}
	}

	return &FatFile{
		Arches:    arches,
		FatHeader: ff.FatHeader,
		c:         f.Close,
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
			return nil, err
		}

		if input.Arch != "" {
			if obj.CPUString() != input.Arch {
				return nil, fmt.Errorf("specified architecture: %s for input file: %s does not match the file's architecture", input.Arch, input.Bin)
			}
		}

		arches[i] = &arch{
			Object:       obj,
			c:            f.Close,
			name:         input.Bin,
			updatedAlign: obj.Align(),
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

func OpenArchiveArches(p string) ([]Arch, error) {
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

		// TODO return "not allowed in an archive" error if it is fat file
		m, err := lmacho.NewArch(f.SectionReader)
		if err != nil {
			return nil, &lmacho.FormatError{Err: fmt.Errorf("not macho file: %w", err)}
		}

		arches = append(arches, &arch{
			Object:       m,
			name:         f.Name,
			updatedAlign: m.Align(),
			c:            ra.Close,
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

	return arches, nil
}
