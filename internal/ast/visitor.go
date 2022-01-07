package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"unicode"
)

type Visitor struct {
	modulePath string

	// Packages
	packages       []*Package
	packagesByPath map[string]*Package

	// Files
	filesByPath   map[string]*File
	filePathByAst map[*ast.File]string

	//
	curPkg  *Package
	curFile *File
}

func NewVisitor(modulePath string) *Visitor {
	return &Visitor{
		modulePath:     modulePath,
		packages:       make([]*Package, 0),
		packagesByPath: make(map[string]*Package),
		filesByPath:    make(map[string]*File),
		filePathByAst:  make(map[*ast.File]string),
	}
}

func (v *Visitor) Visit(n ast.Node) ast.Visitor {
	switch n := n.(type) {
	case *ast.Package:
		pkgPath := ""
		toAdd := make(map[string]*ast.File, len(n.Files))
		toDelete := make([]string, 0, len(n.Files))

		for filePath, astFile := range n.Files {
			absPath, err := filepath.Abs(filePath)
			if err != nil {
				continue
			}

			i := strings.LastIndex(absPath, v.modulePath)
			if i == -1 {
				continue
			}

			if pkgPath == "" {
				pkgPath, _ = filepath.Split(absPath[i:])
				pkgPath = strings.TrimRight(pkgPath, "/")
			}

			toAdd[absPath[i:]] = astFile
			toDelete = append(toDelete, filePath)
		}
		for filePath, astFile := range toAdd {
			n.Files[filePath] = astFile
			v.filePathByAst[astFile] = filePath
		}
		for _, filePath := range toDelete {
			delete(n.Files, filePath)
		}

		curPkg, ok := v.packagesByPath[pkgPath]
		if !ok {
			curPkg = &Package{
				Name:  strings.Trim(n.Name, "\""),
				Path:  pkgPath,
				Files: make(map[string]*File),
			}
			v.packages = append(v.packages, curPkg)
			v.packagesByPath[pkgPath] = curPkg
		}
		v.curPkg = curPkg
		return v

	case *ast.File:
		filePath, ok := v.filePathByAst[n]
		if !ok {
			return nil
		}

		file, ok := v.filesByPath[filePath]
		if !ok {
			file = &File{
				Path:    filePath,
				Imports: make(map[string]*Import),
				Exports: make(map[string]struct{}),
			}
			v.filesByPath[filePath] = file
		}

		for _, astImportSpec := range n.Imports {
			pkgPath := strings.Trim(astImportSpec.Path.Value, "\"")
			pkgName := extractPackageNameFromImportPath(pkgPath)
			if astImportSpec.Name != nil {
				pkgName = astImportSpec.Name.String()
			}
			pkgName = strings.Trim(pkgName, "\"")

			importPkg, ok := v.packagesByPath[pkgPath]
			if !ok {
				importPkg = &Package{
					Name:  pkgName,
					Path:  pkgPath,
					Files: make(map[string]*File),
				}
				v.packages = append(v.packages, importPkg)
				v.packagesByPath[pkgPath] = importPkg
			}

			fileImport := &Import{
				Name:       pkgName,
				Package:    importPkg,
				References: make(map[string]struct{}),
			}
			file.Imports[pkgName] = fileImport
		}
		v.curPkg.Files[filePath] = file
		v.curFile = file
		return v

	case *ast.ImportSpec:
		return nil

	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			v.exportConstants(n)
		case token.TYPE:
			v.exportTypes(n)
		case token.VAR:
			v.exportVars(n)
		}
		return v

	case *ast.FuncDecl:
		v.exportFunc(n)
		return v

	case *ast.SelectorExpr:
		v.referenceSelector(n)
		return v
	}
	return v
}

func (v *Visitor) exportConstants(n *ast.GenDecl) {
	if v.curFile == nil {
		return
	}
	if n.Tok != token.CONST {
		return
	}
	for _, spec := range n.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, ident := range spec.Names {
			name := []rune(ident.String())
			if !unicode.IsUpper(name[1]) {
				continue
			}
			v.curFile.Exports[ident.String()] = struct{}{}
		}
	}
}

func (v *Visitor) exportFunc(n *ast.FuncDecl) {
	if v.curFile == nil {
		return
	}
	name := []rune(n.Name.String())
	if unicode.IsUpper(name[0]) {
		v.curFile.Exports[n.Name.String()] = struct{}{}
	}
}

func (v *Visitor) exportTypes(n *ast.GenDecl) {
	if v.curFile == nil {
		return
	}
	if n.Tok != token.TYPE {
		return
	}
	for _, spec := range n.Specs {
		spec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		name := []rune(spec.Name.String())
		if !unicode.IsUpper(name[0]) {
			continue
		}
		v.curFile.Exports[spec.Name.String()] = struct{}{}
	}
}

func (v *Visitor) exportVars(n *ast.GenDecl) {
	if v.curFile == nil {
		return
	}
	if n.Tok != token.VAR {
		return
	}
	for _, spec := range n.Specs {
		spec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, ident := range spec.Names {
			name := []rune(ident.String())
			if !unicode.IsUpper(name[0]) {
				continue
			}
			v.curFile.Exports[ident.String()] = struct{}{}
		}
	}
}

func (v *Visitor) referenceSelector(n *ast.SelectorExpr) {
	if v.curFile == nil {
		return
	}
	x, ok := n.X.(*ast.Ident)
	if !ok {
		return
	}
	pkgName := x.String()
	fileImport, ok := v.curFile.Imports[pkgName]
	if !ok {
		return
	}
	fileImport.References[n.Sel.String()] = struct{}{}
}

func (v *Visitor) WalkDirFn(
	path string,
	info fs.DirEntry,
	err error,
) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}

	i := strings.LastIndex(path, v.modulePath)
	if i == -1 {
		return fmt.Errorf("File not part of module %s", v.modulePath)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, 0)
	if err != nil {
		return err
	}

	for _, astPkg := range pkgs {
		ast.Walk(v, astPkg)
	}
	return nil
}

func (v *Visitor) Packages() []*Package {
	return v.packages
}

func extractPackageNameFromImportPath(path string) string {
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return path
	}
	return path[i+1:]
}
