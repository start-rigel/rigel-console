package model

// GenerateBuildRequest is the console-facing request contract for recommendation creation.
type GenerateBuildRequest struct {
	Budget    float64 `json:"budget"`
	UseCase   string  `json:"use_case"`
	BuildMode string  `json:"build_mode"`
}

// CatalogRecommendationResponse is the user-facing recommendation assembled from the aggregated price catalog.
type CatalogRecommendationResponse struct {
	CatalogItemCount int              `json:"catalog_item_count"`
	CatalogWarnings  []string         `json:"catalog_warnings,omitempty"`
	Selection        CatalogSelection `json:"selection"`
	Advice           *Advice          `json:"advice,omitempty"`
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
