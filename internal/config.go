package internal

import "image/color"

var (
	True = &Config{
		Line:             Color{&color.RGBA{R: 0, G: 0, B: 0, A: 1}},
		PackageFill:      Color{&color.RGBA{R: 255, G: 255, B: 255, A: 1}},
		FileFill:         Color{&color.RGBA{R: 255, G: 255, B: 255, A: 1}},
		CycleLine:        Color{&color.RGBA{R: 220, G: 38, B: 127, A: 1}},
		CyclePackageFill: Color{&color.RGBA{R: 255, G: 176, B: 0, A: 1}},
		CycleFileFill:    Color{&color.RGBA{R: 220, G: 38, B: 127, A: 1}},
	}
)

type Config struct {
	Line             Color
	PackageFill      Color
	FileFill         Color
	CycleLine        Color
	CyclePackageFill Color
	CycleFileFill    Color
}
