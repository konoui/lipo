package lipo

import (
	"debug/macho"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/konoui/lipo/pkg/lipo/mcpu"
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
	type fatArch struct {
		CpuType      string
		SubCpuType   string
		Arch         string
		Capabilities string
		Offset       uint32
		Size         uint32
		AlignBit     uint32
		Align        int
	}

	type fatBinary struct {
		FatBinary string
		FatMagic  string
		NFatArch  int
		Arches    []*fatArch
	}

	var out strings.Builder
	fat, err := OpenFat(bin)
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
	fat.Close()

	fb := &fatBinary{
		FatBinary: bin,
		FatMagic:  fmt.Sprintf("0x%x", fat.Magic),
		NFatArch:  len(fat.Arches),
		Arches:    make([]*fatArch, 0, len(fat.Arches)),
	}
	for _, a := range fat.Arches {
		c, s := mcpu.StringValues(a.Cpu, a.SubCpu)
		arch := &fatArch{
			Arch:         mcpu.ToString(a.Cpu, a.SubCpu),
			CpuType:      c,
			SubCpuType:   s,
			Capabilities: fmt.Sprintf("0x%x", (a.SubCpu&mcpu.MaskSubType)>>24),
			Offset:       a.Offset,
			Size:         a.Size,
			AlignBit:     a.Align,
			Align:        1 << a.Align,
		}
		fb.Arches = append(fb.Arches, arch)
	}

	if err := tpl.Execute(&out, *fb); err != nil {
		return "", false, err
	}
	return out.String(), true, nil
}
