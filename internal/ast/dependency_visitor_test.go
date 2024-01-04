package ast_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"
)

func TestDependencyVisitor_Visit_EmitsPackages(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main",
				[]*Node{
					{
						"main.go",
						nil,
						"main",
						"",
					},
				},
				"",
				"",
			},
			{
				"no_main",
				[]*Node{
					{
						"no_main.go",
						nil,
						"no_main",
						"",
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								"a",
								"",
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										"b",
										"",
									},
								},
								"",
								"",
							},
						},
						"",
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								"c",
								"",
							},
						},
						"",
						"",
					},
				},
				"",
				"",
			},
			{
				"both",
				[]*Node{
					{
						"main.go",
						nil,
						"main",
						"",
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								"a",
								"",
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										"b",
										"",
									},
								},
								"",
								"",
							},
						},
						"",
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								"c",
								"",
							},
						},
						"",
						"",
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			// don't update directories
			if n.entries != nil {
				return
			}
			// set data for files
			n.data = "package " + n.pkg + "\n\n"
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()

		expectedPackageNamesInOrder := []string{}
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					expectedPackageNamesInOrder = append(expectedPackageNamesInOrder, n.pkg)
					return
				}

				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(filepath.Separator)+"testdata"+string(filepath.Separator)+path,
				)
			},
		)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for _, dirPath := range directoryPathsInOrder {
				dirOut <- dirPath
			}
			close(dirOut)
		}()

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
						cancel()
						t.Fatalf("%s: %s", testCase, err)
					}

					for _, pkg := range pkgs {
						ast.Walk(depVis, pkg)
					}

				case <-ctx.Done():
					depVis.Close()
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case astNode, ok := <-nodeOut:
					if !ok {
						cancel()
						return
					}
					switch astNode := astNode.(type) {
					case *internalAST.Package:
						if len(expectedPackageNamesInOrder) == 0 {
							cancel()
							t.Errorf(
								"%s: read more packages than expected: %s",
								testCase,
								astNode.Name,
							)
						}
						if expectedPackageNamesInOrder[0] != astNode.Name {
							cancel()
							t.Errorf(
								"%s: expected package %s, got %s",
								testCase,
								expectedPackageNamesInOrder[0],
								astNode.Name,
							)
						}
						expectedPackageNamesInOrder = expectedPackageNamesInOrder[1:]
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		if len(expectedPackageNamesInOrder) != 0 {
			t.Errorf(
				"%s: expected package names never received: %s",
				testCase,
				strings.Join(expectedPackageNamesInOrder, ", "),
			)
		}
	}
}

func TestDependencyVisitor_Visit_EmitsFiles(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main",
				[]*Node{
					{
						"main.go",
						nil,
						"main",
						"",
					},
				},
				"",
				"",
			},
			{
				"no_main",
				[]*Node{
					{
						"no_main.go",
						nil,
						"no_main",
						"",
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								"a",
								"",
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										"b",
										"",
									},
								},
								"",
								"",
							},
						},
						"",
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								"c",
								"",
							},
						},
						"",
						"",
					},
				},
				"",
				"",
			},
			{
				"both",
				[]*Node{
					{
						"main.go",
						nil,
						"main",
						"",
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								"a",
								"",
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										"b",
										"",
									},
								},
								"",
								"",
							},
						},
						"",
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								"c",
								"",
							},
						},
						"",
						"",
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			// don't update directories
			if n.entries != nil {
				return
			}
			// set data for files
			n.data = "package " + n.pkg + "\n\n"
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()

		expectedFilenamesInOrder := []string{}
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					expectedFilenamesInOrder = append(
						expectedFilenamesInOrder,
						tmpDir+string(filepath.Separator)+"testdata"+string(filepath.Separator)+path,
					)
					return
				}

				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(filepath.Separator)+"testdata"+string(filepath.Separator)+path,
				)
			},
		)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			for _, dirPath := range directoryPathsInOrder {
				dirOut <- dirPath
			}
			close(dirOut)
		}()

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
						cancel()
						t.Fatalf("%s: %s", testCase, err)
					}

					for _, pkg := range pkgs {
						ast.Walk(depVis, pkg)
					}

				case <-ctx.Done():
					depVis.Close()
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case node, ok := <-nodeOut:
					if !ok {
						cancel()
						return
					}
					switch node := node.(type) {
					case *internalAST.File:
						if len(expectedFilenamesInOrder) == 0 {
							cancel()
							t.Errorf(
								"%s: read more files than expected: %s",
								testCase,
								node.AbsPath,
							)
						}
						if expectedFilenamesInOrder[0] != node.AbsPath {
							cancel()
							t.Errorf(
								"%s: expected file path %s, got %s",
								testCase,
								expectedFilenamesInOrder[0],
								node.AbsPath,
							)
						}
						expectedFilenamesInOrder = expectedFilenamesInOrder[1:]
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		if len(expectedFilenamesInOrder) != 0 {
			t.Errorf(
				"%s: expected package names never received: %s",
				testCase,
				strings.Join(expectedFilenamesInOrder, ", "),
			)
		}
	}
}

