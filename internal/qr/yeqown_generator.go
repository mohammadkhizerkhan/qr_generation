package qr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"strings"

	qrcode "github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
	"golang.org/x/image/draw"
)

type YeqownGenerator struct {
	Size  int
	Level string
}

func NewYeqownGenerator(size int) *YeqownGenerator {
	if size <= 0 {
		size = 640
	}

	return &YeqownGenerator{Size: size, Level: "M"}
}

func (g *YeqownGenerator) Image(content string) (image.Image, error) {
	return g.render(content, nil)
}

func (g *YeqownGenerator) ImageWithIcon(content string, icon image.Image) (image.Image, error) {
	if icon == nil {
		return nil, fmt.Errorf("qr icon image is required")
	}
	return g.render(content, icon)
}

func (g *YeqownGenerator) render(content string, icon image.Image) (image.Image, error) {
	code, err := qrcode.NewWith(content, qrErrorCorrectionOption(g.Level))
	if err != nil {
		return nil, fmt.Errorf("build yeqown qr image: %w", err)
	}

	width := g.Size
	if width > 255 {
		width = 255
	}
	if width <= 0 {
		width = 64
	}

	writerOptions := []standard.ImageOption{standard.WithQRWidth(uint8(width))}
	if icon != nil {
		writerOptions = append(writerOptions, standard.WithLogoImage(icon), standard.WithLogoSafeZone())
	}

	var buf bytes.Buffer
	writer := standard.NewWithWriter(&nopWriteCloser{Writer: &buf}, writerOptions...)
	if err := code.Save(writer); err != nil {
		return nil, fmt.Errorf("save yeqown qr image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("decode yeqown qr image: %w", err)
	}

	if img.Bounds().Dx() != g.Size || img.Bounds().Dy() != g.Size {
		img = resizeImage(img, g.Size, g.Size)
	}

	return img, nil
}

func qrErrorCorrectionOption(level string) qrcode.EncodeOption {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "L":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionLow)
	case "Q":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionQuart)
	case "H":
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionHighest)
	default:
		return qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium)
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
