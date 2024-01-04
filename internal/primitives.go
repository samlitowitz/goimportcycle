package internal

type Package struct {
	DirName string

	ImportPath string
	Name       string

	Files map[string]*File
}

type File struct {
	Package *Package

	FileName string
	AbsPath  string

	Imports map[string]*Import
	Decls   map[string]*Decl
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

type Decl struct {
	File         *File
	ReceiverDecl *Decl

	Name string
}

type Import struct {
	Package *Package
	File    *File

	Name string
	Path string

	ReferencedTypes map[string]*Decl
}