func TestDependencyVisitor_Visit_EmitsImportSpecs(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main.go",
				nil,
				"main",
				`
package main

import (
	"fmt"
	aliasFmt "fmt"
)`,
			},
			{
				"a",
				[]*Node{
					{
						"a.go",
						nil,
						"a",
						`
package a

type A struct {}

func (a A) Error() string {
	return "A"
}
`,
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	makeTree(t, tree)

	testCase := ""
	dirOut := make(chan string)
	depVis, nodeOut := internalAST.NewDependencyVisitor()

	expectedImportPathsInOrder := []string{
		"fmt",
		"fmt",
	}
	directoryPathsInOrder := []string{}
	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			if n.entries == nil {
				return
			}

			directoryPathsInOrder = append(directoryPathsInOrder, tmpDir+string(filepath.Separator)+path)
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for _, dirPath := range directoryPathsInOrder {
			dirOut <- dirPath
		}
		close(dirOut)
	}()

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
					cancel()
					t.Fatalf("%s: %s", testCase, err)
				}

				for _, pkg := range pkgs {
					ast.Walk(depVis, pkg)
				}

			case <-ctx.Done():
				depVis.Close()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case node, ok := <-nodeOut:
				if !ok {
					cancel()
					return
				}
				switch node := node.(type) {
				case *ast.ImportSpec:
					if len(expectedImportPathsInOrder) == 0 {
						cancel()
						t.Errorf(
							"%s: read more import paths than expected: %s",
							testCase,
							node.Name.String(),
						)
					}
					if expectedImportPathsInOrder[0] != node.Path.Value {
						cancel()
						t.Errorf(
							"%s: expected import path %s, got %s",
							testCase,
							expectedImportPathsInOrder[0],
							node.Path.Value,
						)
					}
					expectedImportPathsInOrder = expectedImportPathsInOrder[1:]
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()

	if len(expectedImportPathsInOrder) != 0 {
		t.Errorf(
			"%s: expected package names never received: %s",
			testCase,
			strings.Join(expectedImportPathsInOrder, ", "),
		)
	}
}

func TestDependencyVisitor_Visit_EmitsConstantsTypesAndVars(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main.go",
				nil,
				"main",
				`
package main

const CA = 1
const CB = 2
const (
	CC = 3
	CD = 4
)
const ce = 5

var VA = 1
var VB = 2
var (
	VC = 3
	VD = 4
)
var ve = 5

type TA struct {}
type TB struct {}
type tc struct{}
`,
			},
			{
				"a",
				[]*Node{
					{
						"a.go",
						nil,
						"a",
						`
package a

type A struct {}
`,
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	makeTree(t, tree)

	testCase := ""
	dirOut := make(chan string)
	depVis, nodeOut := internalAST.NewDependencyVisitor()

	expectedNamesInOrder := []string{
		"CA", "CB", "CC", "CD", "ce",
		"VA", "VB", "VC", "VD", "ve",
		"TA", "TB", "tc",
		"A",
	}
	directoryPathsInOrder := []string{}
	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			if n.entries == nil {
				return
			}

			directoryPathsInOrder = append(directoryPathsInOrder, tmpDir+string(filepath.Separator)+path)
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for _, dirPath := range directoryPathsInOrder {
			dirOut <- dirPath
		}
		close(dirOut)
	}()

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
					cancel()
					t.Fatalf("%s: %s", testCase, err)
				}

				for _, pkg := range pkgs {
					ast.Walk(depVis, pkg)
				}

			case <-ctx.Done():
				depVis.Close()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case node, ok := <-nodeOut:
				if !ok {
					cancel()
					return
				}
				switch node := node.(type) {
				case *ast.GenDecl:
					for _, spec := range node.Specs {
						switch node.Tok {
						case token.CONST:
							spec, ok := spec.(*ast.ValueSpec)
							if !ok {
								continue
							}
							for _, ident := range spec.Names {
								if len(expectedNamesInOrder) == 0 {
									cancel()
									t.Errorf(
										"%s: read more general declarations than expected: %s",
										testCase,
										ident.String(),
									)
									return
								}
								if expectedNamesInOrder[0] != ident.String() {
									cancel()
									t.Errorf(
										"%s: expected const %s, got %s",
										testCase,
										expectedNamesInOrder[0],
										ident.String(),
									)
								}
							}
						case token.TYPE:
							spec, ok := spec.(*ast.TypeSpec)
							if !ok {
								continue
							}
							if len(expectedNamesInOrder) == 0 {
								cancel()
								t.Errorf(
									"%s: read more general declarations than expected: %s",
									testCase,
									spec.Name.String(),
								)
								return
							}
							if expectedNamesInOrder[0] != spec.Name.String() {
								cancel()
								t.Errorf(
									"%s: expected type %s, got %s",
									testCase,
									expectedNamesInOrder[0],
									spec.Name.String(),
								)
							}
						case token.VAR:
							spec, ok := spec.(*ast.ValueSpec)
							if !ok {
								continue
							}
							for _, ident := range spec.Names {
								if len(expectedNamesInOrder) == 0 {
									cancel()
									t.Errorf(
										"%s: read more general declarations than expected: %s",
										testCase,
										ident.String(),
									)
									return
								}
								if expectedNamesInOrder[0] != ident.String() {
									cancel()
									t.Errorf(
										"%s: expected const %s, got %s",
										testCase,
										expectedNamesInOrder[0],
										ident.String(),
									)
								}
							}
						}
						expectedNamesInOrder = expectedNamesInOrder[1:]
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()

	if len(expectedNamesInOrder) != 0 {
		t.Errorf(
			"%s: expected names never received: %s",
			testCase,
			strings.Join(expectedNamesInOrder, ", "),
		)
	}
}

func TestDependencyVisitor_Visit_EmitsFunctions(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main.go",
				nil,
				"main",
				`
package main

type TA struct {}

func (r TA) FTA1() {}

func (r *TA) FTA2() {}

type tc struct{}

func (r tc) FC() {}

func F1() {}
`,
			},
			{
				"a",
				[]*Node{
					{
						"a.go",
						nil,
						"a",
						`
package a

type A struct {}

func (a A) FA() {}
`,
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	makeTree(t, tree)

	testCase := ""
	dirOut := make(chan string)
	depVis, nodeOut := internalAST.NewDependencyVisitor()

	expectedNamesInOrder := []string{
		"FTA1", "FTA2",
		"FC",
		"F1",
		"FA",
	}
	directoryPathsInOrder := []string{}
	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			if n.entries == nil {
				return
			}

			directoryPathsInOrder = append(directoryPathsInOrder, tmpDir+string(filepath.Separator)+path)
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for _, dirPath := range directoryPathsInOrder {
			dirOut <- dirPath
		}
		close(dirOut)
	}()

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
					cancel()
					t.Fatalf("%s: %s", testCase, err)
				}

				for _, pkg := range pkgs {
					ast.Walk(depVis, pkg)
				}

			case <-ctx.Done():
				depVis.Close()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case node, ok := <-nodeOut:
				if !ok {
					cancel()
					return
				}
				switch node := node.(type) {
				case *internalAST.FuncDecl:
					if len(expectedNamesInOrder) == 0 {
						cancel()
						t.Errorf(
							"%s: read more functions than expected: %s",
							testCase,
							node.Name.String(),
						)
						return
					}
					if expectedNamesInOrder[0] != node.Name.String() {
						cancel()
						t.Errorf(
							"%s: expected function named %s, got %s",
							testCase,
							expectedNamesInOrder[0],
							node.Name.String(),
						)
					}
					expectedNamesInOrder = expectedNamesInOrder[1:]
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()

	if len(expectedNamesInOrder) != 0 {
		t.Errorf(
			"%s: expected function names never received: %s",
			testCase,
			strings.Join(expectedNamesInOrder, ", "),
		)
	}
}

func TestDependencyVisitor_Visit_EmitsSelectorExprs(t *testing.T) {
	// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L600
	// -- START -- //
	if runtime.GOOS == "ios" {
		restore := chtmpdir(t)
		defer restore()
	}

	tmpDir := t.TempDir()

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal("finding working dir:", err)
	}
	if err = os.Chdir(tmpDir); err != nil {
		t.Fatal("entering temp dir:", err)
	}
	defer os.Chdir(origDir)
	// -- END -- //

	tree := &Node{
		"testdata",
		[]*Node{
			{
				"main.go",
				nil,
				"main",
				`
package main

import "fmt"

func f1() {}

func main() {
	f1()
	fmt.Println("Hello world!")
}
`,
			},
			{
				"a",
				[]*Node{
					{
						"a.go",
						nil,
						"a",
						`
package a

import "fmt"

var println = fmt.Println

func init() {
	println("A hello world!")
}
`,
					},
				},
				"",
				"",
			},
		},
		"",
		"",
	}

	makeTree(t, tree)

	testCase := ""
	dirOut := make(chan string)
	depVis, nodeOut := internalAST.NewDependencyVisitor()

	expectedNamesInOrder := []string{
		"fmt.Println",
		"fmt.Println",
	}
	directoryPathsInOrder := []string{}
	walkTree(
		tree,
		tree.name,
		func(path string, n *Node) {
			if n.entries == nil {
				return
			}

			directoryPathsInOrder = append(directoryPathsInOrder, tmpDir+string(filepath.Separator)+path)
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for _, dirPath := range directoryPathsInOrder {
			dirOut <- dirPath
		}
		close(dirOut)
	}()

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
					cancel()
					t.Fatalf("%s: %s", testCase, err)
				}

				for _, pkg := range pkgs {
					ast.Walk(depVis, pkg)
				}

			case <-ctx.Done():
				depVis.Close()
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case node, ok := <-nodeOut:
				if !ok {
					cancel()
					return
				}
				switch node := node.(type) {
				case *ast.SelectorExpr:
					x, ok := node.X.(*ast.Ident)
					if !ok {
						cancel()
						t.Errorf("%s: expected selector node to have X", testCase)
					}
					actual := x.String() + "." + node.Sel.String()

					if len(expectedNamesInOrder) == 0 {
						cancel()
						t.Errorf(
							"%s: read more functions than expected: %s",
							testCase,
							actual,
						)
						return
					}
					if expectedNamesInOrder[0] != actual {
						cancel()
						t.Errorf(
							"%s: expected function named %s, got %s",
							testCase,
							expectedNamesInOrder[0],
							actual,
						)
					}
					expectedNamesInOrder = expectedNamesInOrder[1:]
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()

	if len(expectedNamesInOrder) != 0 {
		t.Errorf(
			"%s: expected function names never received: %s",
			testCase,
			strings.Join(expectedNamesInOrder, ", "),
		)
	}
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L449
type Node struct {
	name    string
	entries []*Node // nil if the entry is a file
	pkg     string  // file package belongs to, empty if entry is a directory
	data    string  // file content, empty if entry is a directory
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L481
func walkTree(n *Node, path string, f func(path string, n *Node)) {
	f(path, n)
	for _, e := range n.entries {
		walkTree(e, filepath.Join(path, e.name), f)
	}
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L488
func makeTree(t *testing.T, tree *Node) map[string]struct{} {
	directories := make(map[string]struct{})
	walkTree(tree, tree.name, func(path string, n *Node) {
		if n.entries == nil {
			fd, err := os.Create(path)
			if err != nil {
				t.Errorf("makeTree: %v", err)
				return
			}
			if n.data != "" {
				_, err = fd.Write([]byte(n.data))
				if err != nil {
					t.Errorf("makeTree: %v", err)
					return
				}
			}
			fd.Close()
		} else {
			os.Mkdir(path, 0770)
			directories[path] = struct{}{}
		}
	})
	return directories
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L553
func chtmpdir(t *testing.T) (restore func()) {
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	d, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	if err := os.Chdir(d); err != nil {
		t.Fatalf("chtmpdir: %v", err)
	}
	return func() {
		if err := os.Chdir(oldwd); err != nil {
			t.Fatalf("chtmpdir: %v", err)
		}
		os.RemoveAll(d)
	}
}
