package internal

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Package struct {
	DirName string

	ModulePath string
	ModuleRoot string
	Name       string

	Files map[string]*File

	IsStub        bool
	InImportCycle bool
}

func (pkg Package) ImportPath() string {
	if pkg.Name == "main" {
		return ""
	}
	if strings.HasPrefix(pkg.DirName, pkg.ModuleRoot) {
		return fmt.Sprintf(
			"%s/%s",
			pkg.ModulePath,
			strings.TrimPrefix(
				pkg.DirName,
				pkg.ModuleRoot+string(filepath.Separator),
			),
		)
	}
	if strings.HasPrefix(pkg.DirName, pkg.ModulePath) {
		return pkg.DirName
	}

	return pkg.Name
}

func (pkg Package) ModuleRelativePath() string {
	if strings.HasPrefix(pkg.DirName, pkg.ModuleRoot) {
		path := strings.TrimPrefix(
			pkg.DirName,
			pkg.ModuleRoot,
		)
		path = strings.TrimPrefix(path, string(filepath.Separator))
		if pkg.Name != "main" {
			return path
		}
		if path == "" {
			return pkg.Name
		}
		return path + ":" + pkg.Name
	}
	if strings.HasPrefix(pkg.DirName, pkg.ModulePath) {
		path := strings.TrimPrefix(
			pkg.DirName,
			pkg.ModulePath,
		)
		path = strings.TrimPrefix(path, string(filepath.Separator))
		if pkg.Name != "main" {
			return path
		}
		if path == "" {
			return pkg.Name
		}
		return path + ":" + pkg.Name
	}
	return pkg.Name
}

func (pkg Package) UID() string {
	uid := pkg.ImportPath()
	if uid != "" {
		return uid
	}
	return pkg.DirName
}

type File struct {
	Package *Package

	FileName string
	AbsPath  string

	Imports map[string]*Import
	Decls   map[string]*Decl

	IsStub        bool
	InImportCycle bool
}

func (f File) HasDecl(decl *Decl) bool {
	for _, fDecl := range f.Decls {
		if decl.UID() != fDecl.UID() {
			continue
		}
		return true
	}
	return false
}

func (f File) ReferencedFiles() []*File {
	alreadyReferenced := make(map[string]struct{})
	referencedFiles := make([]*File, 0, len(f.Imports))

	for _, imp := range f.Imports {
		for _, typ := range imp.ReferencedTypes {
			if _, ok := alreadyReferenced[typ.File.AbsPath]; ok {
				continue
			}
			alreadyReferenced[typ.File.AbsPath] = struct{}{}
			referencedFiles = append(referencedFiles, typ.File)
		}
	}
	return referencedFiles

}

func (f File) UID() string {
	return f.AbsPath
}

type Decl struct {
	File         *File
	ReceiverDecl *Decl

	Name string
}

func (decl Decl) UID() string {
	return decl.QualifiedName()
}

func (d Decl) QualifiedName() string {
	if d.ReceiverDecl == nil {
		return d.Name
	}
	return d.ReceiverDecl.Name + "." + d.Name
}

type Import struct {
	Package *Package

	Name string
	Path string

	ReferencedTypes map[string]*Decl

	InImportCycle          bool
	ReferencedFilesInCycle map[string]*File
}

func (i Import) UID() string {
	return i.Name
}
