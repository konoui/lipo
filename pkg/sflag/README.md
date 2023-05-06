## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/konoui/lipo/pkg/sflag"
)

func example1() {
	fs := sflag.NewFlagSet("simple usage")
	debug := fs.Bool("debug", "enable debug mode", sflag.WithShortName("d")) // allow -debug and -d
	values := fs.Strings("values", "multiple values")

	err := fs.Parse([]string{"-values", "a", "b", "c", "-d", "arg1", "arg2"})
	if err != nil {
		panic(err)
	}

	fmt.Println(fs.Args())    // [arg1 arg2]
	fmt.Println(debug.Get())  // true
	fmt.Println(values.Get()) // [a b c]
}

func example2() {
	fs := sflag.NewFlagSet("custom flag definition")
	value := sflag.FlagValue(new(time.Duration), time.ParseDuration)
	timeFlag := sflag.Register[time.Duration](fs, value, "time", "-time 20s")

	err := fs.Parse([]string{"-time", "7200s"})
	if err != nil {
		panic(err)
	}

	fmt.Println(timeFlag.Get().Hours()) // 2
}

func main() {
	example1()
	example2()
}
```
