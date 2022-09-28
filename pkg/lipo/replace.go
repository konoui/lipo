package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func ReplaceInputs(rawInputs [][]string) ([]*ReplaceInput, error) {
	if len(rawInputs) == 0 {
		return nil, errors.New("no replace inputs")
	}

	rinputs := make([]*ReplaceInput, 0, len(rawInputs))
	for _, rawIn := range rawInputs {
		if len(rawIn) != 2 {
			return nil, fmt.Errorf("inputs are not arch/file pair %v", rawIn)
		}
		rinputs = append(rinputs, &ReplaceInput{Arch: rawIn[0], Bin: rawIn[1]})
	}
	return rinputs, nil
}

type ReplaceInput struct {
	Arch string
	Bin  string
}

func arches(input []*ReplaceInput) []string {
	arches := make([]string, 0, len(input))
	for _, ri := range input {
		arches = append(arches, ri.Arch)
	}
	return arches
}

func bins(input []*ReplaceInput) []string {
	b := make([]string, 0, len(input))
	for _, ri := range input {
		b = append(b, ri.Bin)
	}
	return b
}

func (l *Lipo) Replace(inputs []*ReplaceInput) error {
	if len(l.in) == 0 {
		return errors.New("no inputs")
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	targets, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		return contain(mcpu.ToString(hdr.Cpu, hdr.SubCpu), arches(inputs))
	})
	if err != nil {
		return fmt.Errorf("search error: arches from fat file: %w", err)
	}
	defer func() { _ = close(targets) }()

	target := targets[0]

	in, err := newCreateInputs(bins(inputs)...)
	if err != nil {
		return fmt.Errorf("create error: from input file: %w", err)
	}

	fatInputs, err := fatArchesFromCreateInputs(in)
	if err != nil {
		return fmt.Errorf("create error: from arches: %w", err)
	}
	defer func() { _ = close(fatInputs) }()

	to := fatInputs[0]

	if !(target.Cpu == to.Cpu && target.SubCpu == to.SubCpu) {
		return errors.New("unexpected input/arch")
	}

	others, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		return !contain(mcpu.ToString(hdr.Cpu, hdr.SubCpu), arches(inputs))
	})
	if err != nil {
		return fmt.Errorf("search error: not-match-arches from fat file: %w", err)
	}
	defer func() { _ = close(others) }()

	fatArches, err := sortByArch(append(others, to))
	if err != nil {
		return fmt.Errorf("sort error: %w", err)
	}

	err = outputFatBinary(l.out, perm, fatArches)
	if err != nil {
		return fmt.Errorf("output fat error: %w", err)
	}

	return nil
}
