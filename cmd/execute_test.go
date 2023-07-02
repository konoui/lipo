package cmd_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/konoui/lipo/cmd"
	"github.com/konoui/lipo/pkg/testlipo"
)

var bm = testlipo.NewBinManager(testlipo.TestDir)

const (
	phOutput     = "<output_file>"
	phInputFat   = "<input_fat_file>"
	phInputThins = "<input_thin_files>"
	phArm64Thin  = "<input_arm64_thin_file>"
	phX86_64Thin = "<input_x86_64_thin_file>"
)

func replace(t *testing.T, p *testlipo.TestLipo, rawArgs []string, mylipo bool) (args []string, outBin string) {
	args = []string{}
	for _, arg := range rawArgs {
		in := arg
		if arg == phOutput {
			outBin = filepath.Join(p.Dir, "output-"+filepath.Base(t.Name()))
			if mylipo {
				outBin += "-mylipo"
			}
			in = strings.ReplaceAll(in, phOutput, outBin)
		}
		in = strings.ReplaceAll(in, phInputFat, p.FatBin)
		in = strings.ReplaceAll(in, phArm64Thin, p.Bin(t, "arm64"))
		in = strings.ReplaceAll(in, phX86_64Thin, p.Bin(t, "x86_64"))
		if in == phInputThins {
			args = append(args, p.Bins(t)...)
			continue
		}
		args = append(args, in)
	}
	return args, outBin
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
			name:         "extract_family with extract",
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
			name:         "-info",
			wantExitCode: 0,
			args:         []string{phInputFat, phArm64Thin, phArm64Thin, "-info"},
		},
		{
			name:         "-detailed_info",
			wantExitCode: 0,
			args:         []string{phInputFat, phInputFat, "-detailed_info"},
		},
		{
			name:         "create with segalign",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phInputThins, "-segalign", "x86_64", "2"},
		},
		{
			name:         "create with arch",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phArm64Thin, "-arch", "x86_64", phX86_64Thin},
		},
		{
			name:         "create with arch with short flag",
			wantExitCode: 0,
			args:         []string{"-c", "-o", phOutput, phArm64Thin, "-a", "x86_64", phX86_64Thin},
		},
		{
			name:         "create with segalign and hideARM64",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phInputThins, "-segalign", "arm64", "1", "-hideARM64"},
			addArches:    []string{"armv7k"},
		},
		{
			name:         "create with fat64",
			wantExitCode: 0,
			args:         []string{"-create", "-output", phOutput, phInputThins, "-fat64"},
		},
		{
			name:         "verify_arch not contains",
			wantExitCode: 1,
			args:         []string{phInputFat, "-verify_arch", "arm64", "x86_64", "arm64e"},
		},
		{
			name:         "-detailed_info but not fat",
			wantExitCode: 0,
			args:         []string{phInputFat, phArm64Thin, "-detailed_info"},
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
			p := testlipo.Setup(t, bm, append(tt.addArches, "arm64", "x86_64"))
			args, gotBin := replace(t, p, tt.args, true)

			gotExitCode := cmd.Execute(outBuf, errBuf, args)
			gotErrMsg := errBuf.String()
			if gotExitCode != tt.wantExitCode {
				t.Errorf("want: %d, got: %d", tt.wantExitCode, gotExitCode)
				t.Log(gotErrMsg)
			}
			if !strings.Contains(gotErrMsg, tt.wantErrMsg) {
				t.Errorf("want: %s, got: %s", tt.wantErrMsg, gotErrMsg)
			}

			if tt.wantExitCode == 0 {
				if gotBin != "" {
					testArgs, wantBin := replace(t, p, tt.args, false)
					lipoExecute(t, p, testArgs)
					diffSha256(t, wantBin, gotBin)
				} else {
					gotResp := outBuf.String()
					wantResp := lipoExecute(t, p, args)
					diffByLine(t, wantResp, gotResp)
				}
			}
		})
	}
}

func lipoExecute(t *testing.T, p *testlipo.TestLipo, args []string) string {
	cmd := exec.Command(p.LipoBin.Bin, args...)
	resp, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("lipoExecute failed: %v\n%s", err, string(resp))
	}
	return string(resp)
}

func diffByLine(t *testing.T, want string, got string) {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	if len(w) != len(g) {
		t.Errorf("len(want) = %d len(got) = %d\n", len(w), len(g))
		return
	}
	for i := 0; i < len(w); i++ {
		// TODO FIXME
		if w[i] != g[i] {
			if w[i] != g[i]+" " {
				t.Errorf("want: %s\ngot: %s\n", w[i], g[i])
			}
		}
	}
}

func diffSha256(t *testing.T, wantBin, gotBin string) {
	t.Helper()
	testlipo.DiffPerm(t, wantBin, gotBin)
	testlipo.PatchFat64Reserved(t, wantBin)
	testlipo.DiffSha256(t, wantBin, gotBin)
}
