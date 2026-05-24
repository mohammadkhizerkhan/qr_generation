package render

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"

	"github.com/mohammadkhizerkhan/qr_generation/assets"
	"github.com/mohammadkhizerkhan/qr_generation/internal/qr"
	"github.com/mohammadkhizerkhan/qr_generation/internal/upi"
)

type CardInput struct {
	UPIURI        string
	MerchantName  string
	MerchantUPIID string
	Description   string
	RenderTemplate string
	PayerName     string
	LogoBase64    string
	BackgroundHex string
	AccentHex     string
	TextHex       string
	QRGenerator   string
}

// GenerationMetrics holds timing data for QR code generation
type GenerationMetrics struct {
	QRGenerationDurationMs   float64            // Time taken to generate QR code
	TemplateRenderDurationMs float64            // Time taken to compose the template image
	ValidationDurationMs     float64            // Time taken for input and UPI validation
	DrawDurationMs           float64            // Time taken to draw template layers
	EncodeDurationMs         float64            // Time taken to PNG-encode output
	TotalRenderDurationMs    float64            // Total time to render entire PNG
	QRGeneratorUsed          string             // Which QR generator was used (skip2, yeqown, piglig)
	RenderMode               string             // prepared or legacy
	GeneratorTimingsMs       map[string]float64 // Individual timings for each generator (for comparison)
}

type PreparedTemplate struct {
	base        *image.RGBA
	brand       *image.RGBA
	layout      Layout
	style       Style
	description string
}

type Renderer struct {
	skipQR        *qr.Generator
	yeqownQR      *qr.YeqownGenerator
	pigligQR      *qr.PigligGenerator
	prepared      *PreparedTemplate
	preparedQR    *PreparedTemplate
	layout        Layout
	style         Style
	merchantFont  *truetype.Font
	secondaryFont *truetype.Font
	logPerf       bool
}

func NewRenderer() *Renderer {
	layout := DefaultLayout()
	style := DefaultStyle()
	merchantFont, secondaryFont, fontErr := loadInterFonts()
	if fontErr != nil {
		log.Printf("qr_renderer inter_font=false err=%v fallback=basicfont", fontErr)
		merchantFont = nil
		secondaryFont = nil
	}

	secondaryFace := createFace(secondaryFont)
	defer closeFace(secondaryFace)
	prepareStart := time.Now()
	prepared, err := prepareTemplate(layout, style, secondaryFace)
	if err != nil {
		panic(err)
	}
	preparedQR, err := prepareTemplateQR(layout, style)
	if err != nil {
		panic(err)
	}
	prepareDuration := time.Since(prepareStart)
	logPerf := os.Getenv("QRGEN_PROFILE") == "1"
	if logPerf {
		log.Printf("qr_renderer startup prepared_template=true prepared_qr_template=true width=%d height=%d brand=%t prepare_ms=%.2f", layout.Width, layout.Height, prepared != nil, float64(prepareDuration.Microseconds())/1000)
	}

	return &Renderer{
		skipQR:        qr.NewGenerator(layout.QRSize),
		yeqownQR:      qr.NewYeqownGenerator(layout.QRSize),
		pigligQR:      qr.NewPigligGenerator(layout.QRSize),
		prepared:      prepared,
		preparedQR:    preparedQR,
		layout:        layout,
		style:         style,
		merchantFont:  merchantFont,
		secondaryFont: secondaryFont,
		logPerf:       logPerf,
	}
}

func (r *Renderer) RenderPNG(input CardInput) ([]byte, error) {
	png, _, err := r.RenderPNGWithMetrics(input)
	return png, err
}

