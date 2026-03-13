package model

// GenerateBuildRequest is the console-facing request contract for build creation.
type GenerateBuildRequest struct {
	Budget    float64           `json:"budget"`
	UseCase   string            `json:"use_case"`
	BuildMode string            `json:"build_mode"`
	Pinned    []string          `json:"pinned_part_ids,omitempty"`
	Filters   map[string]string `json:"filters,omitempty"`
}

// BuildResponse is the console-facing aggregate returned to UI clients.
type BuildResponse struct {
	BuildID      string        `json:"build_id"`
	TotalPrice   float64       `json:"total_price"`
	Currency     string        `json:"currency"`
	Items        []BuildItem   `json:"items"`
	Advice       *Advice       `json:"advice,omitempty"`
	Alternatives []Alternative `json:"alternatives,omitempty"`
	Warnings     []string      `json:"warnings,omitempty"`
}

// CatalogRecommendationResponse is the user-facing recommendation assembled from the aggregated price catalog.
type CatalogRecommendationResponse struct {
	CatalogItemCount int              `json:"catalog_item_count"`
	CatalogWarnings  []string         `json:"catalog_warnings,omitempty"`
	Selection        CatalogSelection `json:"selection"`
	Advice           *Advice          `json:"advice,omitempty"`
}

// BuildItem represents a selected part on the console surface.
type BuildItem struct {
	Category    string   `json:"category"`
	DisplayName string   `json:"display_name"`
	UnitPrice   float64  `json:"unit_price"`
	Source      string   `json:"source"`
	Reasons     []string `json:"reasons,omitempty"`
	Risks       []string `json:"risks,omitempty"`
}

// CatalogSelection is the AI advisor's selected shopping list draft.
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

// Advice is the AI explanation block shown to users.
type Advice struct {
	Summary         string   `json:"summary"`
	Reasons         []string `json:"reasons"`
	FitFor          []string `json:"fit_for"`
	Risks           []string `json:"risks"`
	UpgradeAdvice   []string `json:"upgrade_advice"`
	AlternativeNote string   `json:"alternative_note"`
}

// Alternative is an additional build candidate.
type Alternative struct {
	BuildID    string  `json:"build_id"`
	Label      string  `json:"label"`
	TotalPrice float64 `json:"total_price"`
}

// PartSearchResult is the minimal part card returned by search.
type PartSearchResult struct {
	ID          string `json:"id"`
	Category    string `json:"category"`
	Brand       string `json:"brand"`
	Model       string `json:"model"`
	DisplayName string `json:"display_name"`
}

// AdminProduct is the minimal product card exposed on the console admin surface.
type AdminProduct struct {
	ID             string         `json:"id"`
	SourcePlatform string         `json:"source_platform"`
	ExternalID     string         `json:"external_id"`
	SKUID          string         `json:"sku_id"`
	Title          string         `json:"title"`
	ShopName       string         `json:"shop_name"`
	ShopType       string         `json:"shop_type"`
	Price          float64        `json:"price"`
	Currency       string         `json:"currency"`
	Availability   string         `json:"availability"`
	Attributes     map[string]any `json:"attributes,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
}

type AdminProductFilter struct {
	Keyword          string `json:"keyword,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	ShopType         string `json:"shop_type,omitempty"`
	RealOnly         bool   `json:"real_only,omitempty"`
	SelfOperatedOnly bool   `json:"self_operated_only,omitempty"`
}

