package main

import (
	"flag"
	"github.com/samlitowitz/goimportcycleviz/internal"
	modfile2 "github.com/samlitowitz/goimportcycleviz/internal/modfile"
	"log"
	"path/filepath"
)

func main() {
	var path string
	flag.StringVar(&path, "path", "./", "Files to process")
	flag.Parse()

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	goModFile, err := modfile2.FindGoModFile(absPath)
	if err != nil {
		log.Fatal(err)
	}
	modulePath, err := modfile2.GetModulePath(goModFile)
	if err != nil {
		log.Fatal(err)
	}

	icd := internal.NewImportCycleDetector(modulePath)
	packages, err := icd.Run(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(packages)
}
