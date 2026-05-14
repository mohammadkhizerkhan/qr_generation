package render

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
)

type FontManager struct {
	mu    sync.RWMutex
	fonts map[string]*opentype.Font
	faces map[string]font.Face
}

func NewFontManager(cfg config.TypographyConfig) *FontManager {
	fm := &FontManager{
		fonts: make(map[string]*opentype.Font),
		faces: make(map[string]font.Face),
	}

	fm.loadFont("header", cfg.FontPaths.Header)
	fm.loadFont("body", cfg.FontPaths.Body)
	fm.loadFont("footer", cfg.FontPaths.Footer)

	return fm
}

func (fm *FontManager) loadFont(role, path string) {
	if path == "" {
		return
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	fontData, err := opentype.Parse(content)
	if err != nil {
		return
	}

	fm.mu.Lock()
	fm.fonts[role] = fontData
	fm.mu.Unlock()
}

func (fm *FontManager) Face(role string, size float64) (font.Face, error) {
	if size <= 0 {
		return nil, fmt.Errorf("font size must be positive")
	}

	key := fmt.Sprintf("%s:%.1f", role, size)
	fm.mu.RLock()
	if face, ok := fm.faces[key]; ok {
		fm.mu.RUnlock()
		return face, nil
	}
	fontData, ok := fm.fonts[role]
	fm.mu.RUnlock()
	if !ok {
		return nil, nil
	}

	face, err := opentype.NewFace(fontData, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	fm.mu.Lock()
	fm.faces[key] = face
	fm.mu.Unlock()

	return face, nil
}
