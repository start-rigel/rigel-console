package console

import (
	"context"
	"testing"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
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

func TestGenerateCatalogRecommendation(t *testing.T) {
	service := New(buildClientStub{})
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