// RenderPNGWithMetrics is like RenderPNG but also returns timing metrics for QR generation
func (r *Renderer) RenderPNGWithMetrics(input CardInput) ([]byte, *GenerationMetrics, error) {
	begin := time.Now()
	validationStart := time.Now()
	metrics := &GenerationMetrics{
		GeneratorTimingsMs: make(map[string]float64),
	}

	validated, err := upi.Validate(input.UPIURI)
	if err != nil {
		return nil, nil, err
	}

	merchantName := strings.TrimSpace(input.MerchantName)
	if merchantName == "" {
		merchantName = validated.Params.Get("pn")
	}
	if merchantName == "" {
		return nil, nil, fmt.Errorf("merchant_name is required")
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
		description = "Scan this QR code with any UPI app to transfer"
	}
	metrics.ValidationDurationMs = float64(time.Since(validationStart).Microseconds()) / 1000

	qrImage, qrMetrics, err := r.qrImageWithMetrics(validated.Raw, input.QRGenerator)
	// qrImage, err := r.skipQR.Image(validated.Raw)
	if err != nil {
		return nil, nil, err
	}
	metrics.QRGenerationDurationMs = qrMetrics.QRGenerationDurationMs
	metrics.QRGeneratorUsed = qrMetrics.QRGeneratorUsed
	metrics.GeneratorTimingsMs = qrMetrics.GeneratorTimingsMs

	if r.useTemplateQR(input.RenderTemplate) && r.canUsePreparedQRTemplate() {
		pngData, drawDuration, encodeDuration, templateDuration, err := r.renderPreparedTemplateQRPNG(merchantName, merchantUPIID, qrImage)
		if err != nil {
			return nil, nil, err
		}
		metrics.DrawDurationMs = drawDuration
		metrics.EncodeDurationMs = encodeDuration
		metrics.TemplateRenderDurationMs = templateDuration
		metrics.RenderMode = "template_qr_prepared"
		metrics.TotalRenderDurationMs = float64(time.Since(begin).Microseconds()) / 1000
		r.logMetrics(metrics)
		return pngData, metrics, nil
	}

	if r.canUsePreparedTemplate() {
		pngData, drawDuration, encodeDuration, templateDuration, err := r.renderPreparedPNG(merchantName, merchantUPIID, qrImage)
		if err != nil {
			return nil, nil, err
		}
		metrics.DrawDurationMs = drawDuration
		metrics.EncodeDurationMs = encodeDuration
		metrics.TemplateRenderDurationMs = templateDuration
		metrics.RenderMode = "prepared"
		metrics.TotalRenderDurationMs = float64(time.Since(begin).Microseconds()) / 1000
		r.logMetrics(metrics)
		return pngData, metrics, nil
	}
	metrics.RenderMode = "legacy"

	style, err := r.resolveStyle(input)
	if err != nil {
		return nil, nil, err
	}

	drawStart := time.Now()
	dc := gg.NewContext(r.layout.Width, r.layout.Height)
	dc.SetColor(style.Background)
	dc.Clear()
	merchantFace, secondaryFace := r.newRenderFaces()
	defer closeFace(merchantFace)
	defer closeFace(secondaryFace)

	dc.SetColor(style.Text)
	dc.SetFontFace(merchantFace)
	drawCenteredLines(dc, strings.ToUpper(merchantName), float64(r.layout.Width)/2, r.layout.HeaderY, 3)

	dc.SetColor(style.Muted)
	dc.SetFontFace(secondaryFace)
	drawCenteredLines(dc, "My UPI ID: "+merchantUPIID, float64(r.layout.Width)/2, r.layout.UPIIDY, 2)

	dc.SetColor(style.Muted)
	dc.SetLineWidth(1.2)
	dc.DrawLine(260, r.layout.DividerY, float64(r.layout.Width-260), r.layout.DividerY)
	dc.Stroke()

	dc.SetFontFace(secondaryFace)
	drawCenteredLines(dc, description, float64(r.layout.Width)/2, r.layout.DescriptionY, 2)

	dc.DrawImage(qrImage, r.layout.QRX, r.layout.QRY)

	if r.prepared != nil && r.prepared.brand != nil {
		dc.DrawImage(r.prepared.brand, r.layout.BrandX, r.layout.BrandY)
	}
	metrics.DrawDurationMs = float64(time.Since(drawStart).Microseconds()) / 1000

	var buf bytes.Buffer
	encodeStart := time.Now()
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, nil, fmt.Errorf("encode card png: %w", err)
	}
	metrics.EncodeDurationMs = float64(time.Since(encodeStart).Microseconds()) / 1000

	metrics.TemplateRenderDurationMs = metrics.DrawDurationMs + metrics.EncodeDurationMs
	metrics.TotalRenderDurationMs = float64(time.Since(begin).Microseconds()) / 1000
	r.logMetrics(metrics)
	return buf.Bytes(), metrics, nil
}

func (r *Renderer) canUsePreparedTemplate() bool {
	return r.prepared != nil
}

func (r *Renderer) canUsePreparedQRTemplate() bool {
	return r.preparedQR != nil
}

func (r *Renderer) useTemplateQR(renderTemplate string) bool {
	return strings.EqualFold(strings.TrimSpace(renderTemplate), "template_qr")
}

