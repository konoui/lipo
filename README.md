## LIPO

This `lipo` aims to be compatible with macOS `lipo`.  
macOS lipo is known as an utility for Universal Binary a.k.a Fat Binary.

This is useful for following use-cases.

- On CI/CD (e.g. GitLab) not providing macOS or [macOS `lipo`](https://ss64.com/osx/lipo.html).
- On GitHub Actions not using macOS for [cost-effective](https://docs.github.com/en/billing/managing-billing-for-github-actions/about-billing-for-github-actions)

### INSTALL

#### Download [a latest release from GitHub](https://github.com/konoui/lipo/releases/latest)

For example for Linux on amd64,

```
$ curl -L -o /tmp/lipo https://github.com/konoui/lipo/releases/latest/download/lipo_Linux_amd64
$ chmod +x /tmp/lipo
$ sudo mv /tmp/lipo /usr/local/bin
```

#### Install with `go install`

```
$ go install github.com/konoui/lipo@latest
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

### Supported Options

`-archs`, `-create`, `-extract`, `-extract_family`, `-output`, `-remove`, `-replace`, `-segalign`, `-thin`, `-verify_arch`, `-arch`, `-info`
