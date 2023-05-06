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
	fs := sflag.NewFlagSet("a custom flag definition")

	timeValue := sflag.NewValue(new(time.Duration), time.ParseDuration)
	timeFlag := sflag.Register(fs, timeValue, "time", "-time 20s", sflag.WithDenyDuplicate())

	p := new([]time.Duration)
	timesValue := sflag.NewValue(p, func(v string) ([]time.Duration, error) {
		pd, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		*p = append(*p, pd)
		return *p, nil
	})
	timesFlag := sflag.Register(fs, timesValue, "times", "-times 20s -times 30h")

	err := fs.Parse([]string{"-time", "7200s", "-times", "20m", "-times", "30h"})
	if err != nil {
		panic(err)
	}

	fmt.Println(timeFlag.Get().Hours()) // 2
	fmt.Println(timesFlag.Get())        // [20m0s 30h0m0s]
}

func main() {
	example1()
	example2()
}
```
