package qr

import (
	"image"
	"image/color"
	"sync"
)

var (
	defaultLogoOnce sync.Once
	defaultLogo     image.Image
)

func DefaultLogo() (image.Image, error) {
	defaultLogoOnce.Do(func() {
		logo := image.NewRGBA(image.Rect(0, 0, 32, 32))
		fillRect(logo, image.Rect(0, 0, 32, 32), color.RGBA{R: 164, G: 24, B: 24, A: 255})
		fillRect(logo, image.Rect(4, 4, 28, 28), color.RGBA{R: 255, G: 255, B: 255, A: 255})
		fillRect(logo, image.Rect(10, 4, 16, 28), color.RGBA{R: 164, G: 24, B: 24, A: 255})
		fillRect(logo, image.Rect(16, 10, 28, 16), color.RGBA{R: 164, G: 24, B: 24, A: 255})
		defaultLogo = logo
	})

	return defaultLogo, nil
}

func fillRect(img *image.RGBA, rect image.Rectangle, shade color.RGBA) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			img.SetRGBA(x, y, shade)
		}
	}
}
