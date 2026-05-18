package qr

import (
	"bytes"
	"fmt"
	"image"
	_ "image/png"
	"sync"

	"github.com/mohammadkhizerkhan/qr_generation/assets"
)

var (
	defaultLogoOnce sync.Once
	defaultLogo     image.Image
	defaultLogoErr  error
)

func DefaultLogo() (image.Image, error) {
	defaultLogoOnce.Do(func() {
		img, _, err := image.Decode(bytes.NewReader(assets.IDFCLogo))
		if err != nil {
			defaultLogoErr = fmt.Errorf("decode idfc_logo.png: %w", err)
			return
		}
		defaultLogo = img
	})

	return defaultLogo, defaultLogoErr
}
