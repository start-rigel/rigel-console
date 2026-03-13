package app

import (
	"embed"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/rigel-labs/rigel-console/internal/config"
	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

//go:embed web/index.html web/admin.html
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
	mux.HandleFunc("/api/admin/collect/search", a.handleAdminCollectSearch)
	mux.HandleFunc("/api/admin/collect/batch", a.handleAdminCollectBatch)
	mux.HandleFunc("/api/admin/products", a.handleAdminProducts)
	mux.HandleFunc("/api/admin/parts", a.handleAdminParts)
	mux.HandleFunc("/api/admin/jobs", a.handleAdminJobs)
	mux.HandleFunc("/api/admin/jobs/", a.handleAdminJobRoutes)
	mux.HandleFunc("/catalog/recommend", a.handleGenerateCatalogRecommendation)
	mux.HandleFunc("/build/generate", a.handleGenerateBuild)
	mux.HandleFunc("/build/", a.handleGetBuild)
	mux.HandleFunc("/parts/search", a.handleSearchParts)
	mux.HandleFunc("/admin", a.handleAdmin)
	mux.HandleFunc("/admin/", a.handleAdmin)
	mux.HandleFunc("/", a.handleIndex)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": a.cfg.ServiceName})
}

func (a *App) handleGenerateBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.GenerateBuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.console.GenerateBuild(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
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

func (a *App) handleGetBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	buildID := strings.TrimPrefix(r.URL.Path, "/build/")
	if buildID == "" || buildID == "/" {
		writeError(w, http.StatusBadRequest, "build id is required")
		return
	}
	response, err := a.console.GetBuild(r.Context(), buildID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleSearchParts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.console.SearchParts(r.Context(), strings.TrimSpace(r.URL.Query().Get("keyword")), limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items})
}

func (a *App) handleAdminProducts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.console.ListAdminProducts(r.Context(), model.AdminProductFilter{
		Keyword:          strings.TrimSpace(r.URL.Query().Get("keyword")),
		Limit:            limit,
		ShopType:         strings.TrimSpace(r.URL.Query().Get("shop_type")),
		RealOnly:         parseBoolQuery(r, "real_only"),
		SelfOperatedOnly: parseBoolQuery(r, "self_operated_only"),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items})
}

func (a *App) handleAdminCollectSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.AdminCollectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.console.StartAdminCollection(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleAdminCollectBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.AdminCollectBatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.console.StartAdminBatchCollection(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleAdminParts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.console.ListAdminParts(r.Context(), strings.TrimSpace(r.URL.Query().Get("keyword")), limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items})
}

func (a *App) handleAdminJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := a.console.ListAdminJobs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "items": items})
}

func (a *App) handleAdminJobRoutes(w http.ResponseWriter, r *http.Request) {
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/admin/jobs/")
	if trimmed == "" || trimmed == "/" {
		writeError(w, http.StatusBadRequest, "job id is required")
		return
	}
	if strings.HasSuffix(trimmed, "/retry") {
		a.handleAdminRetryJob(w, r, strings.TrimSuffix(trimmed, "/retry"))
		return
	}
	writeError(w, http.StatusNotFound, "admin job route not found")
}

func (a *App) handleAdminRetryJob(w http.ResponseWriter, r *http.Request, jobID string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if jobID == "" || jobID == "/" {
		writeError(w, http.StatusBadRequest, "job id is required")
		return
	}
	response, err := a.console.RetryAdminJob(r.Context(), jobID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data, err := webFS.ReadFile("web/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (a *App) handleAdmin(w http.ResponseWriter, _ *http.Request) {
	data, err := webFS.ReadFile("web/admin.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
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

func parseBoolQuery(r *http.Request, key string) bool {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	switch strings.ToLower(value) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
