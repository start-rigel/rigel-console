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
  catalog_item_count: number;
  catalog_warnings?: string[];
  selection: CatalogSelection;
  advice?: Advice;
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

export interface CollectorScheduleConfig {
  id: string;
  service_name: string;
  enabled: boolean;
  schedule_time: string;
  request_interval_seconds: number;
  query_limit: number;
  created_at?: string;
  updated_at?: string;
}

export interface CollectorScheduleResponse {
  configured: boolean;
  service_name?: string;
  config?: CollectorScheduleConfig;
}

export interface CollectorScheduleUpsertRequest {
  enabled: boolean;
  schedule_time: string;
  request_interval_seconds: number;
  query_limit: number;
}
