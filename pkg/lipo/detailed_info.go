package lipo

import (
	"bytes"
	"debug/macho"
	"errors"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
	"github.com/konoui/lipo/pkg/util"
)

const detailedInfoTpl = `Fat header in: {{ .FatBinary }}
fat_magic {{ .FatMagic }}
nfat_arch {{ .NFatArch }}
{{ range $i, $v := .Arches -}}
architecture {{ .Arch }}
    cputype {{ .CpuType }}
    cpusubtype {{ .SubCpuType }}
    capabilities {{ .Capabilities }}
    offset {{ .Offset }}
    size {{ .Size }}
    align 2^{{ .AlignBit }} ({{ .Align }})
{{ end -}}	
`

var tpl = template.Must(template.New("detailed_info").Parse(detailedInfoTpl))

func (l *Lipo) DetailedInfo(w io.Writer) error {
	if len(l.in) == 0 {
		return errNoInput
	}

	var out bytes.Buffer

	thin := []string{}
	for _, bin := range l.in {
		v, isFat, err := detailedInfo(bin)
		if err != nil {
			return err
		}
		if isFat {
			out.WriteString(v)
		} else {
			thin = append(thin, v)
		}
	}

	// append thin
	if len(thin) > 0 {
		out.WriteString(strings.Join(thin, "\n") + "\n")
	}

	_, err := w.Write(out.Bytes())
	return err
}

type tplFatArch struct {
	CpuType      string
	SubCpuType   string
	Arch         string
	Capabilities string
	Offset       uint64
	Size         uint64
	AlignBit     uint32
	Align        int
}

type tplFatBinary struct {
	FatBinary string
	FatMagic  string
	NFatArch  string
	Arches    []*tplFatArch
}

func detailedInfo(bin string) (string, bool, error) {
	var out strings.Builder
	ff, err := lmacho.NewFatFile(bin)
	if err != nil {
		var e *lmacho.FormatError
		if errors.As(err, &e) {
			return "", false, fmt.Errorf("can't figure out the architecture type of: %s: %w", bin, err)
		} else if !errors.Is(err, macho.ErrNotFat) {
			return "", false, err
		}

		// if not fat file, assume single macho file
		v, _, err := info(bin)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("input file %s is not a fat file\n%s", bin, v), false, nil
	}

	nFatArch := fmt.Sprintf("%d", len(ff.Arches))
	if len(ff.HiddenArches) > 0 {
		nFatArch = fmt.Sprintf("%d (+%d hidden)", len(ff.Arches), len(ff.HiddenArches))
	}
	fb := &tplFatBinary{
		FatBinary: bin,
		FatMagic:  fmt.Sprintf("0x%x", ff.Magic),
		NFatArch:  nFatArch,
		Arches:    make([]*tplFatArch, 0, len(ff.Arches)),
	}
	fb.Arches = util.Map(ff.Arches, tplArch)
	fb.Arches = append(fb.Arches,
		util.Map(ff.HiddenArches, func(v lmacho.FatArch) *tplFatArch {
			ta := tplArch(v)
			ta.Arch = fmt.Sprintf("%s (hidden)", ta.Arch)
			return ta
		})...)
	if err := tpl.Execute(&out, *fb); err != nil {
		return "", false, err
	}
	return out.String(), true, nil
}

func tplArch(a lmacho.FatArch) *tplFatArch {
	c, s := lmacho.ToCpuValues(a.Cpu, a.SubCpu)
	arch := lmacho.ToCpuString(a.Cpu, a.SubCpu)
	return &tplFatArch{
		Arch:         arch,
		CpuType:      c,
		SubCpuType:   s,
		Capabilities: fmt.Sprintf("0x%x", (a.SubCpu&lmacho.MaskSubCpuType)>>24),
		Offset:       a.Offset,
		Size:         a.Size,
		AlignBit:     a.Align,
		Align:        1 << a.Align,
	}
}
