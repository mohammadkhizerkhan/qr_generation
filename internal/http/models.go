package http

type GenerateRequest struct {
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

type BatchGenerateRequest struct {
	Items []GenerateRequest `json:"items"`
}

type errorResponse struct {
	Error string `json:"error"`
}
