package render

import (
	"bytes"
	"image/png"
	"testing"
)

func TestRenderPNG(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	data, err := renderer.RenderPNG(CardInput{
		UPIURI:       "upi://pay?pa=merchant%40bank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201234",
		MerchantName: "Simba Pvt Ltd",
	})
	if err != nil {
		t.Fatalf("RenderPNG() error = %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != DefaultLayout().Width || bounds.Dy() != DefaultLayout().Height {
		t.Fatalf("image bounds = %v, want %dx%d", bounds, DefaultLayout().Width, DefaultLayout().Height)
	}
}

func TestRenderPNGYeqownBackend(t *testing.T) {
	t.Parallel()

	renderer := NewRenderer()
	data, err := renderer.RenderPNG(CardInput{
		UPIURI:       "upi://pay?pa=merchant%40bank&pn=Simba%20Pvt%20Ltd&tn=Invoice%201235",
		MerchantName: "Simba Pvt Ltd",
		QRGenerator:  "yeqown",
	})
	if err != nil {
		t.Fatalf("RenderPNG() error = %v", err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("png.Decode() error = %v", err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != DefaultLayout().Width || bounds.Dy() != DefaultLayout().Height {
		t.Fatalf("image bounds = %v, want %dx%d", bounds, DefaultLayout().Width, DefaultLayout().Height)
	}
}