func (r *Renderer) renderPreparedPNG(merchantName, merchantUPIID string, qrImage image.Image) ([]byte, float64, float64, float64, error) {
	log.Printf("Using prepared_instance from service")
	begin := time.Now()
	base := cloneRGBA(r.prepared.base)
	dc := gg.NewContextForRGBA(base)
	drawStart := time.Now()
	merchantFace, secondaryFace := r.newRenderFaces()
	defer closeFace(merchantFace)
	defer closeFace(secondaryFace)

	dc.SetColor(r.prepared.style.Text)
	dc.SetFontFace(merchantFace)
	drawCenteredLines(dc, strings.ToUpper(merchantName), float64(r.prepared.layout.Width)/2, r.prepared.layout.HeaderY, 3)

	dc.SetColor(r.prepared.style.Muted)
	dc.SetFontFace(secondaryFace)
	drawCenteredLines(dc, "My UPI ID: "+merchantUPIID, float64(r.prepared.layout.Width)/2, r.prepared.layout.UPIIDY, 2)

	dc.DrawImage(qrImage, r.prepared.layout.QRX, r.prepared.layout.QRY)
	drawDuration := float64(time.Since(drawStart).Microseconds()) / 1000

	var buf bytes.Buffer
	encodeStart := time.Now()
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("encode prepared png: %w", err)
	}
	encodeDuration := float64(time.Since(encodeStart).Microseconds()) / 1000

	return buf.Bytes(), drawDuration, encodeDuration, float64(time.Since(begin).Microseconds()) / 1000, nil
}

func (r *Renderer) renderPreparedTemplateQRPNG(merchantName, merchantUPIID string, qrImage image.Image) ([]byte, float64, float64, float64, error) {
	log.Printf("Using template_qr from assets")
	begin := time.Now()
	base := cloneRGBA(r.preparedQR.base)
	dc := gg.NewContextForRGBA(base)
	drawStart := time.Now()
	merchantFace, secondaryFace := r.newRenderFaces()
	defer closeFace(merchantFace)
	defer closeFace(secondaryFace)

	dc.SetColor(r.preparedQR.style.Text)
	dc.SetFontFace(merchantFace)
	drawCenteredLines(dc, strings.ToUpper(merchantName), float64(r.preparedQR.layout.Width)/2, r.preparedQR.layout.HeaderY, 3)

	dc.SetColor(r.preparedQR.style.Muted)
	dc.SetFontFace(secondaryFace)
	drawCenteredLines(dc, "My UPI ID: "+merchantUPIID, float64(r.preparedQR.layout.Width)/2, r.preparedQR.layout.UPIIDY, 2)

	qrBounds := qrImage.Bounds()
	dstRect := image.Rect(
		r.preparedQR.layout.QRX, r.preparedQR.layout.QRY,
		r.preparedQR.layout.QRX+qrBounds.Dx(), r.preparedQR.layout.QRY+qrBounds.Dy(),
	)
	draw.Draw(base, dstRect, qrImage, qrBounds.Min, draw.Src)
	drawDuration := float64(time.Since(drawStart).Microseconds()) / 1000

	var buf bytes.Buffer
	encodeStart := time.Now()
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, 0, 0, 0, fmt.Errorf("encode template_qr prepared png: %w", err)
	}
	encodeDuration := float64(time.Since(encodeStart).Microseconds()) / 1000

	return buf.Bytes(), drawDuration, encodeDuration, float64(time.Since(begin).Microseconds()) / 1000, nil
}

func prepareTemplate(layout Layout, style Style, secondaryFace font.Face) (*PreparedTemplate, error) {
	base := image.NewRGBA(image.Rect(0, 0, layout.Width, layout.Height))
	dc := gg.NewContextForRGBA(base)
	dc.SetColor(style.Background)
	dc.Clear()

	dc.SetColor(style.Muted)
	dc.SetLineWidth(1)
	dc.DrawLine(260, layout.DividerY, float64(layout.Width-260), layout.DividerY)
	dc.Stroke()

	dc.SetColor(style.Muted)
	dc.SetFontFace(secondaryFace)
	drawCenteredLines(dc, "Scan this QR code with any UPI app to transfer", float64(layout.Width)/2, layout.DescriptionY, 2)

	brandImage, err := decodeAssetImage(assets.IDFCBrand)
	if err != nil {
		return nil, err
	}
	resizedBrand := resizeImage(brandImage, layout.BrandWidth, layout.BrandHeight)
	dc.DrawImage(resizedBrand, layout.BrandX, layout.BrandY)

	return &PreparedTemplate{
		base:        base,
		brand:       resizedBrand,
		layout:      layout,
		style:       style,
		description: "Scan this QR code with any UPI app to transfer",
	}, nil
}

