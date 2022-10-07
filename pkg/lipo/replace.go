package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

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
		return fmt.Errorf("search error: %w", err)
	}
	defer func() { _ = close(targets) }()

	if len(targets) != len(inputs) {
		return fmt.Errorf("replace inputs: want %d but got %d", len(targets), len(inputs))
	}

	in, err := newCreateInputs(bins(inputs)...)
	if err != nil {
		return err
	}

	fatInputs, err := fatArchesFromCreateInputs(in)
	if err != nil {
		return fmt.Errorf("error fat arches: %w", err)
	}
	defer func() { _ = close(fatInputs) }()

	sort.Slice(targets, func(i, j int) bool {
		ih, jh := targets[i], targets[j]
		return mcpu.ToString(ih.Cpu, ih.SubCpu) < mcpu.ToString(jh.Cpu, jh.SubCpu)

	})
	sort.Slice(fatInputs, func(i, j int) bool {
		ih, jh := fatInputs[i], fatInputs[j]
		return mcpu.ToString(ih.Cpu, ih.SubCpu) < mcpu.ToString(jh.Cpu, jh.SubCpu)
	})
	for idx := range fatInputs {
		from, to := targets[idx], fatInputs[idx]
		fromArch := mcpu.ToString(from.Cpu, from.SubCpu)
		toArch := mcpu.ToString(to.Cpu, to.SubCpu)
		if fromArch != toArch {
			return fmt.Errorf("specified architecture: %s for replacement file: %s does not match the file's architecture", fromArch, toArch)
		}
	}

	others, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		return !contain(mcpu.ToString(hdr.Cpu, hdr.SubCpu), arches(inputs))
	})
	if err != nil {
		// Note ignore case that replace all arches of fat bin with inputs
		// e.g.) fat bin contains only arm64, replace arm64 to new arm64 bin
		if !errors.Is(err, errFoundNoFatArch) {
			return fmt.Errorf("search error: %w", err)
		}
	}
	defer func() { _ = close(others) }()

	fatArches, err := sortByArch(append(others, fatInputs...))
	if err != nil {
		return err
	}

	if err := updateAlignBit(fatArches, l.segAligns); err != nil {
		return err
	}

	return outputFatBinary(l.out, perm, fatArches)
}
