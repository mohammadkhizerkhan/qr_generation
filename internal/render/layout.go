package render

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
)

type Layout struct {
	Width         int
	Height        int
	HeaderY       float64
	UPIIDY        float64
	DescriptionY  float64
	HintY         float64
	PayerY        float64
	QRX           int
	QRY           int
	QRSize        int
	TopBar        image.Rectangle
	Panel         image.Rectangle
	Footer        image.Rectangle
	CornerRadius  float64
	TopBarRadius  float64
	FooterRadius  float64
	TextWrapChars int
}

type Style struct {
	Background color.RGBA
	Accent     color.RGBA
	Text       color.RGBA
	Muted      color.RGBA
	Panel      color.RGBA
	FooterText color.RGBA
}

func DefaultLayout() Layout {
	template := config.DefaultConfig().Templates[config.DefaultConfig().DefaultTemplate]
	layout, _ := BuildLayout(template.Layout)
	return layout
}

func DefaultStyle() Style {
	template := config.DefaultConfig().Templates[config.DefaultConfig().DefaultTemplate]
	style, _ := BuildStyle(template.Style)
	return style
}

func BuildLayout(cfg config.LayoutConfig) (Layout, error) {
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return Layout{}, fmt.Errorf("layout width and height must be positive")
	}
	if cfg.QRSize <= 0 {
		return Layout{}, fmt.Errorf("layout qr_size must be positive")
	}

	return Layout{
		Width:         cfg.Width,
		Height:        cfg.Height,
		HeaderY:       cfg.HeaderY,
		UPIIDY:        cfg.UPIIDY,
		DescriptionY:  cfg.DescriptionY,
		HintY:         cfg.HintY,
		PayerY:        cfg.PayerY,
		QRX:           cfg.QRX,
		QRY:           cfg.QRY,
		QRSize:        cfg.QRSize,
		TopBar:        image.Rect(cfg.TopBarX, cfg.TopBarY, cfg.TopBarX+cfg.TopBarWidth, cfg.TopBarY+cfg.TopBarHeight),
		Panel:         image.Rect(cfg.PanelX, cfg.PanelY, cfg.PanelX+cfg.PanelWidth, cfg.PanelY+cfg.PanelHeight),
		Footer:        image.Rect(cfg.FooterX, cfg.FooterY, cfg.FooterX+cfg.FooterWidth, cfg.FooterY+cfg.FooterHeight),
		CornerRadius:  cfg.CornerRadius,
		TopBarRadius:  cfg.TopBarRadius,
		FooterRadius:  cfg.FooterRadius,
		TextWrapChars: cfg.TextWrapChars,
	}, nil
}

func BuildStyle(cfg config.StyleConfig) (Style, error) {
	background, err := ParseHexColor(cfg.BackgroundHex)
	if err != nil {
		return Style{}, fmt.Errorf("background_color: %w", err)
	}
	accent, err := ParseHexColor(cfg.AccentHex)
	if err != nil {
		return Style{}, fmt.Errorf("accent_color: %w", err)
	}
	text, err := ParseHexColor(cfg.TextHex)
	if err != nil {
		return Style{}, fmt.Errorf("text_color: %w", err)
	}
	muted, err := ParseHexColor(cfg.MutedHex)
	if err != nil {
		return Style{}, fmt.Errorf("muted_color: %w", err)
	}
	panel, err := ParseHexColor(cfg.PanelHex)
	if err != nil {
		return Style{}, fmt.Errorf("panel_color: %w", err)
	}
	footerText, err := ParseHexColor(cfg.FooterTextHex)
	if err != nil {
		return Style{}, fmt.Errorf("footer_text_color: %w", err)
	}

	return Style{
		Background: background,
		Accent:     accent,
		Text:       text,
		Muted:      muted,
		Panel:      panel,
		FooterText: footerText,
	}, nil
}

func ResolveLogoRect(canvasWidth int, footer image.Rectangle, cfg config.LogoConfig) image.Rectangle {
	w := max(cfg.Width, 1)
	h := max(cfg.Height, 1)
	x := cfg.X
	y := cfg.Y

	switch strings.ToLower(strings.TrimSpace(cfg.Placement)) {
	case "bottom-left":
		x = footer.Min.X + cfg.Padding
		y = footer.Min.Y + (footer.Dy()-h)/2
	case "bottom-right":
		x = footer.Max.X - w - cfg.Padding
		y = footer.Min.Y + (footer.Dy()-h)/2
	case "bottom-center", "":
		x = footer.Min.X + (footer.Dx()-w)/2
		y = footer.Min.Y + (footer.Dy()-h)/2
	}

	x = clamp(x, 0, max(canvasWidth-w, 0))
	y = max(y, 0)

	return image.Rect(x, y, x+w, y+h)
}

func ParseHexColor(value string) (color.RGBA, error) {
	hex := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(hex) != 6 {
		return color.RGBA{}, fmt.Errorf("must be a 6-digit hex color")
	}

	var rgb [3]uint8
	for index := 0; index < 3; index++ {
		var channel uint8
		if _, err := fmt.Sscanf(hex[index*2:index*2+2], "%02x", &channel); err != nil {
			return color.RGBA{}, fmt.Errorf("invalid hex color")
		}
		rgb[index] = channel
	}

	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}, nil
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
