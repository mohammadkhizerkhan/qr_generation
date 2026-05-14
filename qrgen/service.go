package qrgen

import (
	"io"

	"github.com/mohammadkhizerkhan/qr_generation/internal/batch"
	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
	"github.com/mohammadkhizerkhan/qr_generation/internal/render"
)

type CardRequest struct {
	UPIURI          string `json:"upi_uri"`
	MerchantName    string `json:"merchant_name"`
	MerchantUPIID   string `json:"merchant_upi_id,omitempty"`
	Description     string `json:"description,omitempty"`
	ProviderName    string `json:"provider_name,omitempty"`
	PayerName       string `json:"payer_name,omitempty"`
	LogoBase64      string `json:"logo_base64,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	AccentColor     string `json:"accent_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
	TemplateID      string `json:"template_id,omitempty"`
}

type Service struct {
	config   *config.Config
	renderer *render.Renderer
	batch    *batch.Service
}

func NewService() *Service {
	defaultConfig := config.DefaultConfig()
	return NewServiceWithConfig(&defaultConfig)
}

func NewServiceWithConfig(cfg *config.Config) *Service {
	if cfg == nil {
		defaultConfig := config.DefaultConfig()
		cfg = &defaultConfig
	}

	renderer := render.NewRendererWithConfig(cfg)
	return &Service{
		config:   cfg,
		renderer: renderer,
		batch:    batch.NewService(renderer, cfg.Batch),
	}
}

func (s *Service) RenderPNG(req CardRequest) ([]byte, error) {
	return s.renderer.RenderPNG(toCardInput(req))
}

func (s *Service) RenderArchive(items []CardRequest) ([]byte, error) {
	converted := make([]render.CardInput, 0, len(items))
	for _, item := range items {
		converted = append(converted, toCardInput(item))
	}
	return s.batch.BuildArchive(converted)
}

func (s *Service) RenderArchiveToWriter(items []CardRequest, writer io.Writer) error {
	converted := make([]render.CardInput, 0, len(items))
	for _, item := range items {
		converted = append(converted, toCardInput(item))
	}
	return s.batch.BuildArchiveToWriter(writer, converted)
}

func (s *Service) StreamBatchEnabled() bool {
	return s.config != nil && s.config.Batch.StreamResponse
}

func toCardInput(req CardRequest) render.CardInput {
	return render.CardInput{
		UPIURI:        req.UPIURI,
		MerchantName:  req.MerchantName,
		MerchantUPIID: req.MerchantUPIID,
		Description:   req.Description,
		ProviderName:  req.ProviderName,
		PayerName:     req.PayerName,
		LogoBase64:    req.LogoBase64,
		BackgroundHex: req.BackgroundColor,
		AccentHex:     req.AccentColor,
		TextHex:       req.TextColor,
		TemplateID:    req.TemplateID,
	}
}
