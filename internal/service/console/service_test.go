package console

import (
	"context"
	"testing"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type buildClientStub struct{}

type aiClientStub struct{}

type jdClientStub struct{}

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
	return model.AIAdvisorResponse{Advisory: model.Advice{Summary: "说明文本", Reasons: []string{"原因"}}}, nil
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
		Advisory: model.Advice{Summary: "目录推荐说明", Reasons: []string{"价格目录已聚合"}},
	}, nil
}
func (jdClientStub) ListProducts(context.Context, model.AdminProductFilter) ([]model.AdminProduct, error) {
	return []model.AdminProduct{{ID: "product-1", Title: "RTX 4060 官方自营", Price: 1999, Currency: "CNY", Availability: "in_stock"}}, nil
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

func TestGenerateBuild(t *testing.T) {
	service := New(buildClientStub{}, aiClientStub{}, jdClientStub{})
	response, err := service.GenerateBuild(context.Background(), model.GenerateBuildRequest{Budget: 6000, UseCase: "gaming", BuildMode: "new_only"})
	if err != nil {
		t.Fatalf("GenerateBuild() error = %v", err)
	}
	if response.BuildID == "" {
		t.Fatal("expected build id")
	}
	if response.Advice == nil {
		t.Fatal("expected advice")
	}
}

func TestGenerateCatalogRecommendation(t *testing.T) {
	service := New(buildClientStub{}, aiClientStub{}, jdClientStub{})
	response, err := service.GenerateCatalogRecommendation(context.Background(), model.GenerateBuildRequest{Budget: 6000, UseCase: "gaming", BuildMode: "mixed"})
	if err != nil {
		t.Fatalf("GenerateCatalogRecommendation() error = %v", err)
	}
	if response.CatalogItemCount != 2 {
		t.Fatalf("expected 2 catalog items, got %d", response.CatalogItemCount)
	}
	if response.Advice == nil || response.Advice.Summary == "" {
		t.Fatal("expected catalog advice")
	}
}

func TestListAdminData(t *testing.T) {
	service := New(buildClientStub{}, aiClientStub{}, jdClientStub{})

	products, err := service.ListAdminProducts(context.Background(), model.AdminProductFilter{Keyword: "4060", Limit: 10, RealOnly: true, SelfOperatedOnly: true})
	if err != nil {
		t.Fatalf("ListAdminProducts() error = %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	jobs, err := service.ListAdminJobs(context.Background(), 10)
	if err != nil {
		t.Fatalf("ListAdminJobs() error = %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
}

func TestAdminActions(t *testing.T) {
	service := New(buildClientStub{}, aiClientStub{}, jdClientStub{})

	collectResponse, err := service.StartAdminCollection(context.Background(), model.AdminCollectRequest{
		Keyword:  "RTX 4060",
		Category: "GPU",
		Limit:    2,
		Persist:  true,
	})
	if err != nil {
		t.Fatalf("StartAdminCollection() error = %v", err)
	}
	if collectResponse.JobID == "" {
		t.Fatal("expected collection job id")
	}

	retryResponse, err := service.RetryAdminJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("RetryAdminJob() error = %v", err)
	}
	if retryResponse.RetriedFromJobID != "job-1" {
		t.Fatalf("expected retried_from_job_id job-1, got %q", retryResponse.RetriedFromJobID)
	}

	batchResponse, err := service.StartAdminBatchCollection(context.Background(), model.AdminCollectBatchRequest{Preset: "mvp_base"})
	if err != nil {
		t.Fatalf("StartAdminBatchCollection() error = %v", err)
	}
	if batchResponse.TotalJobs != 8 {
		t.Fatalf("expected 8 batch jobs, got %d", batchResponse.TotalJobs)
	}
}

func sampleBuild() model.BuildEngineResponse {
	return model.BuildEngineResponse{
		BuildRequestID: "build-1",
		Budget:         6000,
		UseCase:        "gaming",
		BuildMode:      "new_only",
		Results: []model.BuildEngineResult{{
			ResultID:   "result-1",
			Role:       "primary",
			TotalPrice: 5899,
			Currency:   "CNY",
			Items:      []model.BuildEngineResultItem{{Category: "CPU", DisplayName: "Ryzen 5 7500F", UnitPrice: 1199, SourcePlatform: "jd"}},
		}},
	}
}
