package upi

import "testing"

func TestValidate(t *testing.T) {
	t.Parallel()

	parsed, err := Validate("upi://pay?pa=merchant%40bank&pn=Shop")
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if parsed.MerchantID != "merchant@bank" {
		t.Fatalf("MerchantID = %q, want merchant@bank", parsed.MerchantID)
	}
}

func TestValidateRejectsInvalidScheme(t *testing.T) {
	t.Parallel()

	if _, err := Validate("https://pay?pa=merchant%40bank"); err != ErrInvalidScheme {
		t.Fatalf("Validate() error = %v, want %v", err, ErrInvalidScheme)
	}
}