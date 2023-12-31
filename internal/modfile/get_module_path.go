package modfile

import (
	"os"

	"golang.org/x/mod/modfile"
)

func GetModulePath(goModFile string) (string, error) {
	goMod, err := os.ReadFile(goModFile)
	if err != nil {
		return "", err
	}
	modPath := modfile.ModulePath(goMod)
	if modPath == "" {
		return "", &ModulePathNotFoundError{}
	}
	return modPath, nil
}
