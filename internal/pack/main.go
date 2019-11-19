package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func compress(raw []byte) []byte {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(raw)
	w.Close()
	return b.Bytes()
}

func formatOutput(pkg, name string, compressed []byte) []byte {
	buf := bytes.NewBuffer(make([]byte, 0))
	fmt.Fprint(buf, "// DO NOT EDIT this code was generated using `pack` command\n")
	fmt.Fprintf(buf, "package %s\n\n", pkg)
	fmt.Fprintf(buf, "import (\n")
	fmt.Fprintln(buf, "\t\"bytes\"")
	fmt.Fprintln(buf, "\t\"compress/gzip\"")
	fmt.Fprintln(buf, "\t\"io/ioutil\"")
	fmt.Fprint(buf, ")\n\n")
	fmt.Fprintf(buf, "func %v() []byte {\n", name)
	fmt.Fprintf(buf, "\tbuf := bytes.NewBuffer([]byte{")
	for i, char := range compressed {
		if i > 0 {
			fmt.Fprint(buf, ", ")
		}
		fmt.Fprint(buf, char)
	}
	fmt.Fprintf(buf, "})")
	fmt.Fprintf(buf, `
	r, err := gzip.NewReader(buf)
	if err != nil {
		panic(err)
	}
	decompressed, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}
	return decompressed`)
	fmt.Fprintf(buf, "\n}")
	return buf.Bytes()
}

func main() {
	source := flag.String("source", "./source.go", "Source file to pack")
	dest := flag.String("dest", "./packed.go", "Destination to write packed code to")
	name := flag.String("name", "packed", "Variable name to use in generated code")
	pkg := flag.String("pkg", "main", "Package name to use")
	// TODO compression
	flag.Parse()

	content, err := ioutil.ReadFile(*source)
	if err != nil {
		log.Println("Cannot read file", *source)
		os.Exit(1)
	}
	ioutil.WriteFile(*dest, formatOutput(*pkg, *name, compress(content)), 0644)
}
