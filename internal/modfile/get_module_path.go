package modfile

import (
	"golang.org/x/mod/modfile"
	"io/ioutil"
)

func GetModulePath(goModFile string) (string, error) {
	goMod, err := ioutil.ReadFile(goModFile)
	if err != nil {
		return "", err
	}
	modPath := modfile.ModulePath(goMod)
	if modPath == "" {
		return "", &ModulePathNotFoundError{}
	}
	return modPath, nil
}
