package lmacho

import "debug/macho"

const MagicFat64 = macho.MagicFat + 1

type fatArch64Header struct {
	FatArchHeader
	Reserved uint32
}
