package qr

import (
	"image"
	"image/color"
	"testing"
)

func TestYeqownGeneratorImage(t *testing.T) {
	t.Parallel()

	g := NewYeqownGenerator(320)
	img, err := g.Image("upi://pay?pa=merchant%40bank&pn=Shop")
	if err != nil {
		t.Fatalf("Image() error = %v", err)
	}
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 320 {
		t.Fatalf("bounds = %v, want 320x320", img.Bounds())
	}
}

func TestYeqownGeneratorImageWithIcon(t *testing.T) {
	t.Parallel()

	g := NewYeqownGenerator(320)
	icon := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			icon.Set(x, y, color.RGBA{R: 30, G: 30, B: 220, A: 255})
		}
	}

	img, err := g.ImageWithIcon("upi://pay?pa=merchant%40bank&pn=Shop", icon)
	if err != nil {
		t.Fatalf("ImageWithIcon() error = %v", err)
	}
	if img.Bounds().Dx() != 320 || img.Bounds().Dy() != 320 {
		t.Fatalf("bounds = %v, want 320x320", img.Bounds())
	}
}
