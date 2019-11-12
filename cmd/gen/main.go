//go:generate go run github.com/go-bindata/go-bindata/go-bindata ../../avl.go
package main

import (
	"flag"
	"io/ioutil"
	"os/exec"
	"regexp"
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

	template := string(MustAsset("../../avl.go"))
	replace(&template, `package ordmap`, "package "+(*pkg))
	replace(&template, `\bKey\b`, *key)
	replace(&template, `\bValue\b`, *value)
	replace(&template, `(?:\b|_)OrdMap\b`, *name)
	replace(&template, `\bEntry\b`, (*name)+"Entry")
	replace(&template, `\bIterator\b`, (*name)+"Iterator")
	replace(&template, `\biteratorStackFrame\b`, (*name)+"IteratorStackFrame")
	replace(&template, `\b\.Less\b`, *less)

	err := ioutil.WriteFile(*target, []byte(template), 0644)
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
