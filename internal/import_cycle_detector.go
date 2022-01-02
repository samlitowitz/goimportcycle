package internal

import (
	"fmt"
	"github.com/samlitowitz/goimportcycleviz/internal/ast/pkg"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

type ImportCycleDetector struct {
	modulePath                  string
	typeDefFileByPackageVisitor *pkg.TypeDefFileByPackageVisitor

	files    map[pkg.FilePath]*File
	packages map[pkg.ImportPath]*Package
}

func NewImportCycleDetector(modulePath string) *ImportCycleDetector {
	return &ImportCycleDetector{
		modulePath:                  modulePath,
		typeDefFileByPackageVisitor: pkg.NewTypeDefFileByPackageVisitor(modulePath),
	}
}

func (icd *ImportCycleDetector) Run(path string) ([]*Package, error) {
	err := filepath.WalkDir(path, icd.buildImportsAndTypeDefs)
	if err != nil {
		return nil, err
	}
	icd.files = make(map[pkg.FilePath]*File)
	icd.packages = make(map[pkg.ImportPath]*Package)
	err = filepath.WalkDir(path, icd.buildImportDepGraph)
	if err != nil {
		return nil, err
	}

	count := 0
	for _, file := range icd.files {
		for _, pkg := range file.Imports {
			if _, ok := icd.packages[pkg.ImportPath]; ok {
				continue
			}
			icd.packages[pkg.ImportPath] = pkg
			count++
		}
	}

	pkgs := make([]*Package, len(icd.packages))
	for _, pkg := range icd.packages {
		pkgs = append(pkgs, pkg)
	}

	return pkgs, nil
}

func (icd *ImportCycleDetector) buildImportsAndTypeDefs(
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

	i := strings.LastIndex(path, icd.modulePath)
	if i == -1 {
		return fmt.Errorf("File not part of module %s", icd.modulePath)
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, 0)
	if err != nil {
		log.Fatal(err)
	}

	for _, astPkg := range pkgs {
		ast.Walk(icd.typeDefFileByPackageVisitor, astPkg)
	}
	return nil
}

func (icd *ImportCycleDetector) buildImportDepGraph(
	path string,
	info fs.DirEntry,
	err error,
) error {
	if err != nil {
		return err
	}

	if info.IsDir() {
		return nil
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return err
	}
	i := strings.LastIndex(path, icd.modulePath)
	if i == -1 {
		return fmt.Errorf("File not part of module %s", icd.modulePath)
	}

	filePath := pkg.FilePath(path[i:])
	fileImports, ok := icd.typeDefFileByPackageVisitor.PackageImportsByFile()[filePath]
	if !ok {
		return nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	content := string(data)
	fileByTypeByPackage := icd.typeDefFileByPackageVisitor.FileByTypeByPackage()

	file, ok := icd.files[filePath]
	if !ok {
		file = &File{
			Name:    filePath,
			Imports: make(map[pkg.ImportPath]*Package),
		}
		icd.files[filePath] = file
	}

	log.Println(filePath)

	for _, fileImport := range fileImports {
		fileByType, ok := fileByTypeByPackage[fileImport.Path]
		if !ok {
			continue
		}
		//importedPkg, ok := icd.packages[fileImport.Path]
		//if !ok {
		//	icd.packages[fileImport.Path] = &Package{
		//		ImportPath: fileImport.Path,
		//		Files:      make(map[pkg.FilePath]*File, 0),
		//	}
		//	importedPkg, _ = icd.packages[fileImport.Path]
		//}
		log.Println(".." + fileImport.Path)

		for typ, defFile := range fileByType {
			externalType := string(fileImport.Name) + "." + string(typ)
			if !strings.Contains(content, externalType) {
				continue
			}
			log.Println("...." + defFile)
			//file.Imports[fileImport.Path] = importedPkg
			//
			//importedFile, ok := icd.files[defFile]
			//if !ok {
			//	icd.files[defFile] = &File{
			//		Name:    defFile,
			//		Imports: make(map[pkg.ImportPath]*Package),
			//	}
			//	importedFile, _ = icd.files[defFile]
			//}
			//if _, ok := importedPkg.Files[defFile]; !ok {
			//	importedPkg.Files[defFile] = importedFile
			//}
		}
	}

	return nil
}
