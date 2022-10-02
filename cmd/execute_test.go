package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestExecute(t *testing.T) {

	tests := []struct {
		name         string
		args         []string
		wantMsg      string
		wantExitCode int
	}{
		// {
		// 	name:         "no value specified",
		// 	wantExitCode: 1,
		// },
		// {
		// 	name:         "create but no input",
		// 	args:         []string{"-create", "-output", "out", "in", "in"},
		// 	wantMsg:      "no such file or directory for in",
		// 	wantExitCode: 1,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			gotExitCode := Execute(buf, tt.args)
			if gotExitCode != tt.wantExitCode {
				t.Errorf("want: %d, got: %d", gotExitCode, tt.wantExitCode)
			}
			gotMsg := buf.String()
			if !strings.Contains(gotMsg, tt.wantMsg) {
				t.Errorf("want: %s, got: %s", tt.wantMsg, gotMsg)
			}
		})
	}
}
