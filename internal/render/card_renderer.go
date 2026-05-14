package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font/basicfont"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
	"github.com/mohammadkhizerkhan/qr_generation/internal/qr"
	"github.com/mohammadkhizerkhan/qr_generation/internal/upi"
)

type CardInput struct {
	UPIURI        string
	MerchantName  string
	MerchantUPIID string
	Description   string
	ProviderName  string
	PayerName     string
	LogoBase64    string
	BackgroundHex string
	AccentHex     string
	TextHex       string
	TemplateID    string
}

type Renderer struct {
	config      *config.Config
	fontManager *FontManager
}

func NewRenderer() *Renderer {
	defaultConfig := config.DefaultConfig()
	return NewRendererWithConfig(&defaultConfig)
}

func NewRendererWithConfig(cfg *config.Config) *Renderer {
	if cfg == nil {
		defaults := config.DefaultConfig()
		cfg = &defaults
	}

	return &Renderer{
		config:      cfg,
		fontManager: NewFontManager(cfg.Typography),
	}
}

func (r *Renderer) RenderPNG(input CardInput) ([]byte, error) {
	validated, err := upi.Validate(input.UPIURI)
	if err != nil {
		return nil, err
	}

	merchantName := strings.TrimSpace(input.MerchantName)
	if merchantName == "" {
		merchantName = validated.Params.Get("pn")
	}
	if merchantName == "" {
		return nil, fmt.Errorf("merchant_name is required")
	}

	merchantUPIID := strings.TrimSpace(input.MerchantUPIID)
	if merchantUPIID == "" {
		merchantUPIID = validated.MerchantID
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		description = validated.Params.Get("tn")
	}
	if description == "" {
		description = "Scan this QR code with any UPI app to transfer funds."
	}

	providerName := strings.TrimSpace(input.ProviderName)
	if providerName == "" {
		providerName = "Powered by QR Generation"
	}

	templateID := strings.TrimSpace(input.TemplateID)
	if templateID == "" {
		templateID = r.config.DefaultTemplate
	}
	template, ok := r.config.Templates[templateID]
	if !ok {
		return nil, fmt.Errorf("unknown template_id %q", templateID)
	}

	layout, err := BuildLayout(template.Layout)
	if err != nil {
		return nil, err
	}

	templateStyle, err := BuildStyle(template.Style)
	if err != nil {
		return nil, err
	}

	style, err := r.resolveStyle(templateStyle, input)
	if err != nil {
		return nil, err
	}

	renderScale := max(r.config.Quality.RenderScale, 1)
	qrGenerator := qr.NewGeneratorWithLevel(layout.QRSize*renderScale, r.config.Quality.QRRecoveryLevel)

	qrImage, err := qrGenerator.Image(validated.Raw)
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(layout.Width*renderScale, layout.Height*renderScale)
	dc.SetColor(style.Background)
	dc.Clear()

	dc.SetColor(style.Accent)
	topBar := scaleRect(layout.TopBar, renderScale)
	dc.DrawRoundedRectangle(float64(topBar.Min.X), float64(topBar.Min.Y), float64(topBar.Dx()), float64(topBar.Dy()), layout.TopBarRadius*float64(renderScale))
	dc.Fill()

	dc.SetColor(style.Panel)
	panel := scaleRect(layout.Panel, renderScale)
	dc.DrawRoundedRectangle(float64(panel.Min.X), float64(panel.Min.Y), float64(panel.Dx()), float64(panel.Dy()), layout.CornerRadius*float64(renderScale))
	dc.Fill()

	dc.SetColor(style.Text)
	r.setFont(dc, r.config.Typography.HeaderFont, r.config.Typography.HeaderSize*float64(renderScale))
	drawCenteredLines(dc, strings.ToUpper(merchantName), float64(layout.Width*renderScale)/2, layout.HeaderY*float64(renderScale), layout.TextWrapChars)

	dc.SetColor(style.Muted)
	r.setFont(dc, r.config.Typography.BodyFont, r.config.Typography.BodySize*float64(renderScale))
	drawCenteredLines(dc, "UPI ID: "+merchantUPIID, float64(layout.Width*renderScale)/2, layout.UPIIDY*float64(renderScale), layout.TextWrapChars)
	drawCenteredLines(dc, description, float64(layout.Width*renderScale)/2, layout.DescriptionY*float64(renderScale), layout.TextWrapChars)
	drawCenteredLines(dc, "Scan with any supported UPI app", float64(layout.Width*renderScale)/2, layout.HintY*float64(renderScale), layout.TextWrapChars)

	dc.DrawImage(qrImage, layout.QRX*renderScale, layout.QRY*renderScale)

	dc.SetColor(style.Accent)
	footer := scaleRect(layout.Footer, renderScale)
	dc.DrawRoundedRectangle(float64(footer.Min.X), float64(footer.Min.Y), float64(footer.Dx()), float64(footer.Dy()), layout.FooterRadius*float64(renderScale))
	dc.Fill()

	logoImage, err := decodeBase64Image(input.LogoBase64)
	if err != nil {
		return nil, err
	}
	logoRect := ResolveLogoRect(layout.Width, layout.Footer, template.Logo)
	logoRect = scaleRect(logoRect, renderScale)
	if logoImage != nil {
		resized := resizeImage(logoImage, logoRect.Dx(), logoRect.Dy())
		dc.DrawImage(resized, logoRect.Min.X, logoRect.Min.Y)
	} else {
		dc.SetColor(style.FooterText)
		r.setFont(dc, r.config.Typography.FooterFont, r.config.Typography.FooterSize*float64(renderScale))
		drawCenteredLines(dc, providerName, float64(layout.Width*renderScale)/2, float64(logoRect.Min.Y+logoRect.Dy()/2), layout.TextWrapChars)
	}

	if strings.TrimSpace(input.PayerName) != "" {
		dc.SetColor(style.Muted)
		drawCenteredLines(dc, "Payer: "+strings.TrimSpace(input.PayerName), float64(layout.Width*renderScale)/2, layout.PayerY*float64(renderScale), layout.TextWrapChars)
	}

	finalImage := dc.Image()
	if renderScale > 1 {
		finalImage = resizeImage(finalImage, layout.Width, layout.Height)
	}

	var buf bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.CompressionLevel(r.config.Quality.PNGCompressionLevel)}
	if err := encoder.Encode(&buf, finalImage); err != nil {
		return nil, fmt.Errorf("encode card png: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *Renderer) resolveStyle(base Style, input CardInput) (Style, error) {
	style := base
	var err error

	if strings.TrimSpace(input.BackgroundHex) != "" {
		style.Background, err = ParseHexColor(input.BackgroundHex)
		if err != nil {
			return Style{}, fmt.Errorf("background_color: %w", err)
		}
	}
	if strings.TrimSpace(input.AccentHex) != "" {
		style.Accent, err = ParseHexColor(input.AccentHex)
		if err != nil {
			return Style{}, fmt.Errorf("accent_color: %w", err)
		}
	}
	if strings.TrimSpace(input.TextHex) != "" {
		style.Text, err = ParseHexColor(input.TextHex)
		if err != nil {
			return Style{}, fmt.Errorf("text_color: %w", err)
		}
	}

	return style, nil
}

func (r *Renderer) setFont(dc *gg.Context, role string, size float64) {
	face, err := r.fontManager.Face(role, size)
	if err == nil && face != nil {
		dc.SetFontFace(face)
		return
	}
	dc.SetFontFace(basicfont.Face7x13)
}

func drawCenteredLines(dc *gg.Context, text string, x, y float64, maxChars int) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}

	wrapped := wrapText(trimmed, max(maxChars, 20))
	for index, line := range wrapped {
		dc.DrawStringAnchored(line, x, y+float64(index*28), 0.5, 0.5)
	}
}

func wrapText(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, int(math.Ceil(float64(len(text))/float64(maxChars))))
	var current strings.Builder
	for _, word := range words {
		candidate := word
		if current.Len() > 0 {
			candidate = current.String() + " " + word
		}
		if len(candidate) > maxChars && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			continue
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(word)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	if len(lines) == 0 {
		return []string{text}
	}

	return lines
}

func decodeBase64Image(value string) (image.Image, error) {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return nil, nil
	}

	if comma := strings.Index(raw, ","); strings.HasPrefix(raw, "data:image/") && comma >= 0 {
		raw = raw[comma+1:]
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode logo: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(decoded))
	if err != nil {
		return nil, fmt.Errorf("decode logo image: %w", err)
	}

	return img, nil
}

func resizeImage(src image.Image, width, height int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func scaleRect(rect image.Rectangle, factor int) image.Rectangle {
	return image.Rect(rect.Min.X*factor, rect.Min.Y*factor, rect.Max.X*factor, rect.Max.Y*factor)
}
