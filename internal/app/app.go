package app

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
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
	cfg               config.Config
	console           *consoleservice.Service
	adminAllowedCIDRs []*net.IPNet
	trustedProxyCIDRs []*net.IPNet
}

func New(cfg config.Config, console *consoleservice.Service) *App {
	return &App{
		cfg:               cfg,
		console:           console,
		adminAllowedCIDRs: parseCIDRs(cfg.AdminAllowedCIDRs),
		trustedProxyCIDRs: parseCIDRs(cfg.TrustedProxyCIDRs),
	}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	staticFS, err := fs.Sub(webFS, "web/dist")
	if err != nil {
		panic(fmt.Sprintf("prepare embedded dist fs: %v", err))
	}
	mux.Handle("/assets/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/healthz", a.handleHealth)
	mux.HandleFunc("/api/v1/bootstrap", a.handleBootstrap)
	mux.HandleFunc("/api/v1/session/anonymous", a.handleAnonymousSession)
	mux.HandleFunc("/api/v1/challenge/verify", a.handleChallengeVerify)
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

func (a *App) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	writeJSON(w, http.StatusOK, model.PublicBootstrapResponse{
		ChallengeProvider: a.cfg.ChallengeProvider,
		ChallengeSiteKey:  a.cfg.ChallengeSiteKey,
	})
}

