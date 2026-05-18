package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font/basicfont"

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
	QRGenerator   string
}

type Renderer struct {
	skipQR   *qr.Generator
	yeqownQR *qr.YeqownGenerator
	layout   Layout
	style    Style
}

func NewRenderer() *Renderer {
	return &Renderer{
		skipQR:   qr.NewGenerator(DefaultLayout().QRSize),
		yeqownQR: qr.NewYeqownGenerator(DefaultLayout().QRSize),
		layout:   DefaultLayout(),
		style:    DefaultStyle(),
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

	style, err := r.resolveStyle(input)
	if err != nil {
		return nil, err
	}

	qrImage, err := r.qrImage(validated.Raw, input.QRGenerator)
	if err != nil {
		return nil, err
	}

	dc := gg.NewContext(r.layout.Width, r.layout.Height)
	dc.SetColor(style.Background)
	dc.Clear()

	dc.SetColor(style.Accent)
	dc.DrawRoundedRectangle(50, 55, float64(r.layout.Width-100), 40, 18)
	dc.Fill()

	dc.SetColor(style.Panel)
	dc.DrawRoundedRectangle(90, 380, float64(r.layout.Width-180), float64(r.layout.QRSize+90), 30)
	dc.Fill()

	dc.SetColor(style.Text)
	dc.SetFontFace(basicfont.Face7x13)
	drawCenteredLines(dc, strings.ToUpper(merchantName), float64(r.layout.Width)/2, r.layout.HeaderY, 3)

	dc.SetColor(style.Muted)
	drawCenteredLines(dc, "UPI ID: "+merchantUPIID, float64(r.layout.Width)/2, r.layout.UPIIDY, 2)
	drawCenteredLines(dc, description, float64(r.layout.Width)/2, r.layout.DescriptionY, 2)
	drawCenteredLines(dc, "Scan with any supported UPI app", float64(r.layout.Width)/2, 365, 2)

	dc.DrawImage(qrImage, r.layout.QRX, r.layout.QRY)

	dc.SetColor(style.Accent)
	dc.DrawRoundedRectangle(200, float64(r.layout.FooterY-25), 500, 150, 26)
	dc.Fill()

	logoImage, err := decodeBase64Image(input.LogoBase64)
	if err != nil {
		return nil, err
	}
	if logoImage != nil {
		resized := resizeImage(logoImage, r.layout.LogoWidth, r.layout.LogoHeight)
		dc.DrawImage(resized, r.layout.LogoX, r.layout.LogoY)
	} else {
		dc.SetColor(color.White)
		drawCenteredLines(dc, providerName, float64(r.layout.Width)/2, float64(r.layout.FooterY+40), 2)
	}

	if strings.TrimSpace(input.PayerName) != "" {
		dc.SetColor(style.Muted)
		drawCenteredLines(dc, "Payer: "+strings.TrimSpace(input.PayerName), float64(r.layout.Width)/2, 1110, 2)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode card png: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *Renderer) qrImage(content, backend string) (image.Image, error) {
	selected := strings.ToLower(strings.TrimSpace(backend))
	if selected == "" || selected == "skip2" {
		return r.skipQR.Image(content)
	}

	if selected != "yeqown" {
		return nil, fmt.Errorf("unknown qr_generator %q", backend)
	}

	icon, err := qr.DefaultLogo()
	if err != nil {
		return nil, err
	}

	return r.yeqownQR.ImageWithIcon(content, icon)
}

func (r *Renderer) resolveStyle(input CardInput) (Style, error) {
	style := r.style
	var err error

	if strings.TrimSpace(input.BackgroundHex) != "" {
		style.Background, err = parseHexColor(input.BackgroundHex)
		if err != nil {
			return Style{}, fmt.Errorf("background_color: %w", err)
		}
	}
	if strings.TrimSpace(input.AccentHex) != "" {
		style.Accent, err = parseHexColor(input.AccentHex)
		if err != nil {
			return Style{}, fmt.Errorf("accent_color: %w", err)
		}
	}
	if strings.TrimSpace(input.TextHex) != "" {
		style.Text, err = parseHexColor(input.TextHex)
		if err != nil {
			return Style{}, fmt.Errorf("text_color: %w", err)
		}
	}

	return style, nil
}

func drawCenteredLines(dc *gg.Context, text string, x, y float64, lines int) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return
	}

	wrapped := wrapText(trimmed, 34*max(lines, 1))
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

func parseHexColor(value string) (color.Color, error) {
	hex := strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(hex) != 6 {
		return nil, fmt.Errorf("must be a 6-digit hex color")
	}

	var rgb [3]uint8
	for index := 0; index < 3; index++ {
		var channel uint8
		if _, err := fmt.Sscanf(hex[index*2:index*2+2], "%02x", &channel); err != nil {
			return nil, fmt.Errorf("invalid hex color")
		}
		rgb[index] = channel
	}

	return color.RGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 255}, nil
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
