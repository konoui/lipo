package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/konoui/lipo/pkg/lipo/lmacho"
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

func (l *Lipo) DetailedInfo() (string, error) {
	if len(l.in) == 0 {
		return "", errNoInput
	}

	var out strings.Builder

	thin := []string{}
	for _, bin := range l.in {
		v, isFat, err := detailedInfo(bin)
		if err != nil {
			return "", err
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
	return out.String(), nil
}

type tplFatArch struct {
	CpuType      string
	SubCpuType   string
	Arch         string
	Capabilities string
	Offset       uint32
	Size         uint32
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
	ff, err := lmacho.OpenFat(bin)
	if err != nil {
		if !errors.Is(err, macho.ErrNotFat) {
			return "", false, err
		}
		// fallback info if thin file
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
		FatMagic:  fmt.Sprintf("0x%x", macho.MagicFat),
		NFatArch:  nFatArch,
		Arches:    make([]*tplFatArch, 0, len(ff.Arches)),
	}
	for _, a := range ff.Arches {
		fb.Arches = append(fb.Arches, tplArch(a))
	}
	for _, a := range ff.HiddenArches {
		ta := tplArch(a)
		ta.Arch = fmt.Sprintf("%s (hidden)", ta.Arch)
		fb.Arches = append(fb.Arches, ta)
	}

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