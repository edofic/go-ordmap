//go:generate go run ../../internal/pack/main.go -source ../../avl.go -name avl -dest template_avl.go
//go:generate go run ../../internal/pack/main.go -source ../../experimental/btree.go -name btree -dest template_btree.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	pkg := flag.String("pkg", "main", "Package name to use for generated code")
	name := flag.String("name", "OrdMap", "Name of the generated type")
	key := flag.String("key", "Key", "Name of the key type to use")
	value := flag.String("value", "Value", "Name of the value type to use")
	target := flag.String("target", "./ordmap.go", "Path for the generated code")
	less := flag.String("less", ".Less", "Operation to use for comparison")
	doFmt := flag.Bool("fmt", true, "Run `go fmt` on the generated files")
	flag.Parse()

	code := string(avl())
	replace(&code, `package ordmap`, "package "+(*pkg))
	replace(&code, `\bKey\b`, *key)
	replace(&code, `\bValue\b`, *value)
	replace(&code, `(?:\b|_)OrdMap\b`, *name)
	replace(&code, `\bEntry\b`, (*name)+"Entry")
	replace(&code, `\bIterator\b`, (*name)+"Iterator")
	replace(&code, `\biteratorStackFrame\b`, (*name)+"IteratorStackFrame")
	replace(&code, `\b\.Less\b`, *less)

	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintln(buf, "// DO NOT EDIT tis code was generated using go-ordmap code generation")
	fmt.Fprintln(buf, "// go run github.com/edofic/go-ordmap/cmd/gen "+strings.Join(os.Args[1:], " "))
	buf.WriteString(code)

	err := ioutil.WriteFile(*target, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	if *doFmt {
		err = exec.Command("go", "fmt", *target).Run()
		if err != nil {
			panic(err)
		}
	}
}

func replace(src *string, re, repl string) {
	*src = regexp.MustCompile(re).ReplaceAllString(*src, repl)
}
