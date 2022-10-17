package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
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

func detailedInfo(bin string) (string, bool, error) {
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

	var out strings.Builder
	fatArches, err := fatArchesFromFatBin(bin)
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

	hideArches := util.Filter(fatArches, func(v *fatArch) bool { return v.hidden })
	nFatArch := fmt.Sprintf("%d", len(fatArches))
	if len(hideArches) > 0 {
		nFatArch = fmt.Sprintf("%d (+%d hidden)", len(fatArches)-len(hideArches), len(hideArches))
	}
	fb := &tplFatBinary{
		FatBinary: bin,
		FatMagic:  fmt.Sprintf("0x%x", macho.MagicFat),
		NFatArch:  nFatArch,
		Arches:    make([]*tplFatArch, 0, len(fatArches)),
	}
	for _, a := range fatArches {
		c, s := mcpu.StringValues(a.Cpu, a.SubCpu)
		arch := mcpu.ToString(a.Cpu, a.SubCpu)
		if a.hidden {
			arch = fmt.Sprintf("%s (hidden)", arch)
		}
		fb.Arches = append(fb.Arches, &tplFatArch{
			Arch:         arch,
			CpuType:      c,
			SubCpuType:   s,
			Capabilities: fmt.Sprintf("0x%x", (a.SubCpu&mcpu.MaskSubType)>>24),
			Offset:       a.Offset,
			Size:         a.Size,
			AlignBit:     a.Align,
			Align:        1 << a.Align,
		})
	}

	if err := tpl.Execute(&out, *fb); err != nil {
		return "", false, err
	}
	return out.String(), true, nil
}
