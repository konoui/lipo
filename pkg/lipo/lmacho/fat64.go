package lmacho

type fatArch64Header struct {
	FatArchHeader
	Reserved uint32
}
