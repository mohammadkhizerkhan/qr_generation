package http

type GenerateRequest struct {
	UPIURI          string `json:"upi_uri"`
	MerchantName    string `json:"merchant_name"`
	MerchantUPIID   string `json:"merchant_upi_id,omitempty"`
	Description     string `json:"description,omitempty"`
	ProviderName    string `json:"provider_name,omitempty"`
	PayerName       string `json:"payer_name,omitempty"`
	LogoBase64      string `json:"logo_base64,omitempty"`
	QRGenerator     string `json:"qr_generator,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	AccentColor     string `json:"accent_color,omitempty"`
	TextColor       string `json:"text_color,omitempty"`
}

type BatchGenerateRequest struct {
	Items       []GenerateRequest `json:"items"`
	// Concurrency controls the worker pool size for parallel rendering; 0 uses runtime.NumCPU().
	Concurrency int               `json:"concurrency,omitempty"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// MetricsResponse contains timing data and the PNG image for performance analysis
type MetricsResponse struct {
	ImageBase64              string             `json:"image_base64"` // PNG as base64
	QRGenerationDurationMs   float64            `json:"qr_generation_duration_ms"`
	TemplateRenderDurationMs float64            `json:"template_render_duration_ms"`
	TotalRenderDurationMs    float64            `json:"total_render_duration_ms"`
	QRGeneratorUsed          string             `json:"qr_generator_used"`
	RenderMode               string             `json:"render_mode"`
	GeneratorTimingsMs       map[string]float64 `json:"generator_timings_ms"` // Timing breakdown for all generators
}
