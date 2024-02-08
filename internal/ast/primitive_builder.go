package ast

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/samlitowitz/goimportcycle/internal"
)

type PrimitiveBuilder struct {
	modulePath    string
	moduleRootDir string

	packagesByUID map[string]*internal.Package
	filesByUID    map[string]*internal.File

	curPkg  *internal.Package
	curFile *internal.File
}

func NewPrimitiveBuilder(modulePath, moduleRootDir string) *PrimitiveBuilder {
	return &PrimitiveBuilder{
		modulePath:    modulePath,
		moduleRootDir: moduleRootDir,

		packagesByUID: make(map[string]*internal.Package),
		filesByUID:    make(map[string]*internal.File),
	}
}

func (builder *PrimitiveBuilder) MarkupImportCycles() error {
	stk := NewFileStack()
	for _, baseFile := range builder.filesByUID {
		stk.Push(baseFile)
		err := builder.markupImportCycles(baseFile, stk)
		if err != nil {
			return err
		}
		stk.Pop()
	}
	return nil
}

func (builder *PrimitiveBuilder) markupImportCycles(
	baseFile *internal.File,
	stk *FileStack,
) error {
	for _, imp := range baseFile.Imports {
		for _, typ := range imp.ReferencedTypes {
			refFile := typ.File
			if stk.Contains(refFile) {
				for i := stk.Len() - 1; i > 0; i-- {
					curFile := stk.At(i)
					curFile.InImportCycle = true
					curFile.Package.InImportCycle = true
					imp.InImportCycle = true
					imp.ReferencedFilesInCycle[refFile.UID()] = refFile
					if curFile.UID() == baseFile.UID() {
						return nil
					}
				}
			}
			stk.Push(refFile)
			err := builder.markupImportCycles(refFile, stk)
			if err != nil {
				return err
			}
			stk.Pop()
		}
	}
	for _, refFile := range baseFile.ReferencedFiles() {
		// If you eventually import yourself, it's a cycle
		if stk.Contains(refFile) {
			for i := stk.Len() - 1; i > 0; i-- {
				curFile := stk.At(i)
				curFile.InImportCycle = true
				curFile.Package.InImportCycle = true
				if curFile.UID() == baseFile.UID() {
					return nil
				}
			}
		}
		stk.Push(refFile)
		err := builder.markupImportCycles(refFile, stk)
		if err != nil {
			return err
		}
		stk.Pop()
	}
	return nil
}

func (builder *PrimitiveBuilder) AddNode(node ast.Node) error {
	switch node := node.(type) {
	case *Package:
		return builder.addPackage(node)

	case *File:
		return builder.addFile(node)

	case *ImportSpec:
		return builder.addImport(node)

	case *FuncDecl:
		return builder.addFuncDecl(node)

	case *ast.GenDecl:
		return builder.addGenDecl(node)

	case *SelectorExpr:
		return builder.addSelectorExpr(node)
	}

	return nil
}

func (builder *PrimitiveBuilder) Files() []*internal.File {
	files := make([]*internal.File, 0, len(builder.filesByUID))
	for _, file := range builder.filesByUID {
		files = append(files, file)
	}
	return files
}

func (builder *PrimitiveBuilder) Packages() []*internal.Package {
	pkgs := make([]*internal.Package, 0, len(builder.packagesByUID))
	for _, pkg := range builder.packagesByUID {
		pkgs = append(pkgs, pkg)
	}
	return pkgs
}

func (builder *PrimitiveBuilder) addPackage(node *Package) error {
	newPkg := buildPackage(
		builder.modulePath,
		builder.moduleRootDir,
		node.DirName,
		node.Name,
		len(node.Files),
	)
	newPkgUID := newPkg.UID()

	pkg, pkgExists := builder.packagesByUID[newPkgUID]

	if pkgExists && !pkg.IsStub {
		// return custom error, duplicate package
	}

	// replace stub with the real thing
	if pkgExists && pkg.IsStub {
		// Does this work or do we need a full traversal?
		copyPackage(pkg, newPkg)
		for _, file := range pkg.Files {
			if !file.IsStub {
				continue
			}
			if len(file.Decls) > 0 {
				continue
			}
			// remove stub files with no declarations
			delete(pkg.Files, file.UID())
			delete(builder.filesByUID, file.UID())
		}
	}

	// totally new package
	if !pkgExists {
		builder.packagesByUID[newPkgUID] = newPkg
	}

	builder.curPkg = builder.packagesByUID[newPkgUID]
	return nil
}

