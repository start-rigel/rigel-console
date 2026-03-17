package console

import (
	"context"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type BuildEngineClient interface {
	GetPriceCatalog(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEnginePriceCatalog, error)
	GenerateCatalogAdvice(ctx context.Context, req model.GenerateBuildRequest, catalog model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error)
}

type Service struct {
	buildClient BuildEngineClient
}

func New(buildClient BuildEngineClient) *Service {
	return &Service{buildClient: buildClient}
}

func (s *Service) GenerateCatalogRecommendation(ctx context.Context, req model.GenerateBuildRequest) (model.CatalogRecommendationResponse, error) {
	catalog, err := s.buildClient.GetPriceCatalog(ctx, req)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	advice, err := s.buildClient.GenerateCatalogAdvice(ctx, req, catalog)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	return model.CatalogRecommendationResponse{
		CatalogItemCount: len(catalog.Items),
		CatalogWarnings:  catalog.Warnings,
		Selection:        advice.Selection,
		Advice:           &advice.Advisory,
	}, nil
}
