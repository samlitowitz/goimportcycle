package main

import (
	"flag"
	"github.com/samlitowitz/goimportcycleviz/internal"
	"github.com/samlitowitz/goimportcycleviz/internal/dot"
	"github.com/samlitowitz/goimportcycleviz/internal/modfile"
	"io/ioutil"
	"log"
	"path/filepath"
)

func main() {
	var path string
	var dotFile string
	flag.StringVar(&path, "path", "./", "Files to process")
	flag.StringVar(&dotFile, "dot", "", "DOT file for output")
	flag.Parse()

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	goModFile, err := modfile.FindGoModFile(absPath)
	if err != nil {
		log.Fatal(err)
	}
	modulePath, err := modfile.GetModulePath(goModFile)
	if err != nil {
		log.Fatal(err)
	}

	icd := internal.NewImportGrapher(modulePath)
	filesByFilePath, err := icd.Run(path)
	if err != nil {
		log.Fatal(err)
	}

	files := make([]*internal.File, 0, len(filesByFilePath))
	for _, file := range filesByFilePath {
		files = append(files, file)
	}

	if dotFile != "" {
		output, err := dot.Marshal(files)
		if err != nil {
			log.Fatal(err)
		}

		ioutil.WriteFile(dotFile, output, 0644)
	}

}
