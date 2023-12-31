package tmp

type Package struct {
	Name  string
	Path  string
	Files map[string]*File
}

type File struct {
	Package *Package
	Path    string
	Imports map[string]*Import

	Exports map[string]struct{} // Exported constants, functions, types, and variables
}

type Import struct {
	Name       string
	Package    *Package
	References map[string]struct{} // Importee constants, functions, types, and variables referenced by importer
}
