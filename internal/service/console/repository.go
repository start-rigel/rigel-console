package console

import (
	"context"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
)

type KeywordSeedRepository interface {
	ListKeywordSeeds(ctx context.Context, filter model.KeywordSeedFilter) (model.KeywordSeedListResponse, error)
	GetKeywordSeed(ctx context.Context, id string) (model.KeywordSeed, bool, error)
	CreateKeywordSeed(ctx context.Context, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error)
	UpdateKeywordSeed(ctx context.Context, id string, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error)
	SetKeywordSeedEnabled(ctx context.Context, id string, enabled bool) (model.KeywordSeed, error)
}
