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

type CatalogRecommendationResponse struct {
	RequestStatus    RequestStatus             `json:"request_status"`
	Provider         string                    `json:"provider"`
	FallbackUsed     bool                      `json:"fallback_used"`
	Request          BuildRequestEcho          `json:"request"`
	Summary          string                    `json:"summary"`
	EstimatedTotal   float64                   `json:"estimated_total"`
	WithinBudget     bool                      `json:"within_budget"`
	Warnings         []string                  `json:"warnings,omitempty"`
	BuildItems       []BuildRecommendationItem `json:"build_items"`
	Advice           BuildAdvice               `json:"advice"`
	CatalogItemCount int                       `json:"catalog_item_count,omitempty"`
	CatalogWarnings  []string                  `json:"catalog_warnings,omitempty"`
	Selection        *CatalogSelection         `json:"selection,omitempty"`
}

type RequestStatus struct {
	CacheHit            bool   `json:"cache_hit"`
	RemainingAIRequests int    `json:"remaining_ai_requests"`
	CooldownSeconds     int    `json:"cooldown_seconds"`
	ChallengeRequired   bool   `json:"challenge_required"`
	ChallengePassed     bool   `json:"challenge_passed,omitempty"`
	RiskLevel           string `json:"risk_level,omitempty"`
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

type BuildRequestEcho struct {
	Budget    float64 `json:"budget"`
	UseCase   string  `json:"use_case"`
	BuildMode string  `json:"build_mode"`
	Notes     string  `json:"notes,omitempty"`
}

type BuildAdvice struct {
	Reasons       []string `json:"reasons"`
	Risks         []string `json:"risks"`
	UpgradeAdvice []string `json:"upgrade_advice"`
}

type BuildProductRef struct {
	DisplayName string  `json:"display_name"`
	Model       string  `json:"model"`
	Price       float64 `json:"price"`
	MinPrice    float64 `json:"min_price"`
	MaxPrice    float64 `json:"max_price"`
	SampleCount int     `json:"sample_count"`
}

type BuildRecommendationItem struct {
	Category           string            `json:"category"`
	TargetModel        string            `json:"target_model"`
	SelectionReason    string            `json:"selection_reason"`
	PriceBasis         string            `json:"price_basis"`
	Confidence         float64           `json:"confidence"`
	RecommendedProduct *BuildProductRef  `json:"recommended_product,omitempty"`
	CandidateProducts  []BuildProductRef `json:"candidate_products,omitempty"`
	Missing            bool              `json:"missing"`
	Reason             string            `json:"reason,omitempty"`
	SuggestedKeyword   string            `json:"suggested_keyword,omitempty"`
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

type SystemSettingsResponse struct {
	AIRuntime struct {
		BaseURL                string `json:"base_url"`
		Model                  string `json:"model"`
		TimeoutSeconds         int    `json:"timeout_seconds"`
		Enabled                bool   `json:"enabled"`
		GatewayTokenConfigured bool   `json:"gateway_token_configured"`
		GatewayTokenMasked     string `json:"gateway_token_masked"`
		APITokenConfigured     bool   `json:"api_token_configured"`
		APITokenMasked         string `json:"api_token_masked"`
	} `json:"ai_runtime"`
	CatalogAILimits struct {
		MaxModelsPerCategory int `json:"max_models_per_category"`
	} `json:"catalog_ai_limits"`
}

type UpdateSystemSettingsRequest struct {
	AIRuntime struct {
		BaseURL           string `json:"base_url,omitempty"`
		GatewayToken      string `json:"gateway_token,omitempty"`
		APIToken          string `json:"api_token,omitempty"`
		Model             string `json:"model,omitempty"`
		TimeoutSeconds    int    `json:"timeout_seconds,omitempty"`
		Enabled           *bool  `json:"enabled,omitempty"`
		ClearGatewayToken bool   `json:"clear_gateway_token,omitempty"`
		ClearAPIToken     bool   `json:"clear_api_token,omitempty"`
	} `json:"ai_runtime"`
	CatalogAILimits struct {
		MaxModelsPerCategory int `json:"max_models_per_category,omitempty"`
	} `json:"catalog_ai_limits"`
}

type AnonymousSessionResponse struct {
	AnonymousID             string `json:"anonymous_id"`
	CooldownSeconds         int    `json:"cooldown_seconds"`
	RemainingAIRequests     int    `json:"remaining_ai_requests"`
	ChallengeRequired       bool   `json:"challenge_required"`
	ChallengePassed         bool   `json:"challenge_passed,omitempty"`
	RiskLevel               string `json:"risk_level,omitempty"`
	SessionExpiresInSeconds int    `json:"session_expires_in_seconds,omitempty"`
}

type ChallengeVerifyRequest struct {
	AnonymousID       string `json:"anonymous_id"`
	DeviceFingerprint string `json:"device_fingerprint,omitempty"`
	ChallengeToken    string `json:"challenge_token"`
}

type ChallengeVerifyResponse struct {
	Verified             bool   `json:"verified"`
	PassExpiresInSeconds int    `json:"pass_expires_in_seconds,omitempty"`
	RiskLevel            string `json:"risk_level,omitempty"`
}

type PublicBootstrapResponse struct {
	ChallengeProvider string `json:"challenge_provider,omitempty"`
	ChallengeSiteKey  string `json:"challenge_site_key,omitempty"`
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

type CollectorScheduleConfig struct {
	ID                     string    `json:"id"`
	ServiceName            string    `json:"service_name"`
	Enabled                bool      `json:"enabled"`
	ScheduleTime           string    `json:"schedule_time"`
	RequestIntervalSeconds int       `json:"request_interval_seconds"`
	QueryLimit             int       `json:"query_limit"`
	CreatedAt              time.Time `json:"created_at,omitempty"`
	UpdatedAt              time.Time `json:"updated_at,omitempty"`
}

type CollectorScheduleResponse struct {
	Configured  bool                    `json:"configured"`
	ServiceName string                  `json:"service_name,omitempty"`
	Config      CollectorScheduleConfig `json:"config"`
}

type CollectorScheduleUpsertRequest struct {
	Enabled                bool   `json:"enabled"`
	ScheduleTime           string `json:"schedule_time"`
	RequestIntervalSeconds int    `json:"request_interval_seconds"`
	QueryLimit             int    `json:"query_limit"`
}
