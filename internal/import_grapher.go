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

type ImportGrapher struct {
	modulePath                  string
	typeDefFileByPackageVisitor *pkg.TypeDefFileByPackageVisitor

	files map[pkg.FilePath]*File
}

func NewImportGrapher(modulePath string) *ImportGrapher {
	return &ImportGrapher{
		modulePath:                  modulePath,
		typeDefFileByPackageVisitor: pkg.NewTypeDefFileByPackageVisitor(modulePath),
	}
}

func (icd *ImportGrapher) Run(path string) (map[pkg.FilePath]*File, error) {
	err := filepath.WalkDir(path, icd.buildImportsAndTypeDefs)
	if err != nil {
		return nil, err
	}
	icd.files = make(map[pkg.FilePath]*File)
	err = filepath.WalkDir(path, icd.buildImportDepGraph)
	if err != nil {
		return nil, err
	}

	return icd.files, nil
}

func (icd *ImportGrapher) buildImportsAndTypeDefs(
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

func (icd *ImportGrapher) buildImportDepGraph(
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
		filePackage := filepath.Dir(string(filePath))
		file = &File{
			Name:    filePath,
			Package: pkg.ImportPath(filePackage),
			Imports: make(map[pkg.ImportPath]*File),
		}
		icd.files[filePath] = file
	}

	log.Println(filePath)

	for _, fileImport := range fileImports {
		fileByType, ok := fileByTypeByPackage[fileImport.Path]
		if !ok {
			continue
		}
		log.Println(".." + fileImport.Path)

		for typ, defFile := range fileByType {
			externalType := string(fileImport.Name) + "." + string(typ)
			if !strings.Contains(content, externalType) {
				continue
			}
			log.Println("...." + defFile)
			importedFile, ok := icd.files[defFile]
			if !ok {
				importedFile = &File{
					Name:    defFile,
					Imports: make(map[pkg.ImportPath]*File),
				}
				icd.files[defFile] = importedFile
			}
			file.Imports[fileImport.Path] = importedFile
		}
	}

	return nil
}
