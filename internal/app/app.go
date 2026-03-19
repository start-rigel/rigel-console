package app

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/rigel-labs/rigel-console/internal/config"
	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

//go:embed web/dist
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
	staticFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		panic(fmt.Sprintf("prepare embedded dist fs: %v", err))
	}
	mux.Handle("/assets/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/v1/session/anonymous", a.handleAnonymousSession)
	mux.HandleFunc("/catalog/recommend", a.handleGenerateCatalogRecommendation)
	mux.HandleFunc("/admin/login", a.handleAdminLogin)
	mux.HandleFunc("/admin/logout", a.handleAdminLogout)
	mux.HandleFunc("/admin/api/v1/jd/schedule", a.handleAdminJDSchedule)
	mux.HandleFunc("/admin/api/v1/keyword-seeds/template", a.handleAdminTemplate)
	mux.HandleFunc("/admin/api/v1/keyword-seeds/export", a.handleAdminExport)
	mux.HandleFunc("/admin/api/v1/keyword-seeds/import", a.handleAdminImport)
	mux.HandleFunc("/admin/api/v1/keyword-seeds/", a.handleAdminKeywordSeedItem)
	mux.HandleFunc("/admin/api/v1/keyword-seeds", a.handleAdminKeywordSeeds)
	mux.HandleFunc("/admin/keywords/import", a.handleAdminKeywordImportPage)
	mux.HandleFunc("/admin/keywords/new", a.handleAdminKeywordForm)
	mux.HandleFunc("/admin/keywords/", a.handleAdminKeywordRoutes)
	mux.HandleFunc("/admin/keywords", a.handleAdminKeywords)
	mux.HandleFunc("/admin/jd-schedule", a.handleAdminJDSchedulePage)
	mux.HandleFunc("/admin", a.handleAdminHome)
	mux.HandleFunc("/", a.handleIndex)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": a.cfg.ServiceName})
}

func (a *App) handleAnonymousSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	existing := a.readCookie(r, a.cfg.AnonymousCookieName)
	session := a.console.IssueAnonymousSession(existing)
	a.setCookie(w, a.cfg.AnonymousCookieName, session.AnonymousID, 30*24*time.Hour)
	writeJSON(w, http.StatusOK, session)
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
	anonymousID := strings.TrimSpace(r.Header.Get("X-Anonymous-Id"))
	if anonymousID == "" {
		anonymousID = a.readCookie(r, a.cfg.AnonymousCookieName)
	}
	session := a.console.IssueAnonymousSession(anonymousID)
	a.setCookie(w, a.cfg.AnonymousCookieName, session.AnonymousID, 30*24*time.Hour)

	response, err := a.console.GenerateCatalogRecommendation(r.Context(), req, session.AnonymousID)
	if err != nil {
		var rateLimited consoleservice.ErrRateLimited
		if errors.As(err, &rateLimited) {
			writeJSON(w, http.StatusTooManyRequests, map[string]any{
				"error": map[string]any{
					"code":             "rate_limited",
					"message":          "请求过于频繁，请稍后再试。",
					"cooldown_seconds": rateLimited.CooldownSeconds,
				},
			})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.serveSPA(w)
	case http.MethodPost:
		var req model.AdminLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		if !a.console.AuthenticateAdmin(strings.TrimSpace(req.Username), req.Password) {
			writeError(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		a.setCookie(w, a.cfg.AdminCookieName, "ok", 12*time.Hour)
		writeJSON(w, http.StatusOK, model.AdminLoginResponse{Username: strings.TrimSpace(req.Username)})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.AdminCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (a *App) handleAdminHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}
	a.serveSPA(w)
}

func (a *App) handleAdminKeywords(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/keywords" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}
	a.serveSPA(w)
}

func (a *App) handleAdminKeywordForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/keywords/new" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}
	a.serveSPA(w)
}

func (a *App) handleAdminKeywordImportPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/keywords/import" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}
	a.serveSPA(w)
}

func (a *App) handleAdminJDSchedulePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin/jd-schedule" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if !a.requireAdmin(w, r) {
		return
	}
	a.serveSPA(w)
}

func (a *App) handleAdminKeywordRoutes(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/edit") {
		a.serveSPA(w)
		return
	}
	writeError(w, http.StatusNotFound, "not found")
}

func (a *App) handleAdminJDSchedule(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		payload, err := a.console.GetCollectorScheduleConfig(r.Context())
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, payload)
	case http.MethodPut:
		var req model.CollectorScheduleUpsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		payload, err := a.console.UpdateCollectorScheduleConfig(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, payload)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAdminKeywordSeeds(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		filter := model.KeywordSeedFilter{
			Category: r.URL.Query().Get("category"),
			Brand:    r.URL.Query().Get("brand"),
			Keyword:  r.URL.Query().Get("keyword"),
			Page:     parseInt(r.URL.Query().Get("page"), 1),
			PageSize: parseInt(r.URL.Query().Get("page_size"), 20),
		}
		if rawEnabled := r.URL.Query().Get("enabled"); rawEnabled != "" {
			enabled := strings.EqualFold(rawEnabled, "true")
			filter.Enabled = &enabled
		}
		writeJSON(w, http.StatusOK, a.console.ListKeywordSeeds(filter))
	case http.MethodPost:
		var req model.KeywordSeedUpsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		item, err := a.console.CreateKeywordSeed(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAdminKeywordSeedItem(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/api/v1/keyword-seeds/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	id := parts[0]

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			item, ok := a.console.GetKeywordSeed(id)
			if !ok {
				writeError(w, http.StatusNotFound, "keyword seed not found")
				return
			}
			writeJSON(w, http.StatusOK, item)
		case http.MethodPut:
			var req model.KeywordSeedUpsertRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			item, err := a.console.UpdateKeywordSeed(id, req)
			if err != nil {
				status := http.StatusBadRequest
				var notFound consoleservice.ErrNotFound
				if errors.As(err, &notFound) {
					status = http.StatusNotFound
				}
				writeError(w, status, err.Error())
				return
			}
			writeJSON(w, http.StatusOK, item)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	switch parts[1] {
	case "enable":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		item, err := a.console.SetKeywordSeedEnabled(id, true)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	case "disable":
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		item, err := a.console.SetKeywordSeedEnabled(id, false)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, item)
	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

func (a *App) handleAdminTemplate(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	content, err := a.console.TemplateWorkbook()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeBinary(w, "keyword_seeds_template.xlsx", content)
}

func (a *App) handleAdminExport(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	content, err := a.console.ExportKeywordSeedsExcel()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeBinary(w, "keyword_seeds.xlsx", content)
}

func (a *App) handleAdminImport(w http.ResponseWriter, r *http.Request) {
	if !a.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	result, err := a.console.ImportKeywordSeeds(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (a *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	a.serveSPA(w)
}

func (a *App) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if a.readCookie(r, a.cfg.AdminCookieName) == "ok" {
		return true
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/") {
		writeError(w, http.StatusUnauthorized, "login required")
		return false
	}
	http.Redirect(w, r, "/admin/login", http.StatusFound)
	return false
}

func (a *App) readCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (a *App) setCookie(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *App) serveSPA(w http.ResponseWriter) {
	data, err := webFS.ReadFile("web/dist/index.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func writeBinary(w http.ResponseWriter, name string, content []byte) {
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
