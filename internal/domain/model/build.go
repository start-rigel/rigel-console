package model

import "time"

// GenerateBuildRequest is the console-facing request contract for recommendation creation.
type GenerateBuildRequest struct {
	Budget              float64         `json:"budget"`
	UseCase             string          `json:"use_case"`
	BuildMode           string          `json:"build_mode"`
	BrandPreference     BrandPreference `json:"brand_preference"`
	SpecialRequirements []string        `json:"special_requirements,omitempty"`
	Notes               string          `json:"notes,omitempty"`
}

type BrandPreference struct {
	CPU string `json:"cpu,omitempty"`
	GPU string `json:"gpu,omitempty"`
}

// CatalogRecommendationResponse is the user-facing recommendation assembled from the aggregated price catalog.
type CatalogRecommendationResponse struct {
	RequestStatus    RequestStatus    `json:"request_status"`
	CatalogItemCount int              `json:"catalog_item_count"`
	CatalogWarnings  []string         `json:"catalog_warnings,omitempty"`
	Selection        CatalogSelection `json:"selection"`
	Advice           *Advice          `json:"advice,omitempty"`
}

type RequestStatus struct {
	CacheHit            bool `json:"cache_hit"`
	RemainingAIRequests int  `json:"remaining_ai_requests"`
	CooldownSeconds     int  `json:"cooldown_seconds"`
}

// CatalogSelection is the selected shopping list draft returned from build-engine.
type CatalogSelection struct {
	Budget         float64                     `json:"budget"`
	UseCase        string                      `json:"use_case"`
	BuildMode      string                      `json:"build_mode"`
	EstimatedTotal float64                     `json:"estimated_total"`
	Warnings       []string                    `json:"warnings,omitempty"`
	SelectedItems  []CatalogRecommendationItem `json:"selected_items"`
}

// CatalogRecommendationItem is one chosen part model from the price catalog.
type CatalogRecommendationItem struct {
	Category        string   `json:"category"`
	DisplayName     string   `json:"display_name"`
	NormalizedKey   string   `json:"normalized_key"`
	SampleCount     int      `json:"sample_count"`
	SelectedPrice   float64  `json:"selected_price"`
	MedianPrice     float64  `json:"median_price"`
	SourcePlatforms []string `json:"source_platforms,omitempty"`
	Reasons         []string `json:"reasons,omitempty"`
}

// Advice is the recommendation explanation block shown to users.
type Advice struct {
	Summary         string   `json:"summary"`
	Reasons         []string `json:"reasons"`
	FitFor          []string `json:"fit_for"`
	Risks           []string `json:"risks"`
	UpgradeAdvice   []string `json:"upgrade_advice"`
	AlternativeNote string   `json:"alternative_note"`
}

// BuildEnginePriceCatalog mirrors build-engine's aggregated price catalog response.
type BuildEnginePriceCatalog struct {
	UseCase   string                   `json:"use_case"`
	BuildMode string                   `json:"build_mode"`
	Warnings  []string                 `json:"warnings,omitempty"`
	Items     []BuildEngineCatalogItem `json:"items"`
}

type BuildEngineCatalogItem struct {
	Category        string                         `json:"category"`
	Brand           string                         `json:"brand"`
	Model           string                         `json:"model"`
	DisplayName     string                         `json:"display_name"`
	NormalizedKey   string                         `json:"normalized_key"`
	SampleCount     int                            `json:"sample_count"`
	AvgPrice        float64                        `json:"avg_price"`
	MedianPrice     float64                        `json:"median_price"`
	MinPrice        float64                        `json:"min_price"`
	MaxPrice        float64                        `json:"max_price"`
	Platforms       []string                       `json:"platforms,omitempty"`
	SourceBreakdown []BuildEngineCatalogSourceItem `json:"source_breakdown,omitempty"`
}

type BuildEngineCatalogSourceItem struct {
	SourcePlatform string  `json:"source_platform"`
	SampleCount    int     `json:"sample_count"`
	AvgPrice       float64 `json:"avg_price"`
	MinPrice       float64 `json:"min_price"`
	MaxPrice       float64 `json:"max_price"`
}

// CatalogAdviceResponse mirrors build-engine's catalog recommendation response.
type CatalogAdviceResponse struct {
	Provider     string           `json:"provider"`
	FallbackUsed bool             `json:"fallback_used"`
	Selection    CatalogSelection `json:"selection"`
	Advisory     Advice           `json:"advisory"`
}

type AnonymousSessionResponse struct {
	AnonymousID         string `json:"anonymous_id"`
	CooldownSeconds     int    `json:"cooldown_seconds"`
	RemainingAIRequests int    `json:"remaining_ai_requests"`
	ChallengeRequired   bool   `json:"challenge_required"`
}

type KeywordSeed struct {
	ID             string    `json:"id"`
	Category       string    `json:"category"`
	Keyword        string    `json:"keyword"`
	CanonicalModel string    `json:"canonical_model"`
	Brand          string    `json:"brand,omitempty"`
	Aliases        []string  `json:"aliases,omitempty"`
	Priority       int       `json:"priority"`
	Enabled        bool      `json:"enabled"`
	Notes          string    `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

type KeywordSeedUpsertRequest struct {
	Category       string   `json:"category"`
	Keyword        string   `json:"keyword"`
	CanonicalModel string   `json:"canonical_model"`
	Brand          string   `json:"brand,omitempty"`
	Aliases        []string `json:"aliases,omitempty"`
	Priority       int      `json:"priority"`
	Enabled        bool     `json:"enabled"`
	Notes          string   `json:"notes,omitempty"`
}

type KeywordSeedListResponse struct {
	Items    []KeywordSeed `json:"items"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int           `json:"total"`
}

type KeywordSeedFilter struct {
	Category string
	Brand    string
	Keyword  string
	Enabled  *bool
	Page     int
	PageSize int
}

type AdminLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminLoginResponse struct {
	Username string `json:"username"`
}

type KeywordSeedImportError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

type KeywordSeedImportResponse struct {
	JobID         string                   `json:"job_id"`
	ImportedCount int                      `json:"imported_count"`
	FailedCount   int                      `json:"failed_count"`
	Errors        []KeywordSeedImportError `json:"errors,omitempty"`
}
