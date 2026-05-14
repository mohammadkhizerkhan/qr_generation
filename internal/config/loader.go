package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file %q: %w", path, err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("decode config file %q: %w", path, err)
	}

	normalize(&cfg)
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("validate config file %q: %w", path, err)
	}

	return &cfg, nil
}

func LoadOrDefault(path string) (*Config, error) {
	cfg, err := Load(path)
	if err == nil {
		return cfg, nil
	}

	if !os.IsNotExist(extractRootError(err)) {
		return nil, err
	}

	defaults := DefaultConfig()
	normalize(&defaults)
	return &defaults, nil
}

func extractRootError(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	if strings.Contains(message, "no such file or directory") {
		return os.ErrNotExist
	}
	return err
}

func normalize(cfg *Config) {
	defaults := DefaultConfig()

	if strings.TrimSpace(cfg.Version) == "" {
		cfg.Version = defaults.Version
	}
	if strings.TrimSpace(cfg.DefaultTemplate) == "" {
		cfg.DefaultTemplate = defaults.DefaultTemplate
	}
	if len(cfg.Templates) == 0 {
		cfg.Templates = defaults.Templates
	}

	base := defaults.Templates[defaults.DefaultTemplate]
	if defaultTpl, ok := defaults.Templates[cfg.DefaultTemplate]; ok {
		base = defaultTpl
	}

	for key, template := range cfg.Templates {
		cfg.Templates[key] = mergeTemplate(base, template)
	}

	if cfg.Typography.HeaderSize <= 0 {
		cfg.Typography.HeaderSize = defaults.Typography.HeaderSize
	}
	if cfg.Typography.BodySize <= 0 {
		cfg.Typography.BodySize = defaults.Typography.BodySize
	}
	if cfg.Typography.FooterSize <= 0 {
		cfg.Typography.FooterSize = defaults.Typography.FooterSize
	}
	if cfg.Typography.LineHeight <= 0 {
		cfg.Typography.LineHeight = defaults.Typography.LineHeight
	}
	if strings.TrimSpace(cfg.Typography.HeaderFont) == "" {
		cfg.Typography.HeaderFont = defaults.Typography.HeaderFont
	}
	if strings.TrimSpace(cfg.Typography.BodyFont) == "" {
		cfg.Typography.BodyFont = defaults.Typography.BodyFont
	}
	if strings.TrimSpace(cfg.Typography.FooterFont) == "" {
		cfg.Typography.FooterFont = defaults.Typography.FooterFont
	}

	if cfg.Quality.RenderScale < 1 {
		cfg.Quality.RenderScale = defaults.Quality.RenderScale
	}
	if strings.TrimSpace(cfg.Quality.QRRecoveryLevel) == "" {
		cfg.Quality.QRRecoveryLevel = defaults.Quality.QRRecoveryLevel
	}
	if cfg.Quality.PNGCompressionLevel < -1 || cfg.Quality.PNGCompressionLevel > 9 {
		cfg.Quality.PNGCompressionLevel = defaults.Quality.PNGCompressionLevel
	}

	if cfg.Batch.MaxConcurrency <= 0 {
		cfg.Batch.MaxConcurrency = defaults.Batch.MaxConcurrency
	}
	if cfg.Batch.MaxBatchSize <= 0 {
		cfg.Batch.MaxBatchSize = defaults.Batch.MaxBatchSize
	}
}

func validate(cfg Config) error {
	if len(cfg.Templates) == 0 {
		return fmt.Errorf("templates cannot be empty")
	}
	if _, ok := cfg.Templates[cfg.DefaultTemplate]; !ok {
		return fmt.Errorf("default_template %q is not defined", cfg.DefaultTemplate)
	}

	return nil
}

func mergeTemplate(base TemplateConfig, candidate TemplateConfig) TemplateConfig {
	merged := base

	if strings.TrimSpace(candidate.Name) != "" {
		merged.Name = candidate.Name
	}

	merged.Layout = mergeLayout(base.Layout, candidate.Layout)
	merged.Style = mergeStyle(base.Style, candidate.Style)
	merged.Logo = mergeLogo(base.Logo, candidate.Logo)

	return merged
}

