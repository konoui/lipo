package lipo

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/konoui/lipo/pkg/lmacho"
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

func (l *Lipo) DetailedInfo(stdout, stderr io.Writer) {
	if len(l.in) == 0 {
		fmt.Fprintln(stderr, "fatal error: "+errNoInput.Error())
		return
	}

	var out strings.Builder

	thin := []string{}
	for _, bin := range l.in {
		v, isFat, err := detailedInfo(bin)
		if err != nil {
			fmt.Fprintln(stderr, "fatal error: "+err.Error())
			return
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

	fmt.Fprint(stdout, out.String())
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

	typ, err := inspect(bin)
	if err != nil {
		return "", false, err
	}

	if typ != inspectFat {
		v, _, err := info(bin)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("input file %s is not a fat file\n%s", bin, v), false, nil
	}

	ff, err := OpenFatFile(bin)
	if err != nil {
		return "", false, err
	}
	defer ff.Close()

	rawArches := util.Map(ff.Arches, func(v Arch) *lmacho.FatArch {
		return v.(*arch).Object.(*lmacho.FatArch)
	})

	hiddenArches := util.Filter(rawArches, func(v *lmacho.FatArch) bool {
		return v.Hidden
	})

	visibleArches := util.Filter(rawArches, func(v *lmacho.FatArch) bool {
		return !v.Hidden
	})

	nFatArch := fmt.Sprintf("%d", len(visibleArches))
	if len(hiddenArches) > 0 {
		nFatArch = fmt.Sprintf("%d (+%d hidden)", len(visibleArches), len(hiddenArches))
	}
	fb := &tplFatBinary{
		FatBinary: bin,
		FatMagic:  fmt.Sprintf("0x%x", ff.Magic),
		NFatArch:  nFatArch,
		Arches:    make([]*tplFatArch, 0, len(visibleArches)),
	}
	fb.Arches = util.Map(visibleArches, tplArch)
	fb.Arches = append(fb.Arches,
		util.Map(hiddenArches, func(v *lmacho.FatArch) *tplFatArch {
			ta := tplArch(v)
			ta.Arch = fmt.Sprintf("%s (hidden)", ta.Arch)
			return ta
		})...)
	if err := tpl.Execute(&out, *fb); err != nil {
		return "", false, err
	}
	return out.String(), true, nil
}

func tplArch(a *lmacho.FatArch) *tplFatArch {
	c, s := lmacho.ToCpuValues(a.CPU(), a.SubCPU())
	return &tplFatArch{
		Arch:         a.CPUString(),
		CpuType:      c,
		SubCpuType:   s,
		Capabilities: fmt.Sprintf("0x%x", (a.SubCPU()&lmacho.MaskSubCpuType)>>24),
		Offset:       a.Offset(),
		Size:         a.Size(),
		AlignBit:     a.Align(),
		Align:        1 << a.Align(),
	}
}
