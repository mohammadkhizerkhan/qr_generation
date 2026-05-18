package render

import "image/color"

type Layout struct {
	Width        int
	Height       int
	HeaderY      float64
	UPIIDY       float64
	DescriptionY float64
	QRX          int
	QRY          int
	QRSize       int
	FooterY      int
	LogoX        int
	LogoY        int
	LogoWidth    int
	LogoHeight   int
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
		UPIIDY:       220,
		DescriptionY: 320,
		QRX:          170,
		QRY:          420,
		QRSize:       560,
		FooterY:      1160,
		LogoX:        290,
		LogoY:        1195,
		LogoWidth:    320,
		LogoHeight:   110,
	}
}

func DefaultStyle() Style {
	return Style{
		Background: color.RGBA{R: 251, G: 248, B: 241, A: 255},
		Accent:     color.RGBA{R: 159, G: 26, B: 26, A: 255},
		Text:       color.RGBA{R: 38, G: 34, B: 29, A: 255},
		Muted:      color.RGBA{R: 103, G: 94, B: 86, A: 255},
		Panel:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
	}
}
