package main

import (
	"log"
	"net/http"
	"os"
	"time"

	api "github.com/mohammadkhizerkhan/qr_generation/internal/http"
	"github.com/mohammadkhizerkhan/qr_generation/qrgen"
)

func main() {
	addr := os.Getenv("QRGEN_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	service := qrgen.NewService()
	server := &http.Server{
		Addr:              addr,
		Handler:           api.NewHandler(service),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("qr server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}