func mergeLayout(base, candidate LayoutConfig) LayoutConfig {
	out := base

	if candidate.Width > 0 {
		out.Width = candidate.Width
	}
	if candidate.Height > 0 {
		out.Height = candidate.Height
	}
	if candidate.HeaderY > 0 {
		out.HeaderY = candidate.HeaderY
	}
	if candidate.UPIIDY > 0 {
		out.UPIIDY = candidate.UPIIDY
	}
	if candidate.DescriptionY > 0 {
		out.DescriptionY = candidate.DescriptionY
	}
	if candidate.HintY > 0 {
		out.HintY = candidate.HintY
	}
	if candidate.PayerY > 0 {
		out.PayerY = candidate.PayerY
	}
	if candidate.QRX > 0 {
		out.QRX = candidate.QRX
	}
	if candidate.QRY > 0 {
		out.QRY = candidate.QRY
	}
	if candidate.QRSize > 0 {
		out.QRSize = candidate.QRSize
	}
	if candidate.TopBarX > 0 {
		out.TopBarX = candidate.TopBarX
	}
	if candidate.TopBarY > 0 {
		out.TopBarY = candidate.TopBarY
	}
	if candidate.TopBarWidth > 0 {
		out.TopBarWidth = candidate.TopBarWidth
	}
	if candidate.TopBarHeight > 0 {
		out.TopBarHeight = candidate.TopBarHeight
	}
	if candidate.PanelX > 0 {
		out.PanelX = candidate.PanelX
	}
	if candidate.PanelY > 0 {
		out.PanelY = candidate.PanelY
	}
	if candidate.PanelWidth > 0 {
		out.PanelWidth = candidate.PanelWidth
	}
	if candidate.PanelHeight > 0 {
		out.PanelHeight = candidate.PanelHeight
	}
	if candidate.FooterX > 0 {
		out.FooterX = candidate.FooterX
	}
	if candidate.FooterY > 0 {
		out.FooterY = candidate.FooterY
	}
	if candidate.FooterWidth > 0 {
		out.FooterWidth = candidate.FooterWidth
	}
	if candidate.FooterHeight > 0 {
		out.FooterHeight = candidate.FooterHeight
	}
	if candidate.CornerRadius > 0 {
		out.CornerRadius = candidate.CornerRadius
	}
	if candidate.TopBarRadius > 0 {
		out.TopBarRadius = candidate.TopBarRadius
	}
	if candidate.FooterRadius > 0 {
		out.FooterRadius = candidate.FooterRadius
	}
	if candidate.TextWrapChars > 0 {
		out.TextWrapChars = candidate.TextWrapChars
	}

	return out
}

func mergeStyle(base, candidate StyleConfig) StyleConfig {
	out := base

	if strings.TrimSpace(candidate.BackgroundHex) != "" {
		out.BackgroundHex = candidate.BackgroundHex
	}
	if strings.TrimSpace(candidate.AccentHex) != "" {
		out.AccentHex = candidate.AccentHex
	}
	if strings.TrimSpace(candidate.TextHex) != "" {
		out.TextHex = candidate.TextHex
	}
	if strings.TrimSpace(candidate.MutedHex) != "" {
		out.MutedHex = candidate.MutedHex
	}
	if strings.TrimSpace(candidate.PanelHex) != "" {
		out.PanelHex = candidate.PanelHex
	}
	if strings.TrimSpace(candidate.FooterTextHex) != "" {
		out.FooterTextHex = candidate.FooterTextHex
	}

	return out
}

func mergeLogo(base, candidate LogoConfig) LogoConfig {
	out := base

	if strings.TrimSpace(candidate.Placement) != "" {
		out.Placement = candidate.Placement
	}
	if candidate.X > 0 {
		out.X = candidate.X
	}
	if candidate.Y > 0 {
		out.Y = candidate.Y
	}
	if candidate.Width > 0 {
		out.Width = candidate.Width
	}
	if candidate.Height > 0 {
		out.Height = candidate.Height
	}
	if candidate.Padding > 0 {
		out.Padding = candidate.Padding
	}

	return out
}
