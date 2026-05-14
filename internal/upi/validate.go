package upi

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrEmptyURI      = errors.New("upi uri is required")
	ErrInvalidScheme = errors.New("upi uri must use upi:// scheme")
	ErrInvalidAction = errors.New("upi uri must target upi://pay")
)

type PaymentURI struct {
	Raw        string
	MerchantID string
	URL        *url.URL
	Params     url.Values
}

func Validate(raw string) (PaymentURI, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return PaymentURI{}, ErrEmptyURI
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return PaymentURI{}, fmt.Errorf("parse upi uri: %w", err)
	}
	if !strings.EqualFold(parsed.Scheme, "upi") {
		return PaymentURI{}, ErrInvalidScheme
	}

	action := strings.Trim(parsed.Host+parsed.Path, "/")
	if !strings.EqualFold(action, "pay") {
		return PaymentURI{}, ErrInvalidAction
	}

	params := parsed.Query()
	merchantID := strings.TrimSpace(params.Get("pa"))
	if merchantID == "" {
		return PaymentURI{}, errors.New("upi uri is missing pa parameter")
	}

	return PaymentURI{
		Raw:        trimmed,
		MerchantID: merchantID,
		URL:        parsed,
		Params:     params,
	}, nil
}
