package main

import (
	"context"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/samlitowitz/goimportcycle/internal/config"

	"github.com/samlitowitz/goimportcycle/internal/dot"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"

	"github.com/samlitowitz/goimportcycle/internal/modfile"
)

func walkDirectories(path string, errChan chan<- error) <-chan string {
	dirOut := make(chan string)

	go func() {
		defer close(dirOut)
		err := filepath.WalkDir(
			path,
			func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					return nil
				}
				if strings.HasPrefix(d.Name(), ".") {
					return fs.SkipDir
				}
				if strings.HasPrefix(d.Name(), "_") {
					return fs.SkipDir
				}
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
				dirOut <- path
				return nil
			},
		)
		if err != nil {
			errChan <- err
		}
	}()

	return dirOut
}

func parseFiles(
	dirOut <-chan string,
	errChan chan<- error,
	done <-chan struct{},
) <-chan ast.Node {
	depVis, nodeOut := internalAST.NewDependencyVisitor()

	go func() {
		for {
			select {
			case dirPath, ok := <-dirOut:
				if !ok {
					depVis.Close()
					return
				}
				fset := token.NewFileSet()
				pkgs, err := parser.ParseDir(fset, dirPath, nil, 0)
				if err != nil {
					errChan <- err
				}

				for _, pkg := range pkgs {
					ast.Walk(depVis, pkg)
				}

			case <-done:
				depVis.Close()
				return
			}
		}
	}()
	return nodeOut
}

func detectInputCycles(
	builder *internalAST.PrimitiveBuilder,
	cancel context.CancelFunc,
	nodeOut <-chan ast.Node,
	errChan chan<- error,
	done <-chan struct{},
) error {
	go func() {
		for {
			select {
			case node, ok := <-nodeOut:
				if !ok {
					cancel()
					return
				}
				err := builder.AddNode(node)
				if err != nil {
					errChan <- err
					return
				}

			case <-done:
				return
			}
		}
	}()
	<-done
	return builder.MarkupImportCycles()
}

func main() {
	// TODO: log verbosity
	// TODO: option to show in cycle only
	var dotFile, path, resolution string
	var debug bool
	flag.StringVar(&dotFile, "dot", "", "DOT file for output")
	flag.StringVar(&path, "path", "./", "Files to process")
	flag.StringVar(&resolution, "resolution", "file", "Resolution, 'file' or 'package'")
	flag.BoolVar(&debug, "debug", false, "Emit debug output")
	flag.Parse()

	cfg := config.Default()
	if debug {
		cfg.Debug.SetOutput(os.Stdout)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	goModFile, err := modfile.FindGoModFile(absPath)
	if err != nil {
		log.Fatal(err)
	}
	cfg.Debug.Printf("go.mod file: %s", goModFile)

	modulePath, err := modfile.GetModulePath(goModFile)
	if err != nil {
		log.Fatal(err)
	}
	moduleRootDir := filepath.Dir(goModFile)
	cfg.Debug.Printf("Module Path: %s", modulePath)
	cfg.Debug.Printf("Module Root Directory: %s", moduleRootDir)

	switch resolution {
	case "file":
		cfg.Resolution = config.FileResolution
		if err != nil {
			log.Fatal(err)
		}
	case "package":
		cfg.Resolution = config.PackageResolution
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatal("resolution must be 'file' or 'package'")
	}

	builder := internalAST.NewPrimitiveBuilder(modulePath, moduleRootDir)
	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error)

	go func() {
		for {
			select {
			case err := <-errChan:
				cancel()
				log.Fatal(err)
			case <-ctx.Done():
				return
			}
		}
	}()

	dirOut := walkDirectories(moduleRootDir, errChan)
	nodeOut := parseFiles(dirOut, errChan, ctx.Done())
	err = detectInputCycles(builder, cancel, nodeOut, errChan, ctx.Done())
	close(errChan)
	if err != nil {
		log.Fatal(err)
	}

	output, err := dot.Marshal(cfg, modulePath, builder.Packages())
	if err != nil {
		log.Fatal(err)
	}
	if dotFile == "" {
		_, err := os.Stdout.Write(output)
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	err = os.WriteFile(dotFile, output, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//v := tmp.NewVisitor(modulePath)
	//err = filepath.WalkDir(path, v.WalkDirFn)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//
	//cfg := internal.True
	//var output []byte
	//
	//switch resolution {
	//case "file":
	//	output, err = file.Marshal(cfg, modulePath, v.Packages())
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//case "package":
	//	output, err = pkg.Marshal(internal.True, modulePath, v.Packages())
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//default:
	//	log.Fatal("resolution must be 'file' or 'package'")
	//}
	//
	//if dotFile == "" {
	//	os.Stdout.Write(output)
	//	return
	//}
	//ioutil.WriteFile(dotFile, output, 0644)
}
