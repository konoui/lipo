package lipo_test

import (
	"path/filepath"
	"testing"
)

func TestLipo_Create(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		p := setup(t)

		// check fat file format
		gotBin := filepath.Join(p.dir, "out-amd64-arm64-binary")
		createFatBin(t, gotBin, p.amd64Bin, p.arm64Bin)

		fatBin := filepath.Join(p.dir, "out-arm64-amd64-binary")
		createFatBin(t, fatBin, p.arm64Bin, p.amd64Bin)
		if p.skip() {
			t.Skip("skip lipo binary test")
		}

		p.lipoDetail(t, gotBin)
		p.lipoDetail(t, fatBin)
		diffSha256(t, p.lipoFatBin, gotBin)
	})
}
