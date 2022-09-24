package lipo

import (
	"debug/macho"
	"errors"
	"os"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
)

func (l *Lipo) Replace(arch, bin string) error {
	if len(l.in) != 1 {
		return errors.New("input must be 1")
	}

	fatBin := l.in[0]
	info, err := os.Stat(fatBin)
	if err != nil {
		return err
	}
	perm := info.Mode().Perm()

	targets, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		return arch == mcpu.ToString(hdr.Cpu, hdr.SubCpu)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(targets) }()

	target := targets[0]

	in, err := newCreateInputs(bin)
	if err != nil {
		return err
	}

	inputs, err := fatArchesFromCreateInputs(in)
	if err != nil {
		return err
	}
	defer func() { _ = close(inputs) }()

	to := inputs[0]

	if !(target.Cpu == to.Cpu && target.SubCpu == to.SubCpu) {
		return errors.New("unexpected input/arch")
	}

	others, err := fatArchesFromFatBin(fatBin, func(hdr *macho.FatArchHeader) bool {
		return arch != mcpu.ToString(hdr.Cpu, hdr.SubCpu)
	})
	if err != nil {
		return err
	}
	defer func() { _ = close(others) }()

	fatArches, err := sortByArch(append(others, to))
	if err != nil {
		return err
	}

	return outputFatBinary(l.out, perm, fatArches)
}
