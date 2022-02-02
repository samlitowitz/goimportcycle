package internal

import (
	"fmt"
	"image/color"
)

type Color struct {
	*color.RGBA
}

func (c Color) Hex() string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}
