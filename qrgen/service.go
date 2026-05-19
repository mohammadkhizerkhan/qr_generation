package qrgen

import (
	"github.com/mohammadkhizerkhan/qr_generation/internal/batch"
	"github.com/mohammadkhizerkhan/qr_generation/internal/render"
)

type CardRequest struct {
	UPIURI          string `json:"upi_uri"`
	MerchantName    string `json:"merchant_name"`
	MerchantUPIID   string `json:"merchant_upi_id,omitempty"`
	Description     string `json:"description,omitempty"`
	PayerName       string `json:"payer_name,omitempty"`
	LogoBase64      string `json:"logo_base64,omitempty"`
	QRGenerator     string `json:"qr_generator,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	AccentColor     string `json:"accent_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

type Service struct {
	renderer *render.Renderer
	batch    *batch.Service
}

func NewService() *Service {
	renderer := render.NewRenderer()
	return &Service{
		renderer: renderer,
		batch:    batch.NewService(renderer),
	}
}

func (s *Service) RenderPNG(req CardRequest) ([]byte, error) {
	return s.renderer.RenderPNG(toCardInput(req))
}

// RenderPNGWithMetrics returns PNG bytes along with timing metrics
func (s *Service) RenderPNGWithMetrics(req CardRequest) ([]byte, *render.GenerationMetrics, error) {
	return s.renderer.RenderPNGWithMetrics(toCardInput(req))
}

// RenderArchive renders all items concurrently and returns a ZIP archive.
// concurrency controls the worker pool size; 0 defaults to runtime.NumCPU().
func (s *Service) RenderArchive(items []CardRequest, concurrency int) ([]byte, error) {
	converted := make([]render.CardInput, 0, len(items))
	for _, item := range items {
		converted = append(converted, toCardInput(item))	
	}
	return s.batch.BuildArchive(converted, concurrency)
}

func toCardInput(req CardRequest) render.CardInput {
	return render.CardInput{
		UPIURI:        req.UPIURI,
		MerchantName:  req.MerchantName,
		MerchantUPIID: req.MerchantUPIID,
		Description:   req.Description,
		PayerName:     req.PayerName,
		LogoBase64:    req.LogoBase64,
		QRGenerator:   req.QRGenerator,
		BackgroundHex: req.BackgroundColor,
		AccentHex:     req.AccentColor,
		TextHex:       req.TextColor,
	}
}
