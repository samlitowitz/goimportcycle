package ast_test

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"runtime"
	"testing"

	"github.com/samlitowitz/goimportcycle/internal"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"
)

func TestPrimitiveBuilder_MarkupImportCycles_None(t *testing.T) {
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
			// none: a -> b -> c
			{
				"none",
				[]*Node{
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								"a",
								`
package a

import "example.com/none/b"

func init() {
	b.Fn()
}
`,
							},
						},
						"",
						"",
					},
					{
						"b",
						[]*Node{
							{
								"b.go",
								nil,
								"b",
								`
package b

import "example.com/none/c"

func Fb() {
	c.Fn()
}
`,
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
								`
package c

func Fn() { }
`,
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
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("example.com", tmpDir+string(os.PathSeparator)+"testdata")

		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		err := builder.MarkupImportCycles()
		if err != nil {
			t.Fatalf(
				"%s: error marking up import cycles: %s",
				testCase,
				err,
			)
		}

		for _, file := range builder.Files() {
			if file.InImportCycle {
				t.Errorf(
					"%s: unexpected import cycle: %s",
					testCase,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_MarkupImportCycles_Simple(t *testing.T) {
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
			// simple: a -> b -> a
			{
				"simple",
				[]*Node{
					{
						"notincycle",
						[]*Node{
							{
								"notincycle.go",
								nil,
								"notincycle",
								`
package notincycle

func NICFn() { }
`,
							},
						},
						"",
						"",
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

import "example.com/simple/b"
import "example.com/simple/notincycle"

func AFn() {
	b.BFn()
	notincycle.NICFn()
}
`,
							},
						},
						"",
						"",
					},
					{
						"b",
						[]*Node{
							{
								"b.go",
								nil,
								"b",
								`
package b

import "example.com/simple/a"

func BFn() {
	a.AFn()
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
			},
		},
		"",
		"",
	}
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("example.com", tmpDir+string(os.PathSeparator)+"testdata")

		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		err := builder.MarkupImportCycles()
		if err != nil {
			t.Fatalf(
				"%s: error marking up import cycles: %s",
				testCase,
				err,
			)
		}

		for _, file := range builder.Files() {
			if file.Package.Name == "notincycle" {
				continue
			}
			if !file.Package.InImportCycle {
				t.Errorf(
					"%s: %s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
					file.Package.Name,
				)
			}
			if !file.InImportCycle {
				t.Errorf(
					"%s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_MarkupImportCycles_Transitive(t *testing.T) {
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
			// transitive: a -> b -> c -> a
			{
				"transitive",
				[]*Node{
					{
						"notincycle",
						[]*Node{
							{
								"notincycle.go",
								nil,
								"notincycle",
								`
package notincycle

func NICFn() { }
`,
							},
						},
						"",
						"",
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

import "example.com/transitive/b"

func AFn() {
	b.BFn()
}
`,
							},
						},
						"",
						"",
					},
					{
						"b",
						[]*Node{
							{
								"b.go",
								nil,
								"b",
								`
package b

import "example.com/transitive/c"

func BFn() {
	c.CFn()
}
`,
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
								`
package c

import "example.com/transitive/a"

func CFn() {
	a.AFn()
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
			},
		},
		"",
		"",
	}
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("example.com", tmpDir+string(os.PathSeparator)+"testdata")

		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		err := builder.MarkupImportCycles()
		if err != nil {
			t.Fatalf(
				"%s: error marking up import cycles: %s",
				testCase,
				err,
			)
		}

		for _, file := range builder.Files() {
			if file.Package.Name == "notincycle" {
				continue
			}
			if !file.Package.InImportCycle {
				t.Errorf(
					"%s: %s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
					file.Package.Name,
				)
			}
			if !file.InImportCycle {
				t.Errorf(
					"%s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_MarkupImportCycles_MultipleIndependent(t *testing.T) {
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
			// multiple independent: a -> b -> a, a -> c -> a
			{
				"multiple_independent",
				[]*Node{
					{
						"notincycle",
						[]*Node{
							{
								"notincycle.go",
								nil,
								"notincycle",
								`
package notincycle

func NICFn() { }
`,
							},
						},
						"",
						"",
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

import "example.com/multiple_independent/b"
import "example.com/multiple_independent/c"

func AFn() {
	b.BFn()
	c.CFn()
}
`,
							},
						},
						"",
						"",
					},
					{
						"b",
						[]*Node{
							{
								"b.go",
								nil,
								"b",
								`
package b

import "example.com/multiple_independent/a"

func BFn() {
	a.AFn()
}
`,
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
								`
package c

import "example.com/multiple_independent/a"

func CFn() {
	a.AFn()
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
			},
		},
		"",
		"",
	}
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("example.com", tmpDir+string(os.PathSeparator)+"testdata")

		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		err := builder.MarkupImportCycles()
		if err != nil {
			t.Fatalf(
				"%s: error marking up import cycles: %s",
				testCase,
				err,
			)
		}

		for _, file := range builder.Files() {
			if file.Package.Name == "notincycle" {
				continue
			}
			if !file.Package.InImportCycle {
				t.Errorf(
					"%s: %s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
					file.Package.Name,
				)
			}
			if !file.InImportCycle {
				t.Errorf(
					"%s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_MarkupImportCycles_MultipleInterlinked(t *testing.T) {
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
			// multiple interlinked: a -> b -> a, b -> c -> b
			{
				"multiple_interlinked",
				[]*Node{
					{
						"notincycle",
						[]*Node{
							{
								"notincycle.go",
								nil,
								"notincycle",
								`
package notincycle

func NICFn() { }
`,
							},
						},
						"",
						"",
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

import "example.com/multiple_interlinked/b"

func AFn() {
	b.BFn()
}
`,
							},
						},
						"",
						"",
					},
					{
						"b",
						[]*Node{
							{
								"b.go",
								nil,
								"b",
								`
package b

import "example.com/multiple_interlinked/a"
import "example.com/multiple_interlinked/c"

func BFn() {
	a.AFn()
	c.CFn()
}
`,
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
								`
package c

import "example.com/multiple_interlinked/b"

func CFn() {
	b.BFn()
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
			},
		},
		"",
		"",
	}
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("example.com", tmpDir+string(os.PathSeparator)+"testdata")

		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		err := builder.MarkupImportCycles()
		if err != nil {
			t.Fatalf(
				"%s: error marking up import cycles: %s",
				testCase,
				err,
			)
		}

		for _, file := range builder.Files() {
			if file.Package.Name == "notincycle" {
				continue
			}
			if !file.Package.InImportCycle {
				t.Errorf(
					"%s: %s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
					file.Package.Name,
				)
			}
			if !file.InImportCycle {
				t.Errorf(
					"%s: %s: not in expected import cycle",
					testCase,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_Packages(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedPackagesByDirName := make(map[string]*internal.Package, 0)
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				pkg := &internal.Package{
					DirName: tmpDir + string(os.PathSeparator) + "testdata" + string(os.PathSeparator) + path,
					Name:    n.pkg,
					Files:   nil,
				}
				expectedPackagesByDirName[pkg.DirName] = pkg

				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, pkg := range builder.Packages() {
			if _, ok := expectedPackagesByDirName[pkg.DirName]; !ok {
				t.Errorf(
					"%s: unexpected package: %s \"%s\"",
					testCase,
					pkg.Name,
					pkg.DirName,
				)
			}
			delete(expectedPackagesByDirName, pkg.DirName)
		}

		if len(expectedPackagesByDirName) != 0 {
			for _, pkg := range expectedPackagesByDirName {
				t.Errorf(
					"%s: missing expected package: %s \"%s\"",
					testCase,
					pkg.Name,
					pkg.DirName,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_Files(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedFilesByDirName := make(map[string]*internal.File, 0)
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					file := &internal.File{
						Package:  nil,
						FileName: n.name,
						AbsPath:  tmpDir + string(os.PathSeparator) + "testdata" + string(os.PathSeparator) + path,
						Imports:  nil,
						Decls:    nil,
					}
					expectedFilesByDirName[file.AbsPath] = file
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			if _, ok := expectedFilesByDirName[file.AbsPath]; !ok {
				t.Errorf(
					"%s: unexpected file: %s \"%s\"",
					testCase,
					file.FileName,
					file.AbsPath,
				)
			}
			delete(expectedFilesByDirName, file.AbsPath)
		}

		if len(expectedFilesByDirName) != 0 {
			for _, file := range expectedFilesByDirName {
				t.Errorf(
					"%s: missing expected file: %s \"%s\"",
					testCase,
					file.FileName,
					file.AbsPath,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_Imports(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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

			imports := map[string]string{
				"net/http":      "",
				"fmt":           "",
				n.pkg + "Alias": "fmt",
			}
			importStr := ""
			for imp, path := range imports {
				if path == "" {
					importStr += fmt.Sprintf("\t\"%s\"\n", imp)
					continue
				}
				importStr += fmt.Sprintf("\t%s \"%s\"\n", imp, path)
			}
			// set data for files
			n.data = fmt.Sprintf(
				`
package %s

import (
	%s
)

func init() {
	fmt.Println("Test")
	%sAlias.Println("Test")
}
`,
				n.pkg,
				importStr,
				n.pkg,
			)
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedImportSpecs := []*internal.Import{}
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					impHttp := &internal.Import{
						Package:         nil,
						Name:            "http",
						ReferencedTypes: nil,
					}

					impFmt := &internal.Import{
						Package:         nil,
						Name:            "fmt",
						ReferencedTypes: nil,
					}

					impAlias := &internal.Import{
						Package:         nil,
						Name:            n.pkg + "Alias",
						ReferencedTypes: nil,
					}

					expectedImportSpecs = append(
						expectedImportSpecs,
						impHttp,
						impFmt,
						impAlias,
					)
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			for _, imp := range file.Imports {
				if len(expectedImportSpecs) == 0 {
					t.Errorf(
						"%s: unexpected import: %s",
						testCase,
						imp.Name,
					)
				}

				found := false
				for i, expectedImp := range expectedImportSpecs {
					if expectedImp.Name != imp.Name {
						continue
					}
					found = true
					expectedImportSpecs = append(expectedImportSpecs[:i], expectedImportSpecs[i+1:]...)
					break
				}
				if !found {
					t.Errorf(
						"%s: unexpected import: %s",
						testCase,
						imp.Name,
					)
				}

			}
		}

		if len(expectedImportSpecs) != 0 {
			for _, imp := range expectedImportSpecs {
				t.Errorf(
					"%s: missing expected import: %s",
					testCase,
					imp.Name,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_Imports_StubFixups(t *testing.T) {
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
				"both",
				[]*Node{
					{
						"main.go",
						nil,
						"main",
						`
package main

import "testdata/both/a"
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

import (
	"testdata/both/a/b"
	aliasB "testdata/both/a/b"
)

func init() {
	b.BFn()
	aliasB.BFn()
}

func AFn() { }
`,
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										"b",
										`
package b

import "testdata/both/c"

func init() {
	c.CFn()
}

func BFn() { }
`,
									},
								},
								"b",
								"",
							},
						},
						"a",
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								"c",
								`
package c

import (
	"testdata/both/a"
	"testdata/both/a/b"
)

func init() {
	a.AFn()
	b.BFn()
}

func CFn() { }
`,
							},
						},
						"c",
						"",
					},
				},
				"both",
				"",
			},
		},
		"",
		"",
	}

	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("testdata", tmpDir+string(os.PathSeparator)+"testdata")

		expectedImportSpecs := []*internal.Import{
			{
				Name: "a",
				Path: "testdata/both/a",
			},
			{
				Name: "b",
				Path: "testdata/both/a/b",
			},
			{
				Name: "aliasB",
				Path: "testdata/both/a/b",
			},
			{
				Name: "c",
				Path: "testdata/both/c",
			},
			{
				Name: "a",
				Path: "testdata/both/a",
			},
			{
				Name: "b",
				Path: "testdata/both/a/b",
			},
		}
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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

		go func(t *testing.T) {
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
		}(t)

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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			if file.IsStub {
				t.Errorf(
					"%s: unexpected stub file: %s",
					testCase,
					file.AbsPath,
				)
				for _, decl := range file.Decls {
					t.Errorf(
						"%s: unexpected declaration in stub file: %s",
						testCase,
						decl.UID(),
					)
				}
			}
			for _, imp := range file.Imports {
				if len(expectedImportSpecs) == 0 {
					t.Errorf(
						"%s: unexpected import: %s",
						testCase,
						imp.Name,
					)
				}

				found := false
				for i, expectedImp := range expectedImportSpecs {
					if expectedImp.Name != imp.Name {
						continue
					}
					if expectedImp.Path != imp.Path {
						continue
					}
					found = true
					expectedImportSpecs = append(expectedImportSpecs[:i], expectedImportSpecs[i+1:]...)
					break
				}
				if !found {
					t.Errorf(
						"%s: unexpected import: %s",
						testCase,
						imp.Name,
					)
				}
				for _, refDecl := range imp.ReferencedTypes {
					if refDecl.File == nil {
						t.Errorf(
							"%s: referenced declaration without file assigned: %s",
							testCase,
							refDecl.UID(),
						)
						continue
					}
					if refDecl.File.IsStub {
						t.Errorf(
							"%s: %s: referenced declaration defined in stub: %s: %s",
							testCase,
							file.UID(),
							refDecl.UID(),
							refDecl.File.UID(),
						)
					}
				}

			}
		}

		if len(expectedImportSpecs) != 0 {
			for _, imp := range expectedImportSpecs {
				t.Errorf(
					"%s: missing expected import: %s",
					testCase,
					imp.Name,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_FunctionDeclarations(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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
			n.data = fmt.Sprintf(
				`
package %s

func %sFn() { }

type A struct { }

func (a *A) AFn() { }

func (b *B) BFn() { }

type B struct { }
`,
				n.pkg,
				n.pkg,
			)
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedFuncsByFileName := make(map[string]map[string]*internal.Decl, 0)
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					if _, ok := expectedFuncsByFileName[n.name]; !ok {
						expectedFuncsByFileName[n.name] = make(map[string]*internal.Decl, 1)
					}
					declPkgFn := &internal.Decl{
						File: nil,
						Name: n.pkg + "Fn",
					}
					declAFn := &internal.Decl{
						File: nil,
						Name: "AFn",
					}
					declBFn := &internal.Decl{
						File: nil,
						Name: "BFn",
					}
					expectedFuncsByFileName[n.name][declPkgFn.Name] = declPkgFn
					expectedFuncsByFileName[n.name][declAFn.Name] = declAFn
					expectedFuncsByFileName[n.name][declBFn.Name] = declBFn
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			if _, ok := expectedFuncsByFileName[file.FileName]; !ok {
				t.Errorf(
					"%s: unexpected file: %s \"%s\"",
					testCase,
					file.FileName,
					file.AbsPath,
				)
				continue
			}
			for _, decl := range file.Decls {
				if _, ok := expectedFuncsByFileName[file.FileName][decl.Name]; !ok {
					t.Errorf(
						"%s: unexpected decl: %s in %s",
						testCase,
						decl.Name,
						file.FileName,
					)
					continue
				}
				delete(expectedFuncsByFileName[file.FileName], decl.Name)
			}
			if len(expectedFuncsByFileName[file.FileName]) != 0 {
				for _, decl := range expectedFuncsByFileName[file.FileName] {
					t.Errorf(
						"%s: missing expected decls: %s in %s",
						testCase,
						decl.Name,
						file.FileName,
					)
				}
			}
			delete(expectedFuncsByFileName, file.FileName)
		}

		if len(expectedFuncsByFileName) != 0 {
			for fileName := range expectedFuncsByFileName {
				t.Errorf(
					"%s: missing expected file: %s",
					testCase,
					fileName,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_GeneralDeclarations(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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
			n.data = fmt.Sprintf(
				`
package %s

const %s_CONST = 1
var %s_VAR = 2

type %s_TYPE struct {}
`,
				n.pkg,
				n.pkg,
				n.pkg,
				n.pkg,
			)
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedDeclsByFileName := make(map[string]map[string]*internal.Decl, 0)
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					if _, ok := expectedDeclsByFileName[n.name]; !ok {
						expectedDeclsByFileName[n.name] = make(map[string]*internal.Decl, 1)
					}
					conDecl := &internal.Decl{
						File: nil,
						Name: n.pkg + "_CONST",
					}
					varDecl := &internal.Decl{
						File: nil,
						Name: n.pkg + "_VAR",
					}
					typDecl := &internal.Decl{
						File: nil,
						Name: n.pkg + "_TYPE",
					}
					expectedDeclsByFileName[n.name][conDecl.Name] = conDecl
					expectedDeclsByFileName[n.name][varDecl.Name] = varDecl
					expectedDeclsByFileName[n.name][typDecl.Name] = typDecl
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			if _, ok := expectedDeclsByFileName[file.FileName]; !ok {
				t.Errorf(
					"%s: unexpected file: %s \"%s\"",
					testCase,
					file.FileName,
					file.AbsPath,
				)
				continue
			}
			for _, decl := range file.Decls {
				if _, ok := expectedDeclsByFileName[file.FileName][decl.Name]; !ok {
					t.Errorf(
						"%s: unexpected decl: %s in %s",
						testCase,
						decl.Name,
						file.FileName,
					)
					continue
				}
				delete(expectedDeclsByFileName[file.FileName], decl.Name)
			}
			if len(expectedDeclsByFileName[file.FileName]) != 0 {
				for _, decl := range expectedDeclsByFileName[file.FileName] {
					t.Errorf(
						"%s: missing expected decls: %s in %s",
						testCase,
						decl.Name,
						file.FileName,
					)
				}
			}
			delete(expectedDeclsByFileName, file.FileName)
		}

		if len(expectedDeclsByFileName) != 0 {
			for fileName := range expectedDeclsByFileName {
				t.Errorf(
					"%s: missing expected file: %s",
					testCase,
					fileName,
				)
			}
		}
	}
}

func TestPrimitiveBuilder_AddNode_SelectExpressions(t *testing.T) {
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
				"main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"no_main",
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
								"b",
								"",
							},
						},
						"a",
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
						"c",
						"",
					},
				},
				"both",
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
			n.data = fmt.Sprintf(
				`
package %s

import (
	"fmt"
	%sFmt "fmt"
	"log"
	"os"
)

func init() {
	fmt.Println("Hello")
	%sFmt.Println("Jello")
	file, err := os.Open("file.go") // For read access.
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}
`,
				n.pkg,
				n.pkg,
				n.pkg,
			)
		},
	)
	makeTree(t, tree)

	for _, treeNode := range tree.entries {
		testCase := treeNode.name
		dirOut := make(chan string)
		depVis, nodeOut := internalAST.NewDependencyVisitor()
		builder := internalAST.NewPrimitiveBuilder("", tmpDir)

		expectedImportRefsByFileName := make(map[string]map[string]map[string]*internal.Decl, 0)
		directoryPathsInOrder := []string{}
		walkTree(
			treeNode,
			treeNode.name,
			func(path string, n *Node) {
				if n.entries == nil {
					if _, ok := expectedImportRefsByFileName[n.name]; !ok {
						expectedImportRefsByFileName[n.name] = make(map[string]map[string]*internal.Decl, 1)
					}
					fmtImpName := "fmt"
					if _, ok := expectedImportRefsByFileName[n.name][fmtImpName]; !ok {
						expectedImportRefsByFileName[n.name][fmtImpName] = make(map[string]*internal.Decl, 1)
					}
					pkgAlias := n.pkg + "Fmt"
					if _, ok := expectedImportRefsByFileName[n.name][pkgAlias]; !ok {
						expectedImportRefsByFileName[n.name][pkgAlias] = make(map[string]*internal.Decl, 1)
					}
					osImpName := "os"
					if _, ok := expectedImportRefsByFileName[n.name][osImpName]; !ok {
						expectedImportRefsByFileName[n.name][osImpName] = make(map[string]*internal.Decl, 1)
					}
					logImpName := "log"
					if _, ok := expectedImportRefsByFileName[n.name][logImpName]; !ok {
						expectedImportRefsByFileName[n.name][logImpName] = make(map[string]*internal.Decl, 1)
					}

					unaliased := &internal.Decl{
						Name: "Println",
					}
					aliased := &internal.Decl{
						Name: "Println",
					}
					osOpen := &internal.Decl{
						Name: "Open",
					}
					logFatal := &internal.Decl{
						Name: "Fatal",
					}
					expectedImportRefsByFileName[n.name][fmtImpName][unaliased.Name] = unaliased
					expectedImportRefsByFileName[n.name][pkgAlias][aliased.Name] = aliased
					expectedImportRefsByFileName[n.name][osImpName][osOpen.Name] = osOpen
					expectedImportRefsByFileName[n.name][logImpName][logFatal.Name] = logFatal
					return
				}
				directoryPathsInOrder = append(
					directoryPathsInOrder,
					tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+path,
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
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.File:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.ImportSpec:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.FuncDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *ast.GenDecl:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}

					case *internalAST.SelectorExpr:
						err = builder.AddNode(astNode)
						if err != nil {
							cancel()
							t.Error(err)
						}
					}
				case <-ctx.Done():
					return
				}
			}
		}()
		<-ctx.Done()

		for _, file := range builder.Files() {
			if file.IsStub {
				continue
			}
			if _, ok := expectedImportRefsByFileName[file.FileName]; !ok {
				t.Errorf(
					"%s: unexpected file: %s \"%s\"",
					testCase,
					file.FileName,
					file.AbsPath,
				)
				continue
			}
			for _, imp := range file.Imports {
				if _, ok := expectedImportRefsByFileName[file.FileName][imp.Name]; !ok {
					t.Errorf(
						"%s: unexpected import: %s in %s",
						testCase,
						imp.Name,
						file.FileName,
					)
					continue
				}
				for _, decl := range imp.ReferencedTypes {
					qualifiedName := decl.QualifiedName()
					if _, ok := expectedImportRefsByFileName[file.FileName][imp.Name][qualifiedName]; !ok {
						t.Errorf(
							"%s: unexpected type: %s in %s",
							testCase,
							qualifiedName,
							file.FileName,
						)
						continue
					}
					delete(expectedImportRefsByFileName[file.FileName][imp.Name], qualifiedName)
				}
				if len(expectedImportRefsByFileName[file.FileName][imp.Name]) > 0 {
					continue
				}
				delete(expectedImportRefsByFileName[file.FileName], imp.Name)
			}
			if len(expectedImportRefsByFileName[file.FileName]) != 0 {
				for impName, imp := range expectedImportRefsByFileName[file.FileName] {
					for _, decl := range imp {
						t.Errorf(
							"%s: missing expected decls: %s.%s in %s",
							testCase,
							impName,
							decl.Name,
							file.FileName,
						)
					}

				}
			}
			delete(expectedImportRefsByFileName, file.FileName)
		}

		if len(expectedImportRefsByFileName) != 0 {
			for fileName := range expectedImportRefsByFileName {
				t.Errorf(
					"%s: missing expected file: %s",
					testCase,
					fileName,
				)
			}
		}
	}
}