// AdminJob is the minimal collector job card exposed on the console admin surface.
type AdminJob struct {
	ID             string         `json:"id"`
	JobType        string         `json:"job_type"`
	Status         string         `json:"status"`
	SourcePlatform string         `json:"source_platform"`
	Payload        map[string]any `json:"payload,omitempty"`
	Result         map[string]any `json:"result,omitempty"`
	RetryCount     int            `json:"retry_count"`
	ErrorMessage   string         `json:"error_message,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	UpdatedAt      string         `json:"updated_at,omitempty"`
}

// AdminCollectRequest is the console-facing trigger contract for collector jobs.
type AdminCollectRequest struct {
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Brand    string `json:"brand,omitempty"`
	Limit    int    `json:"limit"`
	Persist  bool   `json:"persist"`
}

type AdminCollectBatchRequest struct {
	Preset   string                `json:"preset,omitempty"`
	Persist  *bool                 `json:"persist,omitempty"`
	Requests []AdminCollectRequest `json:"requests,omitempty"`
}

// AdminCollectResponse is the normalized collector trigger result.
type AdminCollectResponse struct {
	JobID            string         `json:"job_id"`
	RetriedFromJobID string         `json:"retried_from_job_id,omitempty"`
	Mode             string         `json:"mode"`
	Persisted        bool           `json:"persisted"`
	PersistedCount   int            `json:"persisted_count"`
	Products         []AdminProduct `json:"products,omitempty"`
}

type AdminCollectBatchResponse struct {
	Preset         string                  `json:"preset,omitempty"`
	Mode           string                  `json:"mode"`
	TotalJobs      int                     `json:"total_jobs"`
	SuccessfulJobs int                     `json:"successful_jobs"`
	FailedJobs     int                     `json:"failed_jobs"`
	SkippedJobs    int                     `json:"skipped_jobs"`
	AbortedJobs    int                     `json:"aborted_jobs"`
	TotalPersisted int                     `json:"total_persisted"`
	Results        []AdminCollectBatchItem `json:"results,omitempty"`
}

type AdminCollectBatchItem struct {
	Keyword  string                `json:"keyword"`
	Category string                `json:"category"`
	Brand    string                `json:"brand,omitempty"`
	Skipped  bool                  `json:"skipped,omitempty"`
	Aborted  bool                  `json:"aborted,omitempty"`
	Note     string                `json:"note,omitempty"`
	Response *AdminCollectResponse `json:"response,omitempty"`
	Error    string                `json:"error,omitempty"`
}

type GoofishStateFile struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	IsRoot bool   `json:"is_root"`
}

type GoofishValidateRequest struct {
	Strategy         string `json:"strategy,omitempty"`
	AccountStateFile string `json:"account_state_file,omitempty"`
}

type GoofishValidateResponse struct {
	Valid       bool   `json:"valid"`
	StateFile   string `json:"state_file"`
	Keyword     string `json:"keyword"`
	SampleCount int    `json:"sample_count"`
	PageURL     string `json:"page_url,omitempty"`
}

// BuildEngineResponse mirrors the build-engine response contract.
type BuildEngineResponse struct {
	BuildRequestID string              `json:"build_request_id"`
	RequestNo      string              `json:"request_no,omitempty"`
	Budget         float64             `json:"budget"`
	UseCase        string              `json:"use_case"`
	BuildMode      string              `json:"build_mode"`
	Status         string              `json:"status,omitempty"`
	Warnings       []string            `json:"warnings,omitempty"`
	Results        []BuildEngineResult `json:"results"`
}

// BuildEngineResult is one generated candidate returned from build-engine.
type BuildEngineResult struct {
	ResultID      string                     `json:"result_id"`
	Role          string                     `json:"role"`
	TotalPrice    float64                    `json:"total_price"`
	Score         float64                    `json:"score"`
	Currency      string                     `json:"currency"`
	Items         []BuildEngineResultItem    `json:"items"`
	Compatibility []BuildEngineCompatibility `json:"compatibility,omitempty"`
}

// BuildEngineResultItem is one item from build-engine output.
type BuildEngineResultItem struct {
	Category       string   `json:"category"`
	DisplayName    string   `json:"display_name"`
	UnitPrice      float64  `json:"unit_price"`
	SourcePlatform string   `json:"source_platform"`
	Reasons        []string `json:"reasons,omitempty"`
	Risks          []string `json:"risks,omitempty"`
}

// BuildEngineCompatibility is a compatibility finding from build-engine.
type BuildEngineCompatibility struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Passed   bool   `json:"passed"`
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

// AIAdvisorResponse mirrors the ai-advisor response contract.
type AIAdvisorResponse struct {
	BuildRequestID string `json:"build_request_id"`
	ResultID       string `json:"result_id"`
	Provider       string `json:"provider"`
	FallbackUsed   bool   `json:"fallback_used"`
	Advisory       Advice `json:"advisory"`
}

// AIAdvisorCatalogResponse mirrors ai-advisor's catalog recommendation response.
type AIAdvisorCatalogResponse struct {
	Provider     string           `json:"provider"`
	FallbackUsed bool             `json:"fallback_used"`
	Selection    CatalogSelection `json:"selection"`
	Advisory     Advice           `json:"advisory"`
}
