package http

import (
	"encoding/json"
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
	mux.HandleFunc("POST /v1/batch", h.batch)
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	pngData, err := h.service.RenderPNG(qrgen.CardRequest(req))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", "inline; filename=upi-qr.png")
	_, _ = w.Write(pngData)
}

func (h *Handler) batch(w http.ResponseWriter, r *http.Request) {
	var req BatchGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	items := make([]qrgen.CardRequest, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, qrgen.CardRequest(item))
	}

	archiveData, err := h.service.RenderArchive(items)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

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
