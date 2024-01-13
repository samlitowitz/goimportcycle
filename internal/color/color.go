package color

import (
	"fmt"
	"image/color"
)

type Color struct {
	color.Color
}

func (c Color) Hex() string {
	r, g, b, _ := c.Color.RGBA()
	r = r >> 8
	g = g >> 8
	b = b >> 8
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}
