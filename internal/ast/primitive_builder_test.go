package ast_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samlitowitz/goimportcycle/internal"

	internalAST "github.com/samlitowitz/goimportcycle/internal/ast"
)

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
					DirName:    tmpDir + string(filepath.Separator) + "testdata" + string(filepath.Separator) + path,
					ImportPath: "",
					Name:       n.pkg,
					Files:      nil,
				}
				if n.name != "main" {
					pkg.ImportPath = path
				}
				expectedPackagesByDirName[pkg.DirName] = pkg

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
