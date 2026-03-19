export interface BrandPreference {
  cpu: string;
  gpu: string;
}

export interface GenerateBuildRequest {
  budget: number;
  use_case: string;
  build_mode: string;
  brand_preference: BrandPreference;
  special_requirements: string[];
  notes: string;
}

export interface RequestStatus {
  cache_hit: boolean;
  remaining_ai_requests: number;
  cooldown_seconds: number;
}

export interface AnonymousSessionResponse {
  anonymous_id: string;
  cooldown_seconds: number;
  remaining_ai_requests: number;
  challenge_required: boolean;
}

export interface CatalogRecommendationItem {
  category: string;
  display_name: string;
  normalized_key: string;
  sample_count: number;
  selected_price: number;
  median_price: number;
  source_platforms?: string[];
  reasons?: string[];
}

export interface CatalogSelection {
  budget: number;
  use_case: string;
  build_mode: string;
  estimated_total: number;
  warnings?: string[];
  selected_items: CatalogRecommendationItem[];
}

export interface Advice {
  summary: string;
  reasons: string[];
  fit_for: string[];
  risks: string[];
  upgrade_advice: string[];
  alternative_note: string;
}

export interface CatalogRecommendationResponse {
  request_status: RequestStatus;
  provider: string;
  fallback_used: boolean;
  request: {
    budget: number;
    use_case: string;
    build_mode: string;
    notes?: string;
  };
  summary: string;
  estimated_total: number;
  within_budget: boolean;
  warnings?: string[];
  build_items: {
    category: string;
    target_model: string;
    selection_reason: string;
    price_basis: string;
    confidence: number;
    recommended_product?: {
      display_name: string;
      model: string;
      price: number;
      min_price: number;
      max_price: number;
      sample_count: number;
    };
    candidate_products?: Array<{
      display_name: string;
      model: string;
      price: number;
      min_price: number;
      max_price: number;
      sample_count: number;
    }>;
    missing: boolean;
    reason?: string;
    suggested_keyword?: string;
  }[];
  advice: {
    reasons: string[];
    risks: string[];
    upgrade_advice: string[];
  };
  catalog_item_count?: number;
  catalog_warnings?: string[];
  selection?: CatalogSelection;
}

export interface KeywordSeed {
  id: string;
  category: string;
  keyword: string;
  canonical_model: string;
  brand?: string;
  aliases?: string[];
  priority: number;
  enabled: boolean;
  notes?: string;
  created_at?: string;
  updated_at?: string;
}

export interface KeywordSeedListResponse {
  items: KeywordSeed[];
  page: number;
  page_size: number;
  total: number;
}

export interface KeywordSeedUpsertRequest {
  category: string;
  keyword: string;
  canonical_model: string;
  brand: string;
  aliases: string[];
  priority: number;
  enabled: boolean;
  notes: string;
}

export interface AdminLoginResponse {
  username: string;
}

export interface KeywordSeedImportError {
  row: number;
  message: string;
}

export interface KeywordSeedImportResponse {
  job_id: string;
  imported_count: number;
  failed_count: number;
  errors?: KeywordSeedImportError[];
}

export interface SystemSettingsResponse {
  ai_runtime: {
    base_url: string;
    model: string;
    timeout_seconds: number;
    enabled: boolean;
    gateway_token_configured: boolean;
    gateway_token_masked: string;
    api_token_configured: boolean;
    api_token_masked: string;
  };
  catalog_ai_limits: {
    max_models_per_category: number;
  };
}

export interface UpdateSystemSettingsRequest {
  ai_runtime?: {
    base_url?: string;
    gateway_token?: string;
    api_token?: string;
    model?: string;
    timeout_seconds?: number;
    enabled?: boolean;
    clear_gateway_token?: boolean;
    clear_api_token?: boolean;
  };
  catalog_ai_limits?: {
    max_models_per_category?: number;
  };
}
