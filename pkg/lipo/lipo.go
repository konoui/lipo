package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/konoui/lipo/pkg/lipo/ar"
	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/util"
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
	hideArm64 bool
	fat64     bool
}

type SegAlignInput struct {
	Arch     string
	AlignHex string
}

type ArchInput struct {
	Arch string
	Bin  string
}

type ReplaceInput = ArchInput

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

func WithSegAlign(aligns ...*SegAlignInput) Option {
	return func(l *Lipo) {
		l.segAligns = aligns
	}
}

func WithArch(arches ...*ArchInput) Option {
	return func(l *Lipo) {
		l.arches = arches
	}
}

func WithHideArm64() Option {
	return func(l *Lipo) {
		l.hideArm64 = true
	}
}

func WithFat64() Option {
	return func(l *Lipo) {
		l.fat64 = true
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

func hideArmObjectErr(arches fatArches) error {
	for _, arch := range arches {
		if arch.FileHeader.Type == macho.TypeObj {
			return fmt.Errorf("hideARM64 specified but thin file %s is not of type MH_EXECUTE", arch.Name)
		}
	}
	return nil
}

type fileType int

const (
	typeThin = iota
	typeFat
	typeArchive
)

type inspectResult struct {
	arches   []string
	fileType fileType
}

func inspectFile(p string) (*inspectResult, error) {
	ff, err := lmacho.NewFatFile(p)
	if err == nil {
		all := ff.AllArches()
		cpus := make([]string, len(all))
		for i, hdr := range all {
			cpus[i] = lmacho.ToCpuString(hdr.Cpu, hdr.SubCpu)
		}
		return &inspectResult{
			arches:   cpus,
			fileType: typeFat,
		}, nil
	}

	fatErr := err

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	mf, err := macho.NewFile(f)
	if err == nil {
		return &inspectResult{
			arches:   []string{lmacho.ToCpuString(mf.Cpu, mf.SubCpu)},
			fileType: typeThin,
		}, nil
	}

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("cannot reset file pointer: %w", err)
	}

	isAr, mh, err := isArchive(f)
	if err != nil {
		return nil, fmt.Errorf("archive error: %w", err)
	}
	if isAr {
		return &inspectResult{
			arches:   []string{lmacho.ToCpuString(mh.Cpu, mh.SubCpu)},
			fileType: typeArchive,
		}, nil
	}

	return nil, fmt.Errorf("can't figure out the architecture type of: %s: %w", p, fatErr)
}

// isArchive the file it is invalid archive format if err is not nil
// [false, nil] means not archive format
// [true, nil] means valid archive format
// [false, err] means invalid archive format
func isArchive(r io.ReaderAt) (bool, *macho.FileHeader, error) {
	archiver, err := ar.New(r)
	if err != nil {
		if errors.Is(err, ar.ErrNotAr) {
			return false, nil, nil
		}
		return false, nil, err
	}

	makeErr := func(arName string, mh macho.FileHeader, prevCPU macho.Cpu, prevSubCPU lmacho.SubCpu) error {
		return fmt.Errorf("archive member %s cputype (%d) and cpusubtype (%d) does not match previous archive members cputype (%d) and cpusubtype (%d) (all members must match)", arName, mh.Cpu, mh.SubCpu&lmacho.MaskSubCpuType, prevCPU, prevSubCPU&lmacho.MaskSubCpuType)
	}

	prevCPU := macho.Cpu(0)
	prevSubCPU := lmacho.SubCpu(0)
	for {
		obj, err := archiver.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, nil, err
		}

		if strings.HasPrefix(obj.Name, ar.PrefixSymdef) {
			continue
		}

		if _, err := obj.Seek(0, io.SeekStart); err != nil {
			return false, nil, err
		}

		// TODO check fat
		mh, err := macho.NewFile(obj)
		if err != nil {
			return false, nil, err
		}

		if prevCPU == 0 && prevSubCPU == 0 {
			prevCPU = mh.Cpu
			prevSubCPU = mh.SubCpu
		}

		if prevCPU != mh.Cpu {
			return false, nil, makeErr(obj.Name, mh.FileHeader, prevCPU, prevSubCPU)
		}

		if mh.Magic == macho.Magic32 {
			if mh.Cpu == lmacho.CpuTypeArm {
				if mh.SubCpu != prevSubCPU {
					return false, nil, makeErr(obj.Name, mh.FileHeader, prevCPU, prevSubCPU)
				}
			}
		}
		if mh.Magic == macho.Magic64 {
			if mh.Cpu == lmacho.CpuTypeX86_64 || mh.Cpu == lmacho.CpuTypeArm64 {
				if mh.SubCpu != prevSubCPU {
					return false, nil, makeErr(obj.Name, mh.FileHeader, prevCPU, prevSubCPU)
				}
			}
		}
	}
	return true, &macho.FileHeader{
		Cpu:    prevCPU,
		SubCpu: prevSubCPU,
	}, nil
}

func newFatArches(arches ...*ArchInput) (fatArches, error) {
	fatArches := make(fatArches, len(arches))
	for i, arch := range arches {
		fa, err := lmacho.NewFatArch(arch.Bin)
		if err != nil {
			return nil, err
		}
		if arch.Arch != "" {
			if cpu := lmacho.ToCpuString(fa.Cpu, fa.SubCpu); cpu != arch.Arch {
				return nil, fmt.Errorf("specified architecture: %s for input file: %s does not match the file's architecture", arch.Arch, arch.Bin)
			}
		}
		fatArches[i] = *fa
	}

	dup := util.Duplicates(fatArches, func(v lmacho.FatArch) string {
		return lmacho.ToCpuString(v.Cpu, v.SubCpu)
	})
	if dup != nil {
		return nil, fmt.Errorf("duplicate architecture: %s", *dup)
	}

	return fatArches, nil
}

func validateOneInput(inputs []string) error {
	num := len(inputs)
	if num == 0 {
		return errNoInput
	} else if num != 1 {
		return fmt.Errorf("only one input file can be specified")
	}
	return nil
}

func validateInputArches(arches []string) error {
	dup := util.Duplicates(arches, func(v string) string { return v })
	if dup != nil {
		return fmt.Errorf("architecture %s specified multiple times", *dup)
	}

	for _, arch := range arches {
		if !lmacho.IsSupportedCpu(arch) {
			return fmt.Errorf(unsupportedArchFmt, arch)
		}
	}
	return nil
}

func perm(f string) (fs.FileMode, error) {
	info, err := os.Stat(f)
	if err != nil {
		return 0, err
	}
	perm := info.Mode().Perm() & 07777
	return perm, nil
}

// remove return values `a` does not have
func remove[T comparable](a []T, b []T) T {
	m := util.ExistMap(a, func(t T) T { return t })
	for _, v := range b {
		if _, ok := m[v]; !ok {
			return v
		}
	}
	return *new(T)
}
