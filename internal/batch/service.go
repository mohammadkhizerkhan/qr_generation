package batch

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
	"github.com/mohammadkhizerkhan/qr_generation/internal/render"
)

var invalidFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

type Service struct {
	renderer       *render.Renderer
	now            func() time.Time
	maxConcurrency int
	maxBatchSize   int
}

func NewService(renderer *render.Renderer, cfg config.BatchConfig) *Service {
	concurrency := cfg.MaxConcurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	maxBatchSize := cfg.MaxBatchSize
	if maxBatchSize <= 0 {
		maxBatchSize = 1
	}

	return &Service{
		renderer:       renderer,
		now:            time.Now,
		maxConcurrency: concurrency,
		maxBatchSize:   maxBatchSize,
	}
}

func (s *Service) BuildArchive(items []render.CardInput) ([]byte, error) {
	var buf bytes.Buffer
	if err := s.BuildArchiveToWriter(&buf, items); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *Service) BuildArchiveToWriter(writer io.Writer, items []render.CardInput) error {
	if len(items) == 0 {
		return fmt.Errorf("at least one batch item is required")
	}
	if len(items) > s.maxBatchSize {
		return fmt.Errorf("batch size %d exceeds max_batch_size %d", len(items), s.maxBatchSize)
	}

	archive := zip.NewWriter(writer)
	timestamp := s.now().UTC().Format("20060102-150405")
	rendered, err := s.renderConcurrently(items)
	if err != nil {
		return err
	}

	for index, item := range items {
		filename := buildFileName(item, timestamp, index+1)
		fileWriter, err := archive.Create(filename)
		if err != nil {
			return fmt.Errorf("create zip entry: %w", err)
		}
		if _, err := fileWriter.Write(rendered[index]); err != nil {
			return fmt.Errorf("write zip entry: %w", err)
		}
	}

	if err := archive.Close(); err != nil {
		return fmt.Errorf("close zip archive: %w", err)
	}

	return nil
}

func (s *Service) renderConcurrently(items []render.CardInput) ([][]byte, error) {
	results := make([][]byte, len(items))

	sem := make(chan struct{}, s.maxConcurrency)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	for index, item := range items {
		wg.Add(1)
		go func(itemIndex int, itemInput render.CardInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			pngData, err := s.renderer.RenderPNG(itemInput)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("render item %d: %w", itemIndex, err):
				default:
				}
				return
			}

			results[itemIndex] = pngData
		}(index, item)
	}

	wg.Wait()
	close(errCh)

	if err := <-errCh; err != nil {
		return nil, err
	}

	return results, nil
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
