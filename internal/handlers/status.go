package handlers

import (
	"go-pet-shop/internal/handlers/httpx"
	"log/slog"
	"net/http"
)

type HealthResponse struct {
	Status string `json:"status"`
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Received health check request", slog.String("method", r.Method), slog.String("url", r.URL.String()))
	httpx.JSON(w, http.StatusOK, HealthResponse{Status: "OK"})
}
