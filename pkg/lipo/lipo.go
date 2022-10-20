package lipo

import (
	"debug/macho"
	"errors"
	"fmt"

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

func WithHideArm64() Option {
	return func(l *Lipo) {
		l.hideArm64 = true
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

func hideARmObjectErr(arches fatArches) error {
	for _, arch := range arches {
		if arch.FileHeader.Type == macho.TypeObj {
			return fmt.Errorf("hideARM64 specified but thin file %s is not of type MH_EXECUTE", arch.Name)
		}
	}
	return nil
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
	return fatArches, lmacho.ValidateFatArches(fatArches)
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

func validateInputArches(arches []string) error {
	dup := duplicates(arches)
	if dup != "" {
		return fmt.Errorf("architecture %s specified multiple times", dup)
	}

	for _, arch := range arches {
		if !lmacho.IsSupportedCpu(arch) {
			return fmt.Errorf("unsupported architecture %s", arch)
		}
	}
	return nil
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
		if !util.Contains(a, v) {
			return v
		}
	}
	return ""
}
