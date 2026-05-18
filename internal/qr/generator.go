package qr

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"

	qrcode "github.com/skip2/go-qrcode"
)

type Generator struct {
	Size  int
	Level qrcode.RecoveryLevel
}

var ErrIconNotSupported = errors.New("qr icon embedding is not supported by skip2/go-qrcode")

func NewGenerator(size int) *Generator {
	if size <= 0 {
		size = 640
	}

	return &Generator{
		Size:  size,
		Level: qrcode.Medium,
	}
}

func (g *Generator) Image(content string) (image.Image, error) {
	code, err := qrcode.New(content, g.Level)
	if err != nil {
		return nil, fmt.Errorf("build qr image: %w", err)
	}

	return code.Image(g.Size), nil
}

func (g *Generator) ImageWithIcon(content string, _ image.Image) (image.Image, error) {
	return nil, ErrIconNotSupported
}

func (g *Generator) PNG(content string) ([]byte, error) {
	img, err := g.Image(content)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("encode qr png: %w", err)
	}

	return buf.Bytes(), nil
}
