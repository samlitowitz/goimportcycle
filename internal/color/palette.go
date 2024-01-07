package color

import (
	"image/color"
)

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

var (
	Default = &Palette{
		Base: &HalfPalette{
			PackageName: Color{
				RGBA: &color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			PackageBackground: Color{
				RGBA: &color.RGBA{
					R: 255,
					G: 255,
					B: 255,
					A: 0,
				},
			},
			FileName: Color{
				RGBA: &color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			FileBackground: Color{
				RGBA: &color.RGBA{
					R: 255,
					G: 255,
					B: 255,
					A: 0,
				},
			},
			ImportArrow: Color{
				RGBA: &color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				},
			},
		},
		Cycle: &HalfPalette{
			PackageName: Color{
				RGBA: &color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			PackageBackground: Color{
				RGBA: &color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			FileName: Color{
				RGBA: &color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			FileBackground: Color{
				RGBA: &color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				},
			},
			ImportArrow: Color{
				RGBA: &color.RGBA{
					R: 255,
					G: 0,
					B: 0,
					A: 0,
				},
			},
		},
	}
)
