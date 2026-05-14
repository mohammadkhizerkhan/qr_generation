package config

type Config struct {
	Version         string                    `json:"version"`
	DefaultTemplate string                    `json:"default_template"`
	Templates       map[string]TemplateConfig `json:"templates"`
	Typography      TypographyConfig          `json:"typography"`
	Quality         QualityConfig             `json:"quality"`
	Batch           BatchConfig               `json:"batch"`
}

type TemplateConfig struct {
	Name   string       `json:"name"`
	Layout LayoutConfig `json:"layout"`
	Style  StyleConfig  `json:"style"`
	Logo   LogoConfig   `json:"logo"`
}

type LayoutConfig struct {
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	HeaderY       float64 `json:"header_y"`
	UPIIDY        float64 `json:"upi_id_y"`
	DescriptionY  float64 `json:"description_y"`
	HintY         float64 `json:"hint_y"`
	PayerY        float64 `json:"payer_y"`
	QRX           int     `json:"qr_x"`
	QRY           int     `json:"qr_y"`
	QRSize        int     `json:"qr_size"`
	TopBarX       int     `json:"top_bar_x"`
	TopBarY       int     `json:"top_bar_y"`
	TopBarWidth   int     `json:"top_bar_width"`
	TopBarHeight  int     `json:"top_bar_height"`
	PanelX        int     `json:"panel_x"`
	PanelY        int     `json:"panel_y"`
	PanelWidth    int     `json:"panel_width"`
	PanelHeight   int     `json:"panel_height"`
	FooterX       int     `json:"footer_x"`
	FooterY       int     `json:"footer_y"`
	FooterWidth   int     `json:"footer_width"`
	FooterHeight  int     `json:"footer_height"`
	CornerRadius  float64 `json:"corner_radius"`
	TopBarRadius  float64 `json:"top_bar_radius"`
	FooterRadius  float64 `json:"footer_radius"`
	TextWrapChars int     `json:"text_wrap_chars"`
}

type StyleConfig struct {
	BackgroundHex string `json:"background_color"`
	AccentHex     string `json:"accent_color"`
	TextHex       string `json:"text_color"`
	MutedHex      string `json:"muted_color"`
	PanelHex      string `json:"panel_color"`
	FooterTextHex string `json:"footer_text_color"`
}

type LogoConfig struct {
	Placement string `json:"placement"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Padding   int    `json:"padding"`
}

type TypographyConfig struct {
	HeaderFont string  `json:"header_font"`
	BodyFont   string  `json:"body_font"`
	FooterFont string  `json:"footer_font"`
	HeaderSize float64 `json:"header_size"`
	BodySize   float64 `json:"body_size"`
	FooterSize float64 `json:"footer_size"`
	LineHeight float64 `json:"line_height"`
	FontPaths  FontMap `json:"font_paths"`
}

type FontMap struct {
	Header string `json:"header"`
	Body   string `json:"body"`
	Footer string `json:"footer"`
}

type QualityConfig struct {
	QRRecoveryLevel     string `json:"qr_recovery_level"`
	PNGCompressionLevel int    `json:"png_compression_level"`
	RenderScale         int    `json:"render_scale"`
}

type BatchConfig struct {
	MaxConcurrency int  `json:"max_concurrency"`
	MaxBatchSize   int  `json:"max_batch_size"`
	StreamResponse bool `json:"stream_response"`
}

func DefaultConfig() Config {
	return Config{
		Version:         "1.0",
		DefaultTemplate: "standard",
		Templates: map[string]TemplateConfig{
			"standard": {
				Name: "Standard",
				Layout: LayoutConfig{
					Width:         900,
					Height:        1400,
					HeaderY:       120,
					UPIIDY:        220,
					DescriptionY:  320,
					HintY:         365,
					PayerY:        1110,
					QRX:           170,
					QRY:           420,
					QRSize:        560,
					TopBarX:       50,
					TopBarY:       55,
					TopBarWidth:   800,
					TopBarHeight:  40,
					PanelX:        90,
					PanelY:        380,
					PanelWidth:    720,
					PanelHeight:   650,
					FooterX:       200,
					FooterY:       1135,
					FooterWidth:   500,
					FooterHeight:  150,
					CornerRadius:  30,
					TopBarRadius:  18,
					FooterRadius:  26,
					TextWrapChars: 68,
				},
				Style: StyleConfig{
					BackgroundHex: "#FBF8F1",
					AccentHex:     "#9F1A1A",
					TextHex:       "#26221D",
					MutedHex:      "#675E56",
					PanelHex:      "#FFFFFF",
					FooterTextHex: "#FFFFFF",
				},
				Logo: LogoConfig{
					Placement: "bottom-center",
					X:         290,
					Y:         1195,
					Width:     320,
					Height:    110,
					Padding:   20,
				},
			},
		},
		Typography: TypographyConfig{
			HeaderFont: "header",
			BodyFont:   "body",
			FooterFont: "footer",
			HeaderSize: 30,
			BodySize:   20,
			FooterSize: 24,
			LineHeight: 1.2,
			FontPaths:  FontMap{},
		},
		Quality: QualityConfig{
			QRRecoveryLevel:     "M",
			PNGCompressionLevel: -1,
			RenderScale:         1,
		},
		Batch: BatchConfig{
			MaxConcurrency: 4,
			MaxBatchSize:   500,
			StreamResponse: false,
		},
	}
}
