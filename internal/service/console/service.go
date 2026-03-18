package console

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
	"github.com/xuri/excelize/v2"
)

type BuildEngineClient interface {
	GetPriceCatalog(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEnginePriceCatalog, error)
	GenerateCatalogAdvice(ctx context.Context, req model.GenerateBuildRequest, catalog model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error)
}

type cachedRecommendation struct {
	response  model.CatalogRecommendationResponse
	expiresAt time.Time
}

type sessionUsage struct {
	WindowStarted time.Time
	Used          int
	CooldownUntil time.Time
}

type Service struct {
	buildClient      BuildEngineClient
	adminUsername    string
	adminPassword    string
	anonymousLimit   int
	cooldownDuration time.Duration
	cacheTTL         time.Duration

	mu             sync.Mutex
	seedSeq        int
	seeds          map[string]model.KeywordSeed
	recommendCache map[string]cachedRecommendation
	anonymousUsage map[string]sessionUsage
}

func New(buildClient BuildEngineClient, adminUsername, adminPassword string, anonymousLimit int, cooldownDuration time.Duration) *Service {
	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin123456"
	}
	if anonymousLimit <= 0 {
		anonymousLimit = 5
	}
	if cooldownDuration <= 0 {
		cooldownDuration = time.Minute
	}

	now := time.Now().UTC()
	initialSeeds := []model.KeywordSeed{
		newSeed("seed-1", "cpu", "Ryzen 5 7500F", "Ryzen 5 7500F", "AMD", []string{"7500F", "AMD 7500F"}, 100, true, "主流游戏 CPU", now),
		newSeed("seed-2", "gpu", "RTX 4060", "RTX 4060", "NVIDIA", []string{"4060", "RTX4060"}, 100, true, "1080p 主流显卡", now),
		newSeed("seed-3", "motherboard", "B650M", "B650M", "", []string{"B650", "AMD B650M"}, 90, true, "AM5 主流主板", now),
	}
	seeds := make(map[string]model.KeywordSeed, len(initialSeeds))
	for _, seed := range initialSeeds {
		seeds[seed.ID] = seed
	}

	return &Service{
		buildClient:      buildClient,
		adminUsername:    adminUsername,
		adminPassword:    adminPassword,
		anonymousLimit:   anonymousLimit,
		cooldownDuration: cooldownDuration,
		cacheTTL:         10 * time.Minute,
		seedSeq:          len(initialSeeds),
		seeds:            seeds,
		recommendCache:   make(map[string]cachedRecommendation),
		anonymousUsage:   make(map[string]sessionUsage),
	}
}

