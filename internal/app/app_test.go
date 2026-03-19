package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rigel-labs/rigel-console/internal/config"
	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

type buildClientStub struct{}
type jdCollectorClientStub struct{}
type keywordSeedRepoStub struct {
	mu    sync.Mutex
	items map[string]model.KeywordSeed
}

func (buildClientStub) GetPriceCatalog(context.Context, model.GenerateBuildRequest) (model.BuildEnginePriceCatalog, error) {
	return model.BuildEnginePriceCatalog{
		UseCase:   "gaming",
		BuildMode: "mixed",
		Items: []model.BuildEngineCatalogItem{
			{Category: "CPU", DisplayName: "Ryzen 5 7500F", NormalizedKey: "cpu:7500f", MedianPrice: 899, AvgPrice: 920, SampleCount: 5},
			{Category: "GPU", DisplayName: "RTX 4060", NormalizedKey: "gpu:rtx4060", MedianPrice: 2399, AvgPrice: 2410, SampleCount: 6},
		},
	}, nil
}

func (buildClientStub) GenerateCatalogAdvice(context.Context, model.GenerateBuildRequest, model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error) {
	return model.CatalogAdviceResponse{
		Selection: model.CatalogSelection{
			Budget:         6000,
			UseCase:        "gaming",
			BuildMode:      "mixed",
			EstimatedTotal: 3298,
			SelectedItems: []model.CatalogRecommendationItem{
				{Category: "CPU", DisplayName: "Ryzen 5 7500F", SelectedPrice: 899},
				{Category: "GPU", DisplayName: "RTX 4060", SelectedPrice: 2399},
			},
		},
		Advisory: model.Advice{Summary: "目录推荐说明"},
	}, nil
}

func (jdCollectorClientStub) GetScheduleConfig(context.Context) (model.CollectorScheduleResponse, error) {
	return model.CollectorScheduleResponse{
		Configured: true,
		Config: model.CollectorScheduleConfig{
			ID:                     "cfg-1",
			ServiceName:            "rigel-jd-collector",
			Enabled:                true,
			ScheduleTime:           "03:00",
			RequestIntervalSeconds: 3,
			QueryLimit:             5,
		},
	}, nil
}

func (jdCollectorClientStub) UpdateScheduleConfig(_ context.Context, payload model.CollectorScheduleUpsertRequest) (model.CollectorScheduleResponse, error) {
	return model.CollectorScheduleResponse{
		Configured: true,
		Config: model.CollectorScheduleConfig{
			ID:                     "cfg-1",
			ServiceName:            "rigel-jd-collector",
			Enabled:                payload.Enabled,
			ScheduleTime:           payload.ScheduleTime,
			RequestIntervalSeconds: payload.RequestIntervalSeconds,
			QueryLimit:             payload.QueryLimit,
		},
	}, nil
}

func newKeywordSeedRepoStub() *keywordSeedRepoStub {
	return &keywordSeedRepoStub{
		items: map[string]model.KeywordSeed{
			"seed-1": {ID: "seed-1", Category: "cpu", Keyword: "Ryzen 5 7500F", CanonicalModel: "Ryzen 5 7500F", Brand: "AMD", Enabled: true, Priority: 100},
		},
	}
}

func (r *keywordSeedRepoStub) ListKeywordSeeds(_ context.Context, filter model.KeywordSeedFilter) (model.KeywordSeedListResponse, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	items := make([]model.KeywordSeed, 0, len(r.items))
	for _, item := range r.items {
		if filter.Category != "" && !strings.EqualFold(filter.Category, item.Category) {
			continue
		}
		items = append(items, item)
	}
	return model.KeywordSeedListResponse{Items: items, Page: 1, PageSize: 20, Total: len(items)}, nil
}

func (r *keywordSeedRepoStub) GetKeywordSeed(_ context.Context, id string) (model.KeywordSeed, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	return item, ok, nil
}

func (r *keywordSeedRepoStub) CreateKeywordSeed(_ context.Context, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	return model.KeywordSeed{ID: "seed-new", Category: req.Category, Keyword: req.Keyword, CanonicalModel: req.CanonicalModel, Brand: req.Brand, Enabled: req.Enabled, Priority: req.Priority}, nil
}

func (r *keywordSeedRepoStub) UpdateKeywordSeed(_ context.Context, id string, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	return model.KeywordSeed{ID: id, Category: req.Category, Keyword: req.Keyword, CanonicalModel: req.CanonicalModel, Brand: req.Brand, Enabled: req.Enabled, Priority: req.Priority}, nil
}

func (r *keywordSeedRepoStub) SetKeywordSeedEnabled(_ context.Context, id string, enabled bool) (model.KeywordSeed, error) {
	return model.KeywordSeed{ID: id, Category: "cpu", Keyword: "Ryzen 5 7500F", CanonicalModel: "Ryzen 5 7500F", Enabled: enabled, Priority: 100}, nil
}

func newTestApp() *App {
	cfg := config.Config{
		ServiceName:         "rigel-console",
		AdminCookieName:     "rigel_admin_session",
		AdminCSRFCookieName: "rigel_admin_csrf",
		AnonymousCookieName: "rigel_anonymous_id",
		AdminAllowedCIDRs:   []string{"192.0.2.0/24"},
		ChallengeProvider:   "turnstile",
		ChallengeSiteKey:    "site-key",
	}
	console := consoleservice.New(buildClientStub{}, jdCollectorClientStub{}, "admin", "secret", 2, time.Minute, consoleservice.WithKeywordSeedRepository(newKeywordSeedRepoStub()))
	return New(cfg, console)
}

func adminCookies(t *testing.T, application *App) []*http.Cookie {
	t.Helper()
	sessionID, csrfToken, err := application.console.CreateAdminSession(context.Background())
	if err != nil {
		t.Fatalf("CreateAdminSession() error = %v", err)
	}
	return []*http.Cookie{
		{Name: "rigel_admin_session", Value: sessionID},
		{Name: "rigel_admin_csrf", Value: csrfToken},
	}
}

func TestIndex(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminLoginPage(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminKeywordsRequiresLogin(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/keywords", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302, got %d", rec.Code)
	}
}

func TestAdminKeywordAPIWithLogin(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/keyword-seeds", nil)
	for _, cookie := range adminCookies(t, application) {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminJDScheduleAPIWithLogin(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/jd/schedule", nil)
	for _, cookie := range adminCookies(t, application) {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAnonymousSession(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/session/anonymous", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var payload model.AnonymousSessionResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.AnonymousID == "" {
		t.Fatal("expected anonymous id")
	}
}

func TestGenerateCatalogRecommendation(t *testing.T) {
	application := newTestApp()
	body := []byte(`{"budget":6000,"use_case":"gaming","build_mode":"mixed"}`)
	req := httptest.NewRequest(http.MethodPost, "/catalog/recommend", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var payload model.CatalogRecommendationResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.RequestStatus.RemainingAIRequests != 1 {
		t.Fatalf("expected remaining requests 1, got %d", payload.RequestStatus.RemainingAIRequests)
	}
}
