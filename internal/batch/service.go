package batch

import (
	"archive/zip"
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mohammadkhizerkhan/qr_generation/internal/render"
)

var invalidFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

type Service struct {
	renderer *render.Renderer
	now      func() time.Time
}

func NewService(renderer *render.Renderer) *Service {
	return &Service{
		renderer: renderer,
		now:      time.Now,
	}
}

type renderResult struct {
	pngData  []byte
	filename string
	err      error
}

// BuildArchive renders all items concurrently then assembles the ZIP serially.
// concurrency controls the worker pool size; 0 defaults to runtime.NumCPU().
func (s *Service) BuildArchive(items []render.CardInput, concurrency int) ([]byte, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("at least one batch item is required")
	}

	timestamp := s.now().UTC().Format("20060102-150405")
	results := make([]renderResult, len(items))

	runWorkers(len(items), concurrency, func(i int) {
		pngData, err := s.renderer.RenderPNG(items[i])
		results[i] = renderResult{
			pngData:  pngData,
			filename: buildFileName(items[i], timestamp, i+1),
			err:      err,
		}
	})

	var buf bytes.Buffer
	archive := zip.NewWriter(&buf)
	for i, result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("render item %d: %w", i, result.err)
		}
		fileWriter, err := archive.Create(result.filename)
		if err != nil {
			return nil, fmt.Errorf("create zip entry: %w", err)
		}
		if _, err := fileWriter.Write(result.pngData); err != nil {
			return nil, fmt.Errorf("write zip entry: %w", err)
		}
	}

	if err := archive.Close(); err != nil {
		return nil, fmt.Errorf("close zip archive: %w", err)
	}

	return buf.Bytes(), nil
}

func buildFileName(item render.CardInput, timestamp string, index int) string {
	base := firstNonEmpty(item.PayerName, item.MerchantName, item.MerchantUPIID, "upi-qr")
	sanitized := sanitize(base)
	return fmt.Sprintf("%s-%s-%02d.png", sanitized, timestamp, index)
}

func sanitize(value string) string {
	cleaned := strings.Trim(strings.ToLower(value), " ")
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = invalidFilenameChars.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		return "upi-qr"
	}
	return cleaned
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}