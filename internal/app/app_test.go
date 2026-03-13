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
type aiClientStub struct{}
type jdClientStub struct{}
type goofishClientStub struct{}

func (buildClientStub) GenerateBuild(context.Context, model.GenerateBuildRequest) (model.BuildEngineResponse, error) {
	return sampleBuild(), nil
}
func (buildClientStub) GetBuild(context.Context, string) (model.BuildEngineResponse, error) {
	return sampleBuild(), nil
}
func (buildClientStub) SearchParts(context.Context, string, int) ([]model.PartSearchResult, error) {
	return []model.PartSearchResult{{ID: "part-1", Category: "CPU", Brand: "AMD", Model: "Ryzen 5 7500F", DisplayName: "CPU AMD Ryzen 5 7500F"}}, nil
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
func (aiClientStub) GenerateAdvice(context.Context, model.BuildEngineResponse) (model.AIAdvisorResponse, error) {
	return model.AIAdvisorResponse{Advisory: model.Advice{Summary: "说明文本"}}, nil
}
func (aiClientStub) GenerateCatalogAdvice(context.Context, model.GenerateBuildRequest, model.BuildEnginePriceCatalog) (model.AIAdvisorCatalogResponse, error) {
	return model.AIAdvisorCatalogResponse{
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
func (jdClientStub) ListProducts(context.Context, model.AdminProductFilter) ([]model.AdminProduct, error) {
	return []model.AdminProduct{{ID: "product-1", Title: "RTX 4060 官方自营", Price: 1999, Currency: "CNY"}}, nil
}
func (jdClientStub) ListJobs(context.Context, int) ([]model.AdminJob, error) {
	return []model.AdminJob{{ID: "job-1", JobType: "jd_collect", Status: "succeeded"}}, nil
}
func (jdClientStub) TriggerCollection(context.Context, model.AdminCollectRequest) (model.AdminCollectResponse, error) {
	return model.AdminCollectResponse{JobID: "job-2", Persisted: true, PersistedCount: 2}, nil
}
func (jdClientStub) TriggerBatchCollection(context.Context, model.AdminCollectBatchRequest) (model.AdminCollectBatchResponse, error) {
	return model.AdminCollectBatchResponse{Preset: "mvp_base", TotalJobs: 8, TotalPersisted: 8}, nil
}
func (jdClientStub) RetryJob(context.Context, string) (model.AdminCollectResponse, error) {
	return model.AdminCollectResponse{JobID: "job-3", RetriedFromJobID: "job-1", Persisted: true, PersistedCount: 2}, nil
}
func (goofishClientStub) ListStateFiles(context.Context) ([]model.GoofishStateFile, error) {
	return []model.GoofishStateFile{{Name: "goofish_state.json", Path: "/tmp/goofish_state.json", IsRoot: true}}, nil
}
func (goofishClientStub) PromoteStateFile(context.Context, string) (model.GoofishStateFile, error) {
	return model.GoofishStateFile{Name: "goofish_state.json", Path: "/tmp/goofish_state.json", IsRoot: true}, nil
}
func (goofishClientStub) ValidateState(context.Context, model.GoofishValidateRequest) (model.GoofishValidateResponse, error) {
	return model.GoofishValidateResponse{Valid: true, StateFile: "goofish_state.json", Keyword: "电脑 内存", SampleCount: 1}, nil
}

func TestIndex(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGenerateBuild(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	body := []byte(`{"budget":6000,"use_case":"gaming","build_mode":"new_only"}`)
	req := httptest.NewRequest(http.MethodPost, "/build/generate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGenerateCatalogRecommendation(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	body := []byte(`{"budget":6000,"use_case":"gaming","build_mode":"mixed"}`)
	req := httptest.NewRequest(http.MethodPost, "/catalog/recommend", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminProducts(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/products?keyword=4060&limit=10", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCatalogPrices(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/catalog/prices?use_case=gaming&build_mode=mixed", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCollectSearch(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	body := []byte(`{"keyword":"RTX 4060","category":"GPU","limit":2,"persist":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/collect/search", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminCollectBatch(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	body := []byte(`{"preset":"mvp_base"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/collect/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminRetryJob(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	req := httptest.NewRequest(http.MethodPost, "/api/admin/jobs/job-1/retry", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminGoofishStateFiles(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	req := httptest.NewRequest(http.MethodGet, "/api/admin/goofish/state-files", nil)
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminGoofishValidate(t *testing.T) {
	application := New(config.Config{ServiceName: "rigel-console"}, consoleservice.New(buildClientStub{}, aiClientStub{}, jdClientStub{}, goofishClientStub{}))
	body := []byte(`{"account_state_file":"goofish_state.json"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/goofish/state/validate", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	application.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func sampleBuild() model.BuildEngineResponse {
	return model.BuildEngineResponse{BuildRequestID: "build-1", Results: []model.BuildEngineResult{{ResultID: "result-1", Role: "primary", TotalPrice: 5899, Currency: "CNY", Items: []model.BuildEngineResultItem{{Category: "CPU", DisplayName: "Ryzen 5 7500F", UnitPrice: 1199, SourcePlatform: "jd"}}}}}
}
