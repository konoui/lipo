package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/pkg/testlipo"
)

const (
	phOutput     = "<output_file>"
	phInputFat   = "<input_fat_file>"
	phInputThins = "<input_thin_files>"
	phArm64Thin  = "<input_arm64_thin_file>"
	phX86_64Thin = "<input_x86_64_thin_file>"
)

func replace(t *testing.T, p *testlipo.TestLipo, args []string) []string {
	ret := []string{}
	for _, arg := range args {
		in := arg
		in = strings.ReplaceAll(in, phOutput, filepath.Join(p.Dir, "output-"+filepath.Base(t.Name())))
		in = strings.ReplaceAll(in, phInputFat, p.FatBin)
		in = strings.ReplaceAll(in, phArm64Thin, p.Bin(t, "arm64"))
		in = strings.ReplaceAll(in, phX86_64Thin, p.Bin(t, "x86_64"))
		if in == phInputThins {
			ret = append(ret, p.Bins()...)
			continue
		}
		ret = append(ret, in)
	}
	return ret
}

func TestExecute(t *testing.T) {

	tests := []struct {
		name         string
		args         []string
		addArches    []string
		wantErrMsg   string
		wantExitCode int
	}{
		{
			name:         "create",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phInputThins},
		},
		{
			name:         "thin",
			wantExitCode: 0,
			args:         []string{"-thin", "x86_64", "-output", phOutput, phInputFat},
		},
		{
			name:         "remove",
			wantExitCode: 0,
			args:         []string{"-remove", "x86_64", "-output", phOutput, phInputFat},
		},
		{
			name:         "remove two arches from 3 fat binary",
			wantExitCode: 0,
			args:         []string{"-remove", "x86_64", "-remove", "arm64", "-output", phOutput, phInputFat},
			addArches:    []string{"arm64e"},
		},
		{
			name:         "extract",
			wantExitCode: 0,
			args:         []string{"-extract", "x86_64", "-extract", "arm64", "-output", phOutput, phInputFat},
		},
		{
			name:         "extract",
			wantExitCode: 0,
			args:         []string{"-extract", "x86_64", "-extract_family", "arm64", "-output", phOutput, phInputFat},
		},
		{
			name:         "replace",
			wantExitCode: 0,
			args:         []string{"-replace", "arm64", phArm64Thin, "-replace", "x86_64", phX86_64Thin, "-output", phOutput, phInputFat},
		},
		{
			name:         "archs",
			wantExitCode: 0,
			args:         []string{"-archs", phInputFat},
		},
		{
			name:         "verify_arch",
			wantExitCode: 0,
			args:         []string{phInputFat, "-verify_arch", "arm64", "x86_64"},
		},
		{
			name:         "verify_arch not contains",
			wantExitCode: 1,
			args:         []string{phInputFat, "-verify_arch", "arm64", "x86_64", "arm64e"},
		},
		{
			name:         "create",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phInputThins, "-segalign", "x86_64", "2"},
		},
		{
			name:         "TODO usage if no inputs",
			wantExitCode: 1,
		},
		{
			name:         "create but no input",
			args:         []string{"-create", "-output", "out", "in", "in"},
			wantErrMsg:   "no such file or directory",
			wantExitCode: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf, errBuf := &bytes.Buffer{}, &bytes.Buffer{}
			p := testlipo.Setup(t, append(tt.addArches, "arm64", "x86_64")...)
			args := replace(t, p, tt.args)

			gotExitCode := Execute(outBuf, errBuf, args)
			gotErrMsg := errBuf.String()
			if gotExitCode != tt.wantExitCode {
				t.Errorf("want: %d, got: %d", tt.wantExitCode, gotExitCode)
				t.Log(gotErrMsg)
			}
			if !strings.Contains(gotErrMsg, tt.wantErrMsg) {
				t.Errorf("want: %s, got: %s", tt.wantErrMsg, gotErrMsg)
			}
		})
	}
}
