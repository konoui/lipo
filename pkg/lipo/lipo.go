package lipo

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

// createTemp creates a temporary file from file path
func createTemp(path string) (*os.File, error) {
	f, err := os.CreateTemp(filepath.Dir(path), "tmp-lipo-out")
	if err != nil {
		return nil, fmt.Errorf("can't create temporary output file: %w", err)
	}
	return f, nil
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

func perm(f string) (fs.FileMode, error) {
	info, err := os.Stat(f)
	if err != nil {
		return 0, err
	}
	perm := info.Mode().Perm() & 07777
	return perm, nil
}

// diff return values `a` does not have
func diff[T comparable](a []T, b []T) T {
	m := util.ExistenceMap(a, func(t T) T { return t })
	for _, v := range b {
		if _, ok := m[v]; !ok {
			return v
		}
	}
	return *new(T)
}
