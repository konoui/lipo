## LIPO

`lipo` creates Universal Binary a.k.a Fat Binary for macOS.

This is useful for environments such as CI/CD which does not provides macOS or [macOS `lipo`](https://ss64.com/osx/lipo.html)

### INSTALL

#### Go Install command

```
$ go install github.com/konoui/lipo/pkg/lipo@latest
```

### USAGE

```
$ lipo -output <output-binary> -create <arm64-binary> <amd64-binary>
```

For example,

```
$ CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o amd64 example/main.go
$ CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o arm64 example/main.go
$ lipo -output hello-world -create arm64 amd64
```

```
$ ./hello-world
Hello World

$ file hello-world
hello-world: Mach-O universal binary with 2 architectures: [x86_64:Mach-O 64-bit executable x86_64] [arm64]
hello-world (for architecture x86_64): Mach-O 64-bit executable x86_64
hello-world (for architecture arm64): Mach-O 64-bit executable arm64
```

### Note

The `lipo` supports only 64-bit binary.
