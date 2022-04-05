## LIPO

`lipo` creates Universal Binary aka Fat Binary for macOS.

This is very useful on CI/CD which does not support macOS.

### INSTALL

```
$ go install github.com/konoui/lipo/pkg/lipo@latest
```

### USAGE

```
$ lipo -output <output-binary> -create <arm64-binary> <amd64-binary>
```

For example,

```
$ CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o amd64 main.go
$ CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o arm64 main.go
$ lipo -output hello-world -create amd64 amd64
```

```
$ ./hello-world
Hello World
$ file hello-world
hello-world: Mach-O universal binary with 2 architectures: [x86_64:Mach-O 64-bit executable x86_64] [arm64]
hello-world (for architecture x86_64): Mach-O 64-bit executable x86_64
hello-world (for architecture arm64): Mach-O 64-bit executable arm64
```

```
package main

import "fmt"

func main() {
    fmt.Println("Hello World")
}
```

### Note

`lipo` supports only 64-bit binary.