func newSeed(id, category, keyword, canonicalModel, brand string, aliases []string, priority int, enabled bool, notes string, now time.Time) model.KeywordSeed {
	return model.KeywordSeed{
		ID:             id,
		Category:       category,
		Keyword:        keyword,
		CanonicalModel: canonicalModel,
		Brand:          brand,
		Aliases:        aliases,
		Priority:       priority,
		Enabled:        enabled,
		Notes:          notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (s *Service) IssueAnonymousSession(anonymousID string) model.AnonymousSessionResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	if anonymousID == "" {
		anonymousID = "anon_" + randomToken(12)
	}

	status := s.currentUsageLocked(anonymousID)
	return model.AnonymousSessionResponse{
		AnonymousID:         anonymousID,
		CooldownSeconds:     cooldownSeconds(status.CooldownUntil),
		RemainingAIRequests: max(0, s.anonymousLimit-status.Used),
		ChallengeRequired:   false,
	}
}

func (s *Service) AuthenticateAdmin(username, password string) bool {
	return username == s.adminUsername && password == s.adminPassword
}

func (s *Service) GenerateCatalogRecommendation(ctx context.Context, req model.GenerateBuildRequest, anonymousID string) (model.CatalogRecommendationResponse, error) {
	if anonymousID == "" {
		anonymousID = "anon_" + randomToken(12)
	}

	key, err := recommendationKey(req)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}

	now := time.Now()

	s.mu.Lock()
	usage := s.currentUsageLocked(anonymousID)
	if usage.CooldownUntil.After(now) {
		s.mu.Unlock()
		return model.CatalogRecommendationResponse{
			RequestStatus: model.RequestStatus{
				CacheHit:            false,
				RemainingAIRequests: max(0, s.anonymousLimit-usage.Used),
				CooldownSeconds:     cooldownSeconds(usage.CooldownUntil),
			},
		}, ErrRateLimited{CooldownSeconds: cooldownSeconds(usage.CooldownUntil)}
	}
	if cached, ok := s.recommendCache[key]; ok && cached.expiresAt.After(now) {
		response := cached.response
		response.RequestStatus = model.RequestStatus{
			CacheHit:            true,
			RemainingAIRequests: max(0, s.anonymousLimit-usage.Used),
			CooldownSeconds:     0,
		}
		s.mu.Unlock()
		return response, nil
	}
	if usage.Used >= s.anonymousLimit {
		usage.CooldownUntil = now.Add(s.cooldownDuration)
		s.anonymousUsage[anonymousID] = usage
		s.mu.Unlock()
		return model.CatalogRecommendationResponse{
			RequestStatus: model.RequestStatus{
				CacheHit:            false,
				RemainingAIRequests: 0,
				CooldownSeconds:     cooldownSeconds(usage.CooldownUntil),
			},
		}, ErrRateLimited{CooldownSeconds: cooldownSeconds(usage.CooldownUntil)}
	}
	usage.Used++
	s.anonymousUsage[anonymousID] = usage
	remaining := max(0, s.anonymousLimit-usage.Used)
	s.mu.Unlock()

	catalog, err := s.buildClient.GetPriceCatalog(ctx, req)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	advice, err := s.buildClient.GenerateCatalogAdvice(ctx, req, catalog)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	response := model.CatalogRecommendationResponse{
		RequestStatus: model.RequestStatus{
			CacheHit:            false,
			RemainingAIRequests: remaining,
			CooldownSeconds:     0,
		},
		CatalogItemCount: len(catalog.Items),
		CatalogWarnings:  catalog.Warnings,
		Selection:        advice.Selection,
		Advice:           &advice.Advisory,
	}

	s.mu.Lock()
	s.recommendCache[key] = cachedRecommendation{response: response, expiresAt: now.Add(s.cacheTTL)}
	s.mu.Unlock()
	return response, nil
}

func (s *Service) ListKeywordSeeds(filter model.KeywordSeedFilter) model.KeywordSeedListResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	items := make([]model.KeywordSeed, 0, len(s.seeds))
	for _, item := range s.seeds {
		if filter.Category != "" && !strings.EqualFold(item.Category, filter.Category) {
			continue
		}
		if filter.Brand != "" && !strings.Contains(strings.ToLower(item.Brand), strings.ToLower(filter.Brand)) {
			continue
		}
		if filter.Keyword != "" {
			text := strings.ToLower(item.Keyword + " " + item.CanonicalModel + " " + strings.Join(item.Aliases, " "))
			if !strings.Contains(text, strings.ToLower(filter.Keyword)) {
				continue
			}
		}
		if filter.Enabled != nil && item.Enabled != *filter.Enabled {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Priority != items[j].Priority {
			return items[i].Priority > items[j].Priority
		}
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})

	total := len(items)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return model.KeywordSeedListResponse{
		Items:    append([]model.KeywordSeed(nil), items[start:end]...),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}

func (s *Service) GetKeywordSeed(id string) (model.KeywordSeed, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.seeds[id]
	return item, ok
}

