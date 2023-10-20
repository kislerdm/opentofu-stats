package main

import (
	_ "embed"
	"flag"
	"io"
	"log"
	"os"
	"text/template"
)

//go:embed index.html.templ
var indexTemplate string

var indexPage = template.Must(template.New("webpage").Parse(indexTemplate))

func main() {
	var (
		inPath  string
		outPath string
	)
	flag.StringVar(&inPath, "in", "", "Path to read aggregates.")
	flag.StringVar(&outPath, "out", "public/index.html", "Path to store generated HTML page.")
	flag.Parse()

	if inPath == "" {
		println("ERROR: specify path to the input data file")
		flag.Usage()
		os.Exit(1)
	}

	fIN, err := os.Open(inPath)
	if err != nil {
		log.Fatalf("cannot read data from file %s: %v\n", inPath, err)
	}
	defer func() { _ = fIN.Close() }()

	data, err := io.ReadAll(fIN)
	if err != nil {
		log.Fatalf("cannot read from file %s: %v\n", outPath, err)
	}

	fHTML, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("cannot open file for saving HTML page %s: %v\n", outPath, err)
	}
	defer func() { _ = fHTML.Close() }()

	// html/template would require additional transformation of data,
	// hence rely on text/template because no html escape is required.
	if err := indexPage.Execute(fHTML, string(data)); err != nil {
		log.Fatalf("cannot generate HTML page: %v\n", err)
	}
}
