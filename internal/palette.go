package internal

type HalfPalette struct {
	PackageName       Color
	PackageBackground Color
	FileName          Color
	FileBackground    Color
	ImportArrow       Color
}

type Palette struct {
	Base  *HalfPalette
	Cycle *HalfPalette
}