func prepareTemplateQR(layout Layout, style Style) (*PreparedTemplate, error) {
	templateImage, err := decodeAssetImage(assets.TemplateQR)
	if err != nil {
		return nil, err
	}

	templateBase := resizeImage(templateImage, layout.Width, layout.Height)
	return &PreparedTemplate{
		base:   templateBase,
		layout: layout,
		style:  style,
	}, nil
}

func loadInterFonts() (*truetype.Font, *truetype.Font, error) {
	regularFont, err := truetype.Parse(assets.InterRegular)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inter regular: %w", err)
	}
	boldFont, err := truetype.Parse(assets.InterBold)
	if err != nil {
		return nil, nil, fmt.Errorf("parse inter bold: %w", err)
	}
	return boldFont, regularFont, nil
}

func createFace(parsed *truetype.Font) font.Face {
	if parsed == nil {
		return basicfont.Face7x13
	}
	return truetype.NewFace(parsed, &truetype.Options{
		Size:    16,
		DPI:     72,
		Hinting: font.HintingFull,
	})
}

func closeFace(face font.Face) {
	closer, ok := face.(interface{ Close() error })
	if !ok {
		return
	}
	_ = closer.Close()
}

func (r *Renderer) newRenderFaces() (font.Face, font.Face) {
	merchantFace := createFace(r.merchantFont)
	secondaryFace := createFace(r.secondaryFont)
	return merchantFace, secondaryFace
}

func (r *Renderer) logMetrics(metrics *GenerationMetrics) {
	if !r.logPerf {
		return
	}
	log.Printf("qr_renderer mode=%s total_ms=%.2f validate_ms=%.2f qr_ms=%.2f draw_ms=%.2f encode_ms=%.2f template_ms=%.2f generator=%s", metrics.RenderMode, metrics.TotalRenderDurationMs, metrics.ValidationDurationMs, metrics.QRGenerationDurationMs, metrics.DrawDurationMs, metrics.EncodeDurationMs, metrics.TemplateRenderDurationMs, metrics.QRGeneratorUsed)
}

func (r *Renderer) qrImage(content, backend string) (image.Image, error) {
	img, _, err := r.qrImageWithMetrics(content, backend)
	return img, err
}

// qrImageWithMetrics generates QR code and returns timing metrics for all generators
func (r *Renderer) qrImageWithMetrics(content, backend string) (image.Image, *GenerationMetrics, error) {
	selected := strings.ToLower(strings.TrimSpace(backend))
	metrics := &GenerationMetrics{
		GeneratorTimingsMs: make(map[string]float64),
	}

	// Benchmark all generators for comparison
	generators := map[string]func() (image.Image, error){
		"skip2": func() (image.Image, error) {
			startTime := time.Now()
			defer func() {
				metrics.GeneratorTimingsMs["skip2"] = time.Since(startTime).Seconds() * 1000
			}()
			return r.skipQR.Image(content)
		},
		"yeqown": func() (image.Image, error) {
			startTime := time.Now()
			defer func() {
				metrics.GeneratorTimingsMs["yeqown"] = time.Since(startTime).Seconds() * 1000
			}()
			_, err := qr.DefaultLogo()
			if err != nil {
				return nil, err
			}
			return r.yeqownQR.Image(content)
		},
		"piglig": func() (image.Image, error) {
			startTime := time.Now()
			defer func() {
				metrics.GeneratorTimingsMs["piglig"] = time.Since(startTime).Seconds() * 1000
			}()
			_, err := qr.DefaultLogo()
			if err != nil {
				return nil, err
			}
			return r.pigligQR.Image(content)
		},
	}

	// If no backend selected, use skip2 as default
	if selected == "" || selected == "skip2" {
		selected = "skip2"
	}

	// Validate selected generator
	if selected != "skip2" && selected != "yeqown" && selected != "piglig" {
		return nil, nil, fmt.Errorf("unknown qr_generator %q", backend)
	}

	// Generate with selected backend
	generatorFunc, ok := generators[selected]
	if !ok {
		return nil, nil, fmt.Errorf("generator %q not available", selected)
	}

	img, err := generatorFunc()
	if err != nil {
		return nil, nil, err
	}

	metrics.QRGeneratorUsed = selected
	if timing, exists := metrics.GeneratorTimingsMs[selected]; exists {
		metrics.QRGenerationDurationMs = timing
	}

	return img, metrics, nil
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

func decodeAssetImage(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("asset image is empty")
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode asset image: %w", err)
	}

	return img, nil
}

func cloneRGBA(src *image.RGBA) *image.RGBA {
	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, dst.Bounds(), src, src.Bounds().Min, draw.Src)
	return dst
}

func resizeImage(src image.Image, width, height int) *image.RGBA {
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
