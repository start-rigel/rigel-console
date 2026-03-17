package app

import (
	"embed"
	"encoding/json"
	"net/http"

	"github.com/rigel-labs/rigel-console/internal/config"
	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

//go:embed web/index.html
var webFS embed.FS

type App struct {
	cfg     config.Config
	console *consoleservice.Service
}

func New(cfg config.Config, console *consoleservice.Service) *App {
	return &App{cfg: cfg, console: console}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/catalog/recommend", a.handleGenerateCatalogRecommendation)
	mux.HandleFunc("/", a.handleIndex)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": a.cfg.ServiceName})
}

func (a *App) handleGenerateCatalogRecommendation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.GenerateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.console.GenerateCatalogRecommendation(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	data, err := webFS.ReadFile("web/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