func (s *Service) CreateKeywordSeed(req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	if err := validateSeed(req); err != nil {
		return model.KeywordSeed{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.seedSeq++
	id := fmt.Sprintf("seed-%d", s.seedSeq)
	seed := model.KeywordSeed{
		ID:             id,
		Category:       strings.ToLower(strings.TrimSpace(req.Category)),
		Keyword:        strings.TrimSpace(req.Keyword),
		CanonicalModel: strings.TrimSpace(req.CanonicalModel),
		Brand:          strings.TrimSpace(req.Brand),
		Aliases:        normalizeAliases(req.Aliases),
		Priority:       normalizePriority(req.Priority),
		Enabled:        req.Enabled,
		Notes:          strings.TrimSpace(req.Notes),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.seeds[id] = seed
	return seed, nil
}

func (s *Service) UpdateKeywordSeed(id string, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	if err := validateSeed(req); err != nil {
		return model.KeywordSeed{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.seeds[id]
	if !ok {
		return model.KeywordSeed{}, ErrNotFound{Resource: "keyword seed"}
	}
	existing.Category = strings.ToLower(strings.TrimSpace(req.Category))
	existing.Keyword = strings.TrimSpace(req.Keyword)
	existing.CanonicalModel = strings.TrimSpace(req.CanonicalModel)
	existing.Brand = strings.TrimSpace(req.Brand)
	existing.Aliases = normalizeAliases(req.Aliases)
	existing.Priority = normalizePriority(req.Priority)
	existing.Enabled = req.Enabled
	existing.Notes = strings.TrimSpace(req.Notes)
	existing.UpdatedAt = time.Now().UTC()
	s.seeds[id] = existing
	return existing, nil
}

func (s *Service) SetKeywordSeedEnabled(id string, enabled bool) (model.KeywordSeed, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.seeds[id]
	if !ok {
		return model.KeywordSeed{}, ErrNotFound{Resource: "keyword seed"}
	}
	item.Enabled = enabled
	item.UpdatedAt = time.Now().UTC()
	s.seeds[id] = item
	return item, nil
}

func (s *Service) ExportKeywordSeedsExcel() ([]byte, error) {
	s.mu.Lock()
	items := make([]model.KeywordSeed, 0, len(s.seeds))
	for _, item := range s.seeds {
		items = append(items, item)
	}
	s.mu.Unlock()

	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	headers := []string{"category", "keyword", "canonical_model", "brand", "aliases", "priority", "enabled", "notes"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, header)
	}
	for rowIndex, item := range items {
		row := rowIndex + 2
		values := []any{
			item.Category,
			item.Keyword,
			item.CanonicalModel,
			item.Brand,
			strings.Join(item.Aliases, ","),
			item.Priority,
			item.Enabled,
			item.Notes,
		}
		for i, value := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, row)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write export workbook: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *Service) TemplateWorkbook() ([]byte, error) {
	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	rows := [][]any{
		{"category", "keyword", "canonical_model", "brand", "aliases", "priority", "enabled", "notes"},
		{"cpu", "Ryzen 5 7500F", "Ryzen 5 7500F", "AMD", "7500F,AMD 7500F", 100, true, "主流游戏 CPU"},
		{"gpu", "RTX 4060", "RTX 4060", "NVIDIA", "4060,RTX4060", 100, true, "1080p 主流显卡"},
	}
	for rowIndex, row := range rows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			_ = f.SetCellValue(sheet, cell, value)
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("write template workbook: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *Service) ImportKeywordSeeds(file multipart.File) (model.KeywordSeedImportResponse, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("read upload: %w", err)
	}
	return s.importWorkbookBytes(content)
}

func (s *Service) importWorkbookBytes(content []byte) (model.KeywordSeedImportResponse, error) {
	workbook, err := excelize.OpenReader(bytes.NewReader(content))
	if err != nil {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("open workbook: %w", err)
	}
	defer func() { _ = workbook.Close() }()

	sheets := workbook.GetSheetList()
	if len(sheets) == 0 {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("workbook has no sheets")
	}
	rows, err := workbook.GetRows(sheets[0])
	if err != nil {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("read rows: %w", err)
	}
	if len(rows) == 0 {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("workbook is empty")
	}

	headers := rows[0]
	expected := []string{"category", "keyword", "canonical_model", "brand", "aliases", "priority", "enabled", "notes"}
	for i, header := range expected {
		if i >= len(headers) || strings.TrimSpace(headers[i]) != header {
			return model.KeywordSeedImportResponse{}, fmt.Errorf("invalid header at column %d", i+1)
		}
	}

	result := model.KeywordSeedImportResponse{JobID: "job_" + randomToken(8)}
	for i := 1; i < len(rows); i++ {
		row := normalizeRow(rows[i], len(expected))
		req, err := parseSeedRow(row)
		if err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, model.KeywordSeedImportError{Row: i + 1, Message: err.Error()})
			continue
		}
		if _, err := s.CreateKeywordSeed(req); err != nil {
			result.FailedCount++
			result.Errors = append(result.Errors, model.KeywordSeedImportError{Row: i + 1, Message: err.Error()})
			continue
		}
		result.ImportedCount++
	}
	return result, nil
}

func normalizeRow(row []string, size int) []string {
	out := make([]string, size)
	copy(out, row)
	return out
}

func parseSeedRow(row []string) (model.KeywordSeedUpsertRequest, error) {
	priority := 100
	if strings.TrimSpace(row[5]) != "" {
		parsed, err := strconv.Atoi(strings.TrimSpace(row[5]))
		if err != nil {
			return model.KeywordSeedUpsertRequest{}, fmt.Errorf("priority is invalid")
		}
		priority = parsed
	}
	enabled := true
	if strings.TrimSpace(row[6]) != "" {
		switch strings.ToLower(strings.TrimSpace(row[6])) {
		case "true":
			enabled = true
		case "false":
			enabled = false
		default:
			return model.KeywordSeedUpsertRequest{}, fmt.Errorf("enabled is invalid")
		}
	}
	req := model.KeywordSeedUpsertRequest{
		Category:       row[0],
		Keyword:        row[1],
		CanonicalModel: row[2],
		Brand:          row[3],
		Aliases:        splitAliases(row[4]),
		Priority:       priority,
		Enabled:        enabled,
		Notes:          row[7],
	}
	return req, validateSeed(req)
}

func recommendationKey(req model.GenerateBuildRequest) (string, error) {
	type normalizedRequest struct {
		Budget              float64           `json:"budget"`
		UseCase             string            `json:"use_case"`
		BuildMode           string            `json:"build_mode"`
		BrandPreference     map[string]string `json:"brand_preference,omitempty"`
		SpecialRequirements []string          `json:"special_requirements,omitempty"`
		Notes               string            `json:"notes,omitempty"`
	}

	special := append([]string(nil), req.SpecialRequirements...)
	sort.Strings(special)
	brand := map[string]string{}
	if req.BrandPreference.CPU != "" {
		brand["cpu"] = strings.ToLower(strings.TrimSpace(req.BrandPreference.CPU))
	}
	if req.BrandPreference.GPU != "" {
		brand["gpu"] = strings.ToLower(strings.TrimSpace(req.BrandPreference.GPU))
	}
	payload := normalizedRequest{
		Budget:              req.Budget,
		UseCase:             strings.ToLower(strings.TrimSpace(req.UseCase)),
		BuildMode:           strings.ToLower(strings.TrimSpace(req.BuildMode)),
		BrandPreference:     brand,
		SpecialRequirements: special,
		Notes:               strings.TrimSpace(req.Notes),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal recommendation key: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func validateSeed(req model.KeywordSeedUpsertRequest) error {
	category := strings.ToLower(strings.TrimSpace(req.Category))
	switch category {
	case "cpu", "gpu", "motherboard", "ram", "ssd", "psu", "case", "cooler":
	default:
		return fmt.Errorf("category is invalid")
	}
	if strings.TrimSpace(req.Keyword) == "" {
		return fmt.Errorf("keyword is required")
	}
	if strings.TrimSpace(req.CanonicalModel) == "" {
		return fmt.Errorf("canonical_model is required")
	}
	return nil
}

func normalizeAliases(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		key := strings.ToLower(value)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, value)
	}
	return out
}

func splitAliases(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}

func normalizePriority(priority int) int {
	if priority <= 0 {
		return 100
	}
	return priority
}

func randomToken(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(buf)[:size]
}

func cooldownSeconds(until time.Time) int {
	if until.IsZero() {
		return 0
	}
	seconds := int(time.Until(until).Seconds())
	if seconds < 0 {
		return 0
	}
	return seconds
}

func (s *Service) currentUsageLocked(anonymousID string) sessionUsage {
	now := time.Now()
	usage, ok := s.anonymousUsage[anonymousID]
	if !ok || now.Sub(usage.WindowStarted) >= time.Hour {
		usage = sessionUsage{WindowStarted: now}
		s.anonymousUsage[anonymousID] = usage
		return usage
	}
	return usage
}

type ErrRateLimited struct {
	CooldownSeconds int
}

func (e ErrRateLimited) Error() string {
	return "rate limited"
}

type ErrNotFound struct {
	Resource string
}

func (e ErrNotFound) Error() string {
	if e.Resource == "" {
		return "not found"
	}
	return e.Resource + " not found"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
