package qr

import "image"

type ImageGenerator interface {
	Image(content string) (image.Image, error)
	ImageWithIcon(content string, icon image.Image) (image.Image, error)
}
