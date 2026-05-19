package qr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"strings"

	goqr "github.com/piglig/go-qr"
)

type PigligGenerator struct {
	Size  int
	Level string
}

func NewPigligGenerator(size int) *PigligGenerator {
	if size <= 0 {
		size = 640
	}

	return &PigligGenerator{Size: size, Level: "M"}
}

func (g *PigligGenerator) Image(content string) (image.Image, error) {
	return g.render(content, nil)
}

func (g *PigligGenerator) ImageWithIcon(content string, icon image.Image) (image.Image, error) {
	if icon == nil {
		return nil, fmt.Errorf("qr icon image is required")
	}

	return g.render(content, icon)
}

func (g *PigligGenerator) render(content string, icon image.Image) (image.Image, error) {
	code, err := goqr.EncodeText(content, toPigligECC(g.Level))
	if err != nil {
		return nil, fmt.Errorf("build piglig qr image: %w", err)
	}

	options := make([]func(*goqr.QrCodeImgConfig), 0, 1)
	if icon != nil {
		options = append(options, goqr.WithLogo(icon, 0.2))
	}

	config := goqr.NewQrCodeImgConfig(10, 4, options...)
	pngBytes, err := code.ToPNGBytes(config)
	if err != nil {
		return nil, fmt.Errorf("encode piglig qr png: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("decode piglig qr image: %w", err)
	}

	if img.Bounds().Dx() != g.Size || img.Bounds().Dy() != g.Size {
		img = resizeImage(img, g.Size, g.Size)
	}

	return img, nil
}

func toPigligECC(level string) goqr.Ecc {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "L":
		return goqr.Low
	case "Q":
		return goqr.Quartile
	case "H":
		return goqr.High
	default:
		return goqr.Medium
	}
}
