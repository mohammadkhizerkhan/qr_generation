package http

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/mohammadkhizerkhan/qr_generation/qrgen"
)

type Handler struct {
	service *qrgen.Service
}

func NewHandler(service *qrgen.Service) http.Handler {
	h := &Handler{service: service}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("POST /v1/render", h.render)
	mux.HandleFunc("POST /v1/render-metrics", h.renderMetrics)
	mux.HandleFunc("POST /v1/batch", h.batch)
	return withCORS(mux)
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	pngData, metrics, err := h.service.RenderPNGWithMetrics(toCardRequest(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if metrics != nil {
		log.Printf("http_render total_ms=%.2f mode=%s renderer_total_ms=%.2f qr_ms=%.2f template_ms=%.2f", float64(time.Since(requestStart).Microseconds())/1000, metrics.RenderMode, metrics.TotalRenderDurationMs, metrics.QRGenerationDurationMs, metrics.TemplateRenderDurationMs)
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=upi-qr.png")
	_, _ = w.Write(pngData)
}

// renderMetrics endpoint returns PNG + timing metrics in JSON for performance analysis
func (h *Handler) renderMetrics(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	pngData, metrics, err := h.service.RenderPNGWithMetrics(toCardRequest(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	resp := MetricsResponse{
		ImageBase64:              base64.StdEncoding.EncodeToString(pngData),
		QRGenerationDurationMs:   metrics.QRGenerationDurationMs,
		TemplateRenderDurationMs: metrics.TemplateRenderDurationMs,
		TotalRenderDurationMs:    metrics.TotalRenderDurationMs,
		QRGeneratorUsed:          metrics.QRGeneratorUsed,
		RenderMode:               metrics.RenderMode,
		GeneratorTimingsMs:       metrics.GeneratorTimingsMs,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
	requestStart := time.Now()
	var req BatchGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	items := make([]qrgen.CardRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, toCardRequest(item))
	}
	archiveData, err := h.service.RenderArchive(items, req.Concurrency)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	log.Printf("http_batch items=%d concurrency=%d total_ms=%.2f", len(req.Items), req.Concurrency, float64(time.Since(requestStart).Microseconds())/1000)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=upi-batch-"+time.Now().UTC().Format("20060102-150405")+".zip")
	_, _ = w.Write(archiveData)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func toCardRequest(req GenerateRequest) qrgen.CardRequest {
	return qrgen.CardRequest{
		UPIURI:          req.UPIURI,
		MerchantName:    req.MerchantName,
		MerchantUPIID:   req.MerchantUPIID,
		Description:     req.Description,
		RenderTemplate:  req.RenderTemplate,
		PayerName:       req.PayerName,
		LogoBase64:      req.LogoBase64,
		QRGenerator:     req.QRGenerator,
		BackgroundColor: req.BackgroundColor,
		AccentColor:     req.AccentColor,
		TextColor:       req.TextColor,
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "600")
		w.Header().Add("Vary", "Origin")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
