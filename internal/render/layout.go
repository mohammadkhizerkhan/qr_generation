package render

import "image/color"

type Layout struct {
	Width        int
	Height       int
	HeaderY      float64
	UPIIDY       float64
	DividerY     float64
	DescriptionY float64
	QRX          int
	QRY          int
	QRSize       int
	BrandX       int
	BrandY       int
	BrandWidth   int
	BrandHeight  int
}

type Style struct {
	Background color.Color
	Accent     color.Color
	Text       color.Color
	Muted      color.Color
	Panel      color.Color
}

func DefaultLayout() Layout {
	return Layout{
		Width:        900,
		Height:       1400,
		HeaderY:      120,
		UPIIDY:       215,
		DividerY:     285,
		DescriptionY: 390,
		QRX:          170,
		QRY:          470,
		QRSize:       560,
		BrandX:       240,
		BrandY:       1180,
		BrandWidth:   420,
		BrandHeight:  160,
	}
}

func DefaultStyle() Style {
	return Style{
		Background: color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Accent:     color.RGBA{R: 159, G: 26, B: 26, A: 255},
		Text:       color.RGBA{R: 53, G: 48, B: 45, A: 255},
		Muted:      color.RGBA{R: 92, G: 85, B: 79, A: 255},
		Panel:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}