func (builder *PrimitiveBuilder) addFile(node *File) error {
	if builder.curPkg == nil {
		// return custom error, undefined package
	}
	file := &internal.File{
		Package:  builder.packagesByUID[builder.curPkg.UID()],
		FileName: filepath.Base(node.AbsPath),
		AbsPath:  node.AbsPath,
		Imports:  make(map[string]*internal.Import),
		Decls:    make(map[string]*internal.Decl),
	}
	fileUID := file.UID()
	if _, ok := builder.filesByUID[fileUID]; ok {
		// return custom error, duplicate file
	}

	pkgUID := builder.curPkg.UID()
	builder.filesByUID[fileUID] = file
	builder.packagesByUID[pkgUID].Files[fileUID] = builder.filesByUID[fileUID]
	builder.curFile = builder.filesByUID[fileUID]

	return nil
}

func (builder *PrimitiveBuilder) addImport(node *ImportSpec) error {
	if builder.curPkg == nil {
		// return custom error, undefined package
	}
	if builder.curFile == nil {
		// return custom error, undefined file
	}
	imp := &internal.Import{
		Name:                   node.Name.String(),
		Path:                   node.Path.Value,
		ReferencedTypes:        make(map[string]*internal.Decl),
		ReferencedFilesInCycle: make(map[string]*internal.File),
	}
	if node.IsAliased {
		imp.Name = node.Alias
	}
	impUID := imp.UID()
	if _, ok := builder.curFile.Imports[impUID]; ok {
		// return custom error, duplicate import
	}

	// if the package exists, use it, otherwise use a stub
	pkg := buildPackage(builder.modulePath, builder.moduleRootDir, imp.Path, imp.Name, 1)
	pkg.IsStub = true
	if _, ok := builder.packagesByUID[pkg.UID()]; ok {
		pkg = builder.packagesByUID[pkg.UID()]
	} else {
		fileStub := buildStubFile(pkg)
		pkg.Files[fileStub.UID()] = fileStub
		builder.filesByUID[fileStub.UID()] = fileStub
		builder.packagesByUID[pkg.UID()] = pkg
	}

	imp.Package = pkg

	builder.curFile.Imports[impUID] = imp
	return nil
}

func (builder *PrimitiveBuilder) addFuncDecl(node *FuncDecl) error {
	if builder.curPkg == nil {
		// return custom error, undefined package
	}
	if builder.curFile == nil {
		// return custom error, undefined file
	}
	if node.Name.String() == "" {
		// return custom error, invalid function name
	}
	declUID := node.QualifiedName
	if _, ok := builder.curFile.Decls[declUID]; ok {
		// return custom error, duplicate decl
	}
	// TODO: receiver methods should never be received and should be skipped
	var receiverDecl *internal.Decl
	for _, file := range builder.curPkg.Files {
		if _, ok := file.Decls[node.ReceiverName]; ok {
			receiverDecl = file.Decls[node.ReceiverName]
			break
		}
	}
	decl := &internal.Decl{
		File:         builder.curFile,
		ReceiverDecl: receiverDecl,
		Name:         node.Name.String(),
	}
	decl = builder.fixupStubDecl(decl)
	builder.curFile.Decls[declUID] = decl

	return nil
}

func (builder *PrimitiveBuilder) addGenDecl(node *ast.GenDecl) error {
	if builder.curPkg == nil {
		// return custom error, undefined package
	}
	if builder.curFile == nil {
		// return custom error, undefined file
	}
	for _, spec := range node.Specs {
		switch spec := spec.(type) {
		case *ast.TypeSpec:
			if node.Tok != token.TYPE {
				// return custom error, invalid declaration
			}
			if _, ok := builder.curFile.Decls[spec.Name.String()]; ok {
				// return custom error, duplicate decl
			}
			if spec.Name.String() == "" {
				// return custom error, invalid type name
			}
			decl := &internal.Decl{
				File:         builder.curFile,
				ReceiverDecl: nil,
				Name:         spec.Name.String(),
			}
			decl = builder.fixupStubDecl(decl)
			builder.curFile.Decls[decl.UID()] = decl

		case *ast.ValueSpec:
			if node.Tok != token.CONST && node.Tok != token.VAR {
				// return custom error, invalid declaration
			}
			for _, name := range spec.Names {
				if _, ok := builder.curFile.Decls[name.String()]; ok {
					// return custom error, duplicate decl
				}
				if name.String() == "" {
					// return custom error, invalid const/var name
				}
				decl := &internal.Decl{
					File:         builder.curFile,
					ReceiverDecl: nil,
					Name:         name.String(),
				}
				decl = builder.fixupStubDecl(decl)
				builder.curFile.Decls[decl.UID()] = decl
			}

		default:
			// return custom error, unhandled spec type
		}
	}
	return nil
}

