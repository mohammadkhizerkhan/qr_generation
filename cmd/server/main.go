package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/mohammadkhizerkhan/qr_generation/internal/config"
	api "github.com/mohammadkhizerkhan/qr_generation/internal/http"
	"github.com/mohammadkhizerkhan/qr_generation/qrgen"
)

func main() {
	addr := os.Getenv("QRGEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	configPath := os.Getenv("QRGEN_CONFIG")
	if configPath == "" {
		configPath = "config.json"
	}

	cfg, err := config.LoadOrDefault(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	service := qrgen.NewServiceWithConfig(cfg)
	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewHandler(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("qr server listening on %s with template %s", addr, cfg.DefaultTemplate)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
