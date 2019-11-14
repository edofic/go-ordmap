package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

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

	buf := bytes.NewBuffer(make([]byte, 0))
	fmt.Fprintf(buf, "package %s\n\n", *pkg)
	fmt.Fprintf(buf, "var %v = string([]byte{", *name)
	for i, char := range content {
		if i > 0 {
			fmt.Fprint(buf, ", ")
		}
		fmt.Fprint(buf, char)
	}
	fmt.Fprint(buf, "})\n")
	ioutil.WriteFile(*dest, buf.Bytes(), 0644)
}
