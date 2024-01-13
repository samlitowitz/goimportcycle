package config_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/samlitowitz/goimportcycle/internal/color"

	"github.com/samlitowitz/goimportcycle/internal/config"
)

func TestFromYamlFile_Empty(t *testing.T) {
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

	configPath := tmpDir + string(os.PathSeparator) + "config.yaml"
	writeConfig(t, configPath, "")
	cfg, err := config.FromYamlFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil {
		t.Fatal("no config created")
	}
	expected := config.Default()

	compareHalfPalette(t, expected.Palette.Base, cfg.Palette.Base)
	compareHalfPalette(t, expected.Palette.Cycle, cfg.Palette.Cycle)
}

// from partial yaml
// from entire yaml

func compareHalfPalette(t *testing.T, expected, actual *color.HalfPalette) {
	if expected.PackageName.Hex() != actual.PackageName.Hex() {
		t.Errorf("PackageName: expected %s got %s", expected.PackageName.Hex(), actual.PackageName.Hex())
	}
	if expected.PackageBackground.Hex() != actual.PackageBackground.Hex() {
		t.Errorf("PackageBackground: expected %s got %s", expected.PackageBackground.Hex(), actual.PackageBackground.Hex())
	}
	if expected.FileName.Hex() != actual.FileName.Hex() {
		t.Errorf("FileName: expected %s got %s", expected.FileName.Hex(), actual.FileName.Hex())
	}
	if expected.FileBackground.Hex() != actual.FileBackground.Hex() {
		t.Errorf("FileBackground: expected %s got %s", expected.FileBackground.Hex(), actual.FileBackground.Hex())
	}
	if expected.ImportArrow.Hex() != actual.ImportArrow.Hex() {
		t.Errorf("ImportArrow: expected %s got %s", expected.ImportArrow.Hex(), actual.ImportArrow.Hex())
	}
}

func writeConfig(t *testing.T, filePath, data string) {
	fd, err := os.Create(filePath)
	if err != nil {
		t.Errorf("makeTree: %v", err)
		return
	}
	if data != "" {
		_, err = fd.Write([]byte(data))
		if err != nil {
			t.Errorf("makeTree: %v", err)
			return
		}
	}
	fd.Close()
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
