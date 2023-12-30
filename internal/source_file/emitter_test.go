package source_file_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/samlitowitz/goimportcycle/internal/source_file"
)

func TestSourceFileEmitter_WalkDirFunc_EmitsNothingAfterError(t *testing.T) {
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

	emitter, output := source_file.NewEmitter()
	go func(output <-chan string) {
		for range output {
			t.Fatal("no directories should be emitted")
		}
	}(output)

	err = filepath.WalkDir(tmpDir+"/NONEXISTENT_DIR", emitter.WalkDirFunc)
	if err == nil {
		t.Fatal("failed to emit error")
	}
	emitter.Close()
}

func TestSourceFileEmitter_WalkDirFunc_EmitsAppropriateFiles(t *testing.T) {
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
			{"a.go", nil},
			{"b", []*Node{}},
			{"c.go", nil},
			{
				"d",
				[]*Node{
					{"e.go", nil},
					{"f", []*Node{}},
					{
						"g",
						[]*Node{
							{"h.go", nil},
						},
					},
				},
			},
		},
	}

	expectedFiles := makeTree(t, tree)

	emitter, output := source_file.NewEmitter()
	go func(output <-chan string, expectedFiles map[string]struct{}) {
		for actualPath := range output {
			if _, ok := expectedFiles[actualPath]; !ok {
				t.Fatalf("unexpected path: %s", actualPath)
			}
			delete(expectedFiles, actualPath)
		}
		if len(expectedFiles) > 0 {
			missedPaths := ""
			for expectedFile := range expectedFiles {
				missedPaths += ", " + expectedFile
			}
			t.Fatalf(
				"not all expected paths sent: missing %s",
				missedPaths,
			)
		}
	}(output, expectedFiles)
	err = filepath.WalkDir(tree.name, emitter.WalkDirFunc)
	if err != nil {
		t.Fatal(err)
	}
	emitter.Close()
}

// REFURL: https://github.com/golang/go/blob/988b718f4130ab5b3ce5a5774e1a58e83c92a163/src/path/filepath/path_test.go#L449
type Node struct {
	name    string
	entries []*Node // nil if the entry is a file
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
	files := make(map[string]struct{})
	walkTree(tree, tree.name, func(path string, n *Node) {
		if n.entries == nil {
			fd, err := os.Create(path)
			if err != nil {
				t.Errorf("makeTree: %v", err)
				return
			}
			fd.Close()
			files[path] = struct{}{}
		} else {
			os.Mkdir(path, 0770)
		}
	})
	return files
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
