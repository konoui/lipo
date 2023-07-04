package cmd

const (
	createDescription = `
Create a universal binary (also known as a fat binary) from input thin binaries.
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
e.g. lipo path/to/fat-binary -remove x86_64 -remove x86_64h -output path/to/new-fat-binary	
`
	replaceDescription = `
Replace the specified architecture in a universal binary with the specified input binary.
e.g. lipo path/to/fat-binary -replace x86_64 path/to/binary.x86_64 -output path/to/new-fat-binary
`
	thinDescription = `
Extract a single-architecture binary from a universal binary and create a single binary.
e.g. lipo path/to/fat-binary -thin arm64e -output path/to/binary.arm64e
`
	archsDescription = `
List all architectures contained in a universal binary.
e.g. lipo path/to/fat-binary -archs
`
	verifyArchDescription = `
Verify that the specified architectures are present in a universal binary.
If present, the exit status is 0; otherwise, the exit status is 1.
e.g. lipo path/to/fat-binary x86_64 x86_64h arm64 arm64e -verify_arch
`
	infoDescription = `
Display brief information on architectures in universal binaries.
e.g. lipo path/to/fat-binary path/to/binary.x86_64 -info
`

	detailedInfoDescription = `
Display detailed information about universal binaries.
e.g. lipo path/to/fat-binary path/to/binary.x86_64 -detailed_info
`
)
