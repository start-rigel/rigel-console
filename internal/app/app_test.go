package app

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

func TestIndex(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestKeywordsPage(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/keywords", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestKeywordEditPage(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/keywords/seed-1/edit", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestKeywordNewPage(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/keywords/new", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestKeywordImportPage(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/keywords/import", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGenerateCatalogRecommendation(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}))
	body := []byte(`{"budget":6000,"use_case":"gaming","build_mode":"mixed"}`)
	req := httptest.NewRequest(http.MethodPost, "/catalog/recommend", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}