func (builder *PrimitiveBuilder) addSelectorExpr(node *SelectorExpr) error {
	if builder.curPkg == nil {
		// return custom error, undefined package
	}
	if builder.curFile == nil {
		// return custom error, undefined file
	}
	imp, hasImp := builder.curFile.Imports[node.ImportName]

	if !hasImp {
		// return custom error, undefined import
	}

	decl := &internal.Decl{
		Name: node.Sel.String(),
	}

	if _, ok := imp.ReferencedTypes[decl.Name]; ok {
		// type already registered
		return nil
	}

	if imp.Package == nil {
		// return custom error, undefined package
	}

	// attempt to find file where declaration is defined
	var foundDecl bool
	var stubFile *internal.File
	for _, file := range imp.Package.Files {
		// track stub file,
		if file.IsStub {
			stubFile = file
			continue
		}
		if !file.HasDecl(decl) {
			continue
		}
		decl.File = file
		foundDecl = true
		break
	}

	// if not file is found, attempt to add to a stub file
	if !foundDecl {
		if stubFile == nil {
			// when stub is empty return
			return nil
		}
		stubDecl, isDeclInStub := stubFile.Decls[decl.UID()]
		if isDeclInStub {
			decl = stubDecl
		}
		if !isDeclInStub {
			decl.File = stubFile
			// add declaration to stub file
			stubFile.Decls[decl.UID()] = decl
		}
	}

	if decl.File == nil {
		// return custom error, missing type declaration
	}

	imp.ReferencedTypes[decl.Name] = decl
	return nil
}

func (builder *PrimitiveBuilder) fixupStubDecl(newDecl *internal.Decl) *internal.Decl {
	for fileUID, file := range builder.curPkg.Files {
		// can only fix-up declarations in stub files
		if !file.IsStub {
			continue
		}
		for stubDeclUID, stubDecl := range file.Decls {
			// can only fix-up the same declaration
			if newDecl.UID() != stubDecl.UID() {
				continue
			}
			// everything is already pointing at the stub declaration
			// update stub declaration with values from the new declaration
			copyDeclaration(stubDecl, newDecl)

			// remove declaration from stub file
			delete(file.Decls, stubDeclUID)

			// if there are no more declarations in the stub file remove it
			if len(file.Decls) == 0 {
				delete(builder.curPkg.Files, fileUID)
				delete(builder.filesByUID, fileUID)
			}
			return stubDecl
		}
	}
	return newDecl
}

func buildPackage(
	modulePath,
	moduleRootDir,
	dirName, name string,
	fileCount int,
) *internal.Package {
	pkg := &internal.Package{
		DirName:    dirName,
		ModulePath: modulePath,
		ModuleRoot: moduleRootDir,
		Name:       name,
		Files:      make(map[string]*internal.File, fileCount),
	}
	return pkg
}

func buildStubFile(pkg *internal.Package) *internal.File {
	return &internal.File{
		Package:  pkg,
		FileName: "stub.go",
		AbsPath: fmt.Sprintf(
			"STUB://%s/stub.go",
			pkg.UID(),
		),
		Imports: make(map[string]*internal.Import),
		Decls:   make(map[string]*internal.Decl),
		IsStub:  true,
	}
}

func copyPackage(to, from *internal.Package) {
	to.DirName = from.DirName
	to.ModuleRoot = from.ModuleRoot
	to.Name = from.Name
	if from.Files != nil {
		for uid, file := range from.Files {
			to.Files[uid] = file
		}
	}
	to.IsStub = from.IsStub
	to.InImportCycle = from.InImportCycle
}

func copyDeclaration(to, from *internal.Decl) {
	to.File = from.File
	to.ReceiverDecl = from.ReceiverDecl
	to.Name = from.Name
}

type FileStack struct {
	indexByUID map[string]int
	stack      []*internal.File
}

func NewFileStack() *FileStack {
	return &FileStack{
		indexByUID: make(map[string]int),
		stack:      make([]*internal.File, 0),
	}
}

func (s *FileStack) Push(f *internal.File) {
	s.stack = append(s.stack, f)
	s.indexByUID[f.UID()] = len(s.stack) - 1
}

func (s *FileStack) Pop() {
	delete(s.indexByUID, s.stack[len(s.stack)-1].UID())
	s.stack = s.stack[0 : len(s.stack)-1]
}

func (s *FileStack) Top() *internal.File {
	if len(s.stack) == 0 {
		return nil
	}
	return s.stack[len(s.stack)-1]
}

func (s *FileStack) At(i int) *internal.File {
	if len(s.stack) == 0 {
		return nil
	}
	if i > (len(s.stack) - 1) {
		return nil
	}
	return s.stack[i]
}

func (s *FileStack) Contains(f *internal.File) bool {
	_, ok := s.indexByUID[f.UID()]
	return ok
}

func (s *FileStack) Len() int {
	return len(s.stack)
}
