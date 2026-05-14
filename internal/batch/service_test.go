package batch

import (
	"archive/zip"
	"bytes"
	"testing"
	"time"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
	"github.com/mohammadkhizerkhan/qr_generation/internal/render"
)

func TestBuildArchive(t *testing.T) {
	t.Parallel()

	svc := NewService(render.NewRenderer(), config.DefaultConfig().Batch)
	svc.now = func() time.Time {
		return time.Date(2026, time.May, 13, 8, 30, 0, 0, time.UTC)
	}

	archiveData, err := svc.BuildArchive([]render.CardInput{{
		UPIURI:       "upi://pay?pa=merchant%40bank&pn=Simba%20Pvt%20Ltd",
		MerchantName: "Simba Pvt Ltd",
		PayerName:    "Khizer Khan",
	}})
	if err != nil {
		t.Fatalf("BuildArchive() error = %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(archiveData), int64(len(archiveData)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}
	if len(reader.File) != 1 {
		t.Fatalf("zip entries = %d, want 1", len(reader.File))
	}
	if got, want := reader.File[0].Name, "khizer-khan-20260513-083000-01.png"; got != want {
		t.Fatalf("zip entry name = %q, want %q", got, want)
	}
}
