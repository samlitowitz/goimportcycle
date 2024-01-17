package dot_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samlitowitz/goimportcycle/internal/config"
	"github.com/samlitowitz/goimportcycle/internal/dot"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"

	"github.com/google/go-cmp/cmp"
)

func TestMarshal_PackageNameWithDotsOrDashes_FileScope(t *testing.T) {
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
						"main.go",
						nil,
						"main",
						`
package main

import a "example.com/simple/a-a"

func main() {
	a.AFn()
}

`,
					},
					{
						"a-a",
						[]*Node{
							{
								"a.go",
								nil,
								"a_a",
								`
package a_a

import b "example.com/simple/b.b"

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
						"b.b",
						[]*Node{
							{
								"b.go",
								nil,
								"b_b",
								`
package b_b

import a "example.com/simple/a-a"

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
		builder := internalAST.NewPrimitiveBuilder(
			"example.com/"+testCase,
			tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+testCase,
		)

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
					err = builder.AddNode(astNode)
					if err != nil {
						cancel()
						t.Error(err)
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
			if file.Package.Name == "main" {
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
		expected := `digraph {
	labelloc="t";
	label="example.com/simple";
	rankdir="TB";
	node [shape="rect"];

	subgraph "cluster_pkg_a_a" {
		label="a-a";
		style="filled";
		fontcolor="#ff0000";
		fillcolor="#ffffff";

		"pkg_a_a_file_a" [label="a.go", style="filled", fontcolor="#ff0000", fillcolor="#ffffff"];
	};

	subgraph "cluster_pkg_b_b" {
		label="b.b";
		style="filled";
		fontcolor="#ff0000";
		fillcolor="#ffffff";

		"pkg_b_b_file_b" [label="b.go", style="filled", fontcolor="#ff0000", fillcolor="#ffffff"];
	};

	subgraph "cluster_pkg_main" {
		label="main";
		style="filled";
		fontcolor="#000000";
		fillcolor="#ffffff";

		"pkg_main_file_main" [label="main.go", style="filled", fontcolor="#000000", fillcolor="#ffffff"];
	};

	"pkg_a_a_file_a" -> "pkg_b_b_file_b" [color="#ff0000"];
	"pkg_b_b_file_b" -> "pkg_a_a_file_a" [color="#ff0000"];
	"pkg_main_file_main" -> "pkg_a_a_file_a" [color="#000000"];
}
`

		cfg := config.Default()
		cfg.Resolution = config.FileResolution
		actual, err := dot.Marshal(cfg, "example.com/"+testCase, builder.Packages())
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(expected, string(actual)) {
			t.Fatal(cmp.Diff(expected, string(actual)))
		}
	}
}

func TestMarshal_PackageNameWithDotsOrDashes_PackageScope(t *testing.T) {
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
						"main.go",
						nil,
						"main",
						`
package main

import a "example.com/simple/a-a"

func main() {
	a.AFn()
}

`,
					},
					{
						"a-a",
						[]*Node{
							{
								"a.go",
								nil,
								"a_a",
								`
package a_a

import b "example.com/simple/b.b"

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
						"b.b",
						[]*Node{
							{
								"b.go",
								nil,
								"b_b",
								`
package b_b

import a "example.com/simple/a-a"

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
		builder := internalAST.NewPrimitiveBuilder(
			"example.com/"+testCase,
			tmpDir+string(os.PathSeparator)+"testdata"+string(os.PathSeparator)+testCase,
		)

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
					err = builder.AddNode(astNode)
					if err != nil {
						cancel()
						t.Error(err)
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
			if file.Package.Name == "main" {
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
		expected := `digraph {
	labelloc="t";
	label="example.com/simple";
	rankdir="TB";
	node [shape="rect"];

	"pkg_a_a" [label="a-a", style="filled", fontcolor="#ff0000", fillcolor="#ffffff"];
	"pkg_b_b" [label="b.b", style="filled", fontcolor="#ff0000", fillcolor="#ffffff"];
	"pkg_main" [label="main", style="filled", fontcolor="#000000", fillcolor="#ffffff"];
	"pkg_a_a" -> "pkg_b_b" [color="#ff0000"];
	"pkg_b_b" -> "pkg_a_a" [color="#ff0000"];
	"pkg_main" -> "pkg_a_a" [color="#000000"];
}
`

		cfg := config.Default()
		cfg.Resolution = config.PackageResolution
		actual, err := dot.Marshal(cfg, "example.com/"+testCase, builder.Packages())
		if err != nil {
			t.Fatal(err)
		}
		if !cmp.Equal(expected, string(actual)) {
			t.Fatal(cmp.Diff(expected, string(actual)))
		}
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
