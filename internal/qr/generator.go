package qr

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"

	qrcode "github.com/skip2/go-qrcode"
)

type Generator struct {
	Size  int
	Level qrcode.RecoveryLevel
}

func NewGenerator(size int) *Generator {
	return NewGeneratorWithLevel(size, "M")
}

func NewGeneratorWithLevel(size int, level string) *Generator {
	if size <= 0 {
		size = 640
	}

	return &Generator{
		Size:  size,
		Level: parseRecoveryLevel(level),
	}
}

func parseRecoveryLevel(level string) qrcode.RecoveryLevel {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "L":
		return qrcode.Low
	case "Q":
		return qrcode.High
	case "H":
		return qrcode.Highest
	default:
		return qrcode.Medium
	}
}

func (g *Generator) Image(content string) (image.Image, error) {
	code, err := qrcode.New(content, g.Level)
	if err != nil {
		return nil, fmt.Errorf("build qr image: %w", err)
	}

	return code.Image(g.Size), nil
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
