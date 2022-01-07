package main

import (
	"flag"
	"github.com/samlitowitz/goimportcycle/internal/ast"
	"github.com/samlitowitz/goimportcycle/internal/file"
	"github.com/samlitowitz/goimportcycle/internal/pkg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/samlitowitz/goimportcycle/internal/modfile"
)

func main() {
	var dotFile, path, resolution string
	flag.StringVar(&dotFile, "dot", "", "DOT file for output")
	flag.StringVar(&path, "path", "./", "Files to process")
	flag.StringVar(&resolution, "resolution", "file", "Resolution, 'file' or 'package'")
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

	v := ast.NewVisitor(modulePath)
	err = filepath.WalkDir(path, v.WalkDirFn)
	if err != nil {
		log.Fatal(err)
	}

	var output []byte

	switch resolution {
	case "file":
		output, err = file.Marshal(modulePath, v.Packages())
		if err != nil {
			log.Fatal(err)
		}
	case "package":
		output, err = pkg.Marshal(modulePath, v.Packages())
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("resolution must be 'file' or 'package'")
	}

	if dotFile == "" {
		os.Stdout.Write(output)
		return
	}
	ioutil.WriteFile(dotFile, output, 0644)
}
