package ast_test

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"

	"github.com/samlitowitz/goimportcycle/internal/directory"
)

func TestDependencyVisitor_Visit(t *testing.T) {
	dirEmitter, dirOut := directory.NewEmitter()
	defer dirEmitter.Close()
	depVis, depTokenOut := internalAST.NewDependencyVisitor()
	defer depVis.Close()

	go func(tokenOut <-chan internalAST.Token) {
		for tok := range tokenOut {
			fmt.Printf("Token: %d: %s", tok.Type(), tok.Val())
		}
	}(depTokenOut)

	go func(dirOut <-chan string, depVis *internalAST.DependencyVisitor) {
		for dirPath := range dirOut {
			fset := token.NewFileSet()
			pkgs, err := parser.ParseDir(fset, dirPath, nil, 0)
			if err != nil {
				log.Fatal(err)
			}

			for _, pkg := range pkgs {
				ast.Walk(depVis, pkg)
			}
		}
	}(dirOut, depVis)

	pkg, err := build.Default.Import("../../", ".", build.FindOnly)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%s", pkg.Name)
	err = filepath.WalkDir("../../examples", dirEmitter.WalkDirFunc)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDependencyVisitor_Visit_EmitsPackagesAndFiles_JustMain(t *testing.T) {
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
				"just_main",
				[]*Node{
					{
						"just_main.go",
						nil,
						`
package main

func main() {}
`,
					},
				},
				"",
			},
			{
				"no_main",
				[]*Node{
					{
						"no_main.go",
						nil,
						`
package no_main
`,
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								`
package a
`,
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										`
package b
`,
									},
								},
								"",
							},
						},
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								`
package c
`,
							},
						},
						"",
					},
				},
				"",
			},
			{
				"both",
				[]*Node{
					{
						"main.go",
						nil,
						`
package main
`,
					},
					{
						"a",
						[]*Node{
							{
								"a.go",
								nil,
								`
package a
`,
							},
							{
								"b",
								[]*Node{
									{
										"b.go",
										nil,
										`
package b
`,
									},
								},
								"",
							},
						},
						"",
					},
					{
						"c",
						[]*Node{
							{
								"c.go",
								nil,
								`
package c
`,
							},
						},
						"",
					},
				},
				"",
			},
		},
		"",
	}

	makeTree(t, tree)
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L449
type Node struct {
	name    string
	entries []*Node // nil if the entry is a file
	data    string  // nil if entry is a directory
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
