package console

import (
	"context"
	"fmt"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type BuildEngineClient interface {
	GenerateBuild(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEngineResponse, error)
	GetBuild(ctx context.Context, buildID string) (model.BuildEngineResponse, error)
	SearchParts(ctx context.Context, keyword string, limit int) ([]model.PartSearchResult, error)
}

type AIAdvisorClient interface {
	GenerateAdvice(ctx context.Context, build model.BuildEngineResponse) (model.AIAdvisorResponse, error)
}

type JDCollectorClient interface {
	ListProducts(ctx context.Context, filter model.AdminProductFilter) ([]model.AdminProduct, error)
	ListJobs(ctx context.Context, limit int) ([]model.AdminJob, error)
	TriggerCollection(ctx context.Context, req model.AdminCollectRequest) (model.AdminCollectResponse, error)
	TriggerBatchCollection(ctx context.Context, req model.AdminCollectBatchRequest) (model.AdminCollectBatchResponse, error)
	RetryJob(ctx context.Context, jobID string) (model.AdminCollectResponse, error)
}

type Service struct {
	buildClient BuildEngineClient
	aiClient    AIAdvisorClient
	jdClient    JDCollectorClient
}

func New(buildClient BuildEngineClient, aiClient AIAdvisorClient, jdClient JDCollectorClient) *Service {
	return &Service{buildClient: buildClient, aiClient: aiClient, jdClient: jdClient}
}

func (s *Service) GenerateBuild(ctx context.Context, req model.GenerateBuildRequest) (model.BuildResponse, error) {
	build, err := s.buildClient.GenerateBuild(ctx, req)
	if err != nil {
		return model.BuildResponse{}, err
	}
	return s.composeBuildResponse(ctx, build)
}

func (s *Service) GetBuild(ctx context.Context, buildID string) (model.BuildResponse, error) {
	build, err := s.buildClient.GetBuild(ctx, buildID)
	if err != nil {
		return model.BuildResponse{}, err
	}
	return s.composeBuildResponse(ctx, build)
}

func (s *Service) SearchParts(ctx context.Context, keyword string, limit int) ([]model.PartSearchResult, error) {
	return s.buildClient.SearchParts(ctx, keyword, limit)
}

func (s *Service) ListAdminProducts(ctx context.Context, filter model.AdminProductFilter) ([]model.AdminProduct, error) {
	return s.jdClient.ListProducts(ctx, filter)
}

func (s *Service) ListAdminParts(ctx context.Context, keyword string, limit int) ([]model.PartSearchResult, error) {
	return s.buildClient.SearchParts(ctx, keyword, limit)
}

func (s *Service) ListAdminJobs(ctx context.Context, limit int) ([]model.AdminJob, error) {
	return s.jdClient.ListJobs(ctx, limit)
}

func (s *Service) StartAdminCollection(ctx context.Context, req model.AdminCollectRequest) (model.AdminCollectResponse, error) {
	return s.jdClient.TriggerCollection(ctx, req)
}

func (s *Service) StartAdminBatchCollection(ctx context.Context, req model.AdminCollectBatchRequest) (model.AdminCollectBatchResponse, error) {
	return s.jdClient.TriggerBatchCollection(ctx, req)
}

func (s *Service) RetryAdminJob(ctx context.Context, jobID string) (model.AdminCollectResponse, error) {
	return s.jdClient.RetryJob(ctx, jobID)
}

func (s *Service) composeBuildResponse(ctx context.Context, build model.BuildEngineResponse) (model.BuildResponse, error) {
	if len(build.Results) == 0 {
		return model.BuildResponse{}, fmt.Errorf("build-engine returned no results")
	}
	primary := build.Results[0]
	for _, result := range build.Results {
		if result.Role == "primary" {
			primary = result
			break
		}
	}

	response := model.BuildResponse{
		BuildID:    build.BuildRequestID,
		TotalPrice: primary.TotalPrice,
		Currency:   primary.Currency,
		Warnings:   build.Warnings,
	}
	for _, item := range primary.Items {
		response.Items = append(response.Items, model.BuildItem{
			Category:    item.Category,
			DisplayName: item.DisplayName,
			UnitPrice:   item.UnitPrice,
			Source:      item.SourcePlatform,
			Reasons:     item.Reasons,
			Risks:       item.Risks,
		})
	}
	for _, result := range build.Results {
		if result.ResultID == primary.ResultID {
			continue
		}
		response.Alternatives = append(response.Alternatives, model.Alternative{BuildID: result.ResultID, Label: result.Role, TotalPrice: result.TotalPrice})
	}

	advice, err := s.aiClient.GenerateAdvice(ctx, build)
	if err == nil {
		response.Advice = &advice.Advisory
	}
	return response, nil
}
