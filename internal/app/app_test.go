package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rigel-labs/rigel-console/internal/config"
	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

type buildClientStub struct{}

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

func (buildClientStub) RecommendBuild(context.Context, model.GenerateBuildRequest) (model.CatalogRecommendationResponse, error) {
	return model.CatalogRecommendationResponse{
		Provider:       "build-engine",
		FallbackUsed:   true,
		Summary:        "目录推荐说明",
		EstimatedTotal: 3298,
		WithinBudget:   true,
		BuildItems: []model.BuildRecommendationItem{
			{Category: "CPU", TargetModel: "Ryzen 5 7500F", RecommendedProduct: &model.BuildProductRef{DisplayName: "Ryzen 5 7500F", Price: 899}},
			{Category: "GPU", TargetModel: "RTX 4060", RecommendedProduct: &model.BuildProductRef{DisplayName: "RTX 4060", Price: 2399}},
		},
		Advice: model.BuildAdvice{Reasons: []string{"价格目录已聚合"}},
	}, nil
}

func (buildClientStub) GetSystemSettings(context.Context) (model.SystemSettingsResponse, error) {
	var resp model.SystemSettingsResponse
	resp.CatalogAILimits.MaxModelsPerCategory = 5
	return resp, nil
}

func (buildClientStub) UpdateSystemSettings(context.Context, model.UpdateSystemSettingsRequest) (model.SystemSettingsResponse, error) {
	var resp model.SystemSettingsResponse
	resp.CatalogAILimits.MaxModelsPerCategory = 6
	return resp, nil
}

func newTestApp() *App {
	cfg := config.Config{
		ServiceName:         "rigel-console",
		AdminCookieName:     "rigel_admin_session",
		AnonymousCookieName: "rigel_anonymous_id",
	}
	console := consoleservice.New(buildClientStub{}, "admin", "secret", 2, time.Minute)
	return New(cfg, console)
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
	req.AddCookie(&http.Cookie{Name: "rigel_admin_session", Value: "ok"})
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

func TestAdminSystemSettingsWithLogin(t *testing.T) {
	application := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/admin/api/v1/settings/system", nil)
	req.AddCookie(&http.Cookie{Name: "rigel_admin_session", Value: "ok"})
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
