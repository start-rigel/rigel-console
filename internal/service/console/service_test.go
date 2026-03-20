package console

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
	"github.com/xuri/excelize/v2"
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
		Advisory: model.Advice{Summary: "目录推荐说明", Reasons: []string{"价格目录已聚合"}},
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
	resp.AIRuntime.Model = "openai/gpt-5.4-nano"
	resp.CatalogAILimits.MaxModelsPerCategory = 5
	return resp, nil
}

func (buildClientStub) UpdateSystemSettings(context.Context, model.UpdateSystemSettingsRequest) (model.SystemSettingsResponse, error) {
	var resp model.SystemSettingsResponse
	resp.AIRuntime.Model = "openai/gpt-5.4-nano"
	resp.CatalogAILimits.MaxModelsPerCategory = 6
	return resp, nil
}

func TestGenerateCatalogRecommendationCachesResult(t *testing.T) {
	service := New(buildClientStub{}, "admin", "secret", 2, time.Minute)
	req := model.GenerateBuildRequest{Budget: 6000, UseCase: "gaming", BuildMode: "mixed"}

	first, err := service.GenerateCatalogRecommendation(context.Background(), req, RequestMeta{AnonymousID: "anon-1"})
	if err != nil {
		t.Fatalf("GenerateCatalogRecommendation() first error = %v", err)
	}
	second, err := service.GenerateCatalogRecommendation(context.Background(), req, RequestMeta{AnonymousID: "anon-1"})
	if err != nil {
		t.Fatalf("GenerateCatalogRecommendation() second error = %v", err)
	}
	if first.RequestStatus.CacheHit {
		t.Fatal("expected first request not to hit cache")
	}
	if !second.RequestStatus.CacheHit {
		t.Fatal("expected second request to hit cache")
	}
}

func TestKeywordSeedCRUD(t *testing.T) {
	service := New(buildClientStub{}, "admin", "secret", 2, time.Minute)
	item, err := service.CreateKeywordSeed(context.Background(), model.KeywordSeedUpsertRequest{
		Category:       "cpu",
		Keyword:        "Ryzen 7 7700",
		CanonicalModel: "Ryzen 7 7700",
		Brand:          "AMD",
		Aliases:        []string{"7700"},
		Priority:       110,
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("CreateKeywordSeed() error = %v", err)
	}
	updated, err := service.UpdateKeywordSeed(context.Background(), item.ID, model.KeywordSeedUpsertRequest{
		Category:       "cpu",
		Keyword:        "Ryzen 7 7700",
		CanonicalModel: "Ryzen 7 7700",
		Brand:          "AMD",
		Aliases:        []string{"7700", "AMD 7700"},
		Priority:       120,
		Enabled:        false,
		Notes:          "updated",
	})
	if err != nil {
		t.Fatalf("UpdateKeywordSeed() error = %v", err)
	}
	if updated.Priority != 120 || updated.Enabled {
		t.Fatalf("unexpected updated seed: %+v", updated)
	}
}

func TestImportKeywordSeeds(t *testing.T) {
	service := New(buildClientStub{}, "admin", "secret", 2, time.Minute)
	file := excelize.NewFile()
	sheet := file.GetSheetName(0)
	rows := [][]any{
		{"category", "keyword", "canonical_model", "brand", "aliases", "priority", "enabled", "notes"},
		{"cpu", "Ryzen 5 9600X", "Ryzen 5 9600X", "AMD", "9600X", 100, true, "new"},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			_ = file.SetCellValue(sheet, cell, value)
		}
	}
	buf, err := file.WriteToBuffer()
	if err != nil {
		t.Fatalf("WriteToBuffer() error = %v", err)
	}

	result, err := service.ImportKeywordSeeds(context.Background(), ioNopCloser{Reader: bytes.NewReader(buf.Bytes())})
	if err != nil {
		t.Fatalf("ImportKeywordSeeds() error = %v", err)
	}
	if result.ImportedCount != 1 {
		t.Fatalf("expected imported count 1, got %d", result.ImportedCount)
	}
}

type ioNopCloser struct {
	*bytes.Reader
}

func (ioNopCloser) Close() error {
	return nil
}