func (a *App) handleAnonymousSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	existing := a.readCookie(r, a.cfg.AnonymousCookieName)
	session, err := a.console.IssueAnonymousSession(r.Context(), consoleservice.RequestMeta{
		ClientIP:          a.clientIP(r),
		DeviceFingerprint: strings.TrimSpace(r.Header.Get("X-Device-Fingerprint")),
		AnonymousID:       existing,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	a.setCookie(w, r, a.cfg.AnonymousCookieName, session.AnonymousID, time.Duration(session.SessionExpiresInSeconds)*time.Second)
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
	meta := consoleservice.RequestMeta{
		ClientIP:          a.clientIP(r),
		DeviceFingerprint: strings.TrimSpace(r.Header.Get("X-Device-Fingerprint")),
		AnonymousID:       anonymousID,
	}
	session, err := a.console.IssueAnonymousSession(r.Context(), meta)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	meta.AnonymousID = session.AnonymousID
	a.setCookie(w, r, a.cfg.AnonymousCookieName, session.AnonymousID, time.Duration(session.SessionExpiresInSeconds)*time.Second)

	response, err := a.console.GenerateCatalogRecommendation(r.Context(), req, meta)
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
		var challengeRequired consoleservice.ErrChallengeRequired
		if errors.As(err, &challengeRequired) {
			writeJSON(w, http.StatusForbidden, map[string]any{
				"error": map[string]any{
					"code":               "challenge_required",
					"message":            "当前请求需要先完成安全验证。",
					"challenge_required": true,
					"risk_level":         challengeRequired.RiskLevel,
				},
			})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleChallengeVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req model.ChallengeVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	response, err := a.console.VerifyChallenge(r.Context(), consoleservice.RequestMeta{
		ClientIP:          a.clientIP(r),
		DeviceFingerprint: strings.TrimSpace(req.DeviceFingerprint),
		AnonymousID:       strings.TrimSpace(req.AnonymousID),
	}, req.ChallengeToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *App) handleAdminLogin(w http.ResponseWriter, r *http.Request) {
	if !a.adminAllowed(r) {
		writeError(w, http.StatusForbidden, "admin access is restricted to private network")
		return
	}
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
		sessionID, csrfToken, err := a.console.CreateAdminSession(r.Context())
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		a.setCookie(w, r, a.cfg.AdminCookieName, sessionID, 12*time.Hour)
		a.setCSRFCookie(w, r, csrfToken, 12*time.Hour)
		writeJSON(w, http.StatusOK, model.AdminLoginResponse{Username: strings.TrimSpace(req.Username)})
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (a *App) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	if !a.adminAllowed(r) {
		writeError(w, http.StatusForbidden, "admin access is restricted to private network")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if !a.requireAdminCSRF(w, r) {
		return
	}
	_ = a.console.DeleteAdminSession(r.Context(), a.readCookie(r, a.cfg.AdminCookieName))
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.AdminCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.AdminCSRFCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   r.TLS != nil,
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
		if !a.requireAdminCSRF(w, r) {
			return
		}
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
		payload, err := a.console.ListKeywordSeeds(r.Context(), filter)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, payload)
	case http.MethodPost:
		if !a.requireAdminCSRF(w, r) {
			return
		}
		var req model.KeywordSeedUpsertRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		item, err := a.console.CreateKeywordSeed(r.Context(), req)
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
			item, ok, err := a.console.GetKeywordSeed(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			if !ok {
				writeError(w, http.StatusNotFound, "keyword seed not found")
				return
			}
			writeJSON(w, http.StatusOK, item)
		case http.MethodPut:
			if !a.requireAdminCSRF(w, r) {
				return
			}
			var req model.KeywordSeedUpsertRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
			item, err := a.console.UpdateKeywordSeed(r.Context(), id, req)
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
		if !a.requireAdminCSRF(w, r) {
			return
		}
		item, err := a.console.SetKeywordSeedEnabled(r.Context(), id, true)
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
		if !a.requireAdminCSRF(w, r) {
			return
		}
		item, err := a.console.SetKeywordSeedEnabled(r.Context(), id, false)
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
	content, err := a.console.ExportKeywordSeedsExcel(r.Context())
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
	if !a.requireAdminCSRF(w, r) {
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	result, err := a.console.ImportKeywordSeeds(r.Context(), file)
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
	if !a.adminAllowed(r) {
		writeError(w, http.StatusForbidden, "admin access is restricted to private network")
		return false
	}
	_, ok, err := a.console.ValidateAdminSession(r.Context(), a.readCookie(r, a.cfg.AdminCookieName))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return false
	}
	if ok {
		return true
	}
	if strings.HasPrefix(r.URL.Path, "/admin/api/") {
		writeError(w, http.StatusUnauthorized, "login required")
		return false
	}
	http.Redirect(w, r, "/admin/login", http.StatusFound)
	return false
}

func (a *App) requireAdminCSRF(w http.ResponseWriter, r *http.Request) bool {
	session, ok, err := a.console.ValidateAdminSession(r.Context(), a.readCookie(r, a.cfg.AdminCookieName))
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return false
	}
	if !ok {
		writeError(w, http.StatusUnauthorized, "login required")
		return false
	}
	headerToken := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
	cookieToken := a.readCookie(r, a.cfg.AdminCSRFCookieName)
	if headerToken == "" || cookieToken == "" || headerToken != cookieToken || headerToken != session.CSRFToken {
		writeError(w, http.StatusForbidden, "invalid csrf token")
		return false
	}
	return true
}

func (a *App) readCookie(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (a *App) setCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *App) setCSRFCookie(w http.ResponseWriter, r *http.Request, value string, maxAge time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.cfg.AdminCSRFCookieName,
		Value:    value,
		Path:     "/",
		MaxAge:   int(maxAge.Seconds()),
		Secure:   r.TLS != nil,
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

func parseCIDRs(values []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(values))
	for _, value := range values {
		_, network, err := net.ParseCIDR(strings.TrimSpace(value))
		if err == nil {
			out = append(out, network)
		}
	}
	return out
}

func (a *App) adminAllowed(r *http.Request) bool {
	ip := net.ParseIP(a.clientIP(r))
	if ip == nil {
		return false
	}
	for _, network := range a.adminAllowedCIDRs {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func (a *App) clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	remoteIP := net.ParseIP(host)
	if remoteIP == nil {
		return host
	}

	for _, network := range a.trustedProxyCIDRs {
		if network.Contains(remoteIP) {
			forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
			if net.ParseIP(forwarded) != nil {
				return forwarded
			}
			break
		}
	}
	return remoteIP.String()
}
