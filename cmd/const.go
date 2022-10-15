package cmd

const (
	createDescription = `
Create an universal binary(fat binary) from input thin binaries. 
e.g. lipo path/to/binary.x86_64 path/to/binary.arm64e -create -output path/to/fat-binary
`
	extractDescription = `
Extract the specified architecture from a universal binary and create a new universal binary.
e.g. lipo path/to/fat-binary -extract arm64e -extract x86_64h -output path/to/new-fat-binary
`
	extractFamilyDescription = `
Extract the specified architecture family from a universal binary and create a new universal binary.
e.g. lipo path/to/fat-binary -extract-family x86_64 -output path/to/new-fat-binary
`
	removeDescription = `
Remove the specified architecture from a universal binary and create a new universal binary.
e.g. lipo path/to/fat-binary -extract x86_64 -extract x86_64h -output path/to/new-fat-binary	
`
	replaceDescription = `
Replace the specified architecture in a universal binary with the specified input binary.
e.g. lipo path/to/fat-binary -replace x86_64 path/to/binary.x86_64 -output path/to/new-fat-binary.
`
	thinDescription = `
Extract a single-architecture-binary from a universal binary and create a single binary.
e.g. lipo path/to/fat-binary -thin arm64e -output path/to/binary.arm64e
`
	archsDescription = `
List all architectures contained in a universal binary
e.g. lipo path/to/fat-binary -archs
`
	verifyArchDescription = `
Verify the specified architecture are present in an universal binary.
If present, exit status of 0 otherwise exit status of 1.
e.g. lipo path/to/fat-binary x86_64 x86_64h arm64 arm64e
`
	infoDescription = `
List architectures in the input universal binary.
e.g. lipo path/to/fat-binary path/to/binary.x86_64 -info
`

	detailedInfoDescription = `
Display detailed information about universal binaries.
e.g. lipo path/to/fat-binary -detailed_info
`
)
