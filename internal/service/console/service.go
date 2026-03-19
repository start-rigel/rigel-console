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
	"time"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
	"github.com/xuri/excelize/v2"
)

type BuildEngineClient interface {
	GetPriceCatalog(ctx context.Context, req model.GenerateBuildRequest) (model.BuildEnginePriceCatalog, error)
	GenerateCatalogAdvice(ctx context.Context, req model.GenerateBuildRequest, catalog model.BuildEnginePriceCatalog) (model.CatalogAdviceResponse, error)
}

type JDCollectorClient interface {
	GetScheduleConfig(ctx context.Context) (model.CollectorScheduleResponse, error)
	UpdateScheduleConfig(ctx context.Context, payload model.CollectorScheduleUpsertRequest) (model.CollectorScheduleResponse, error)
}

type RequestMeta struct {
	ClientIP          string
	DeviceFingerprint string
	AnonymousID       string
}

type Option func(*Service)

func WithStore(store securityStore) Option {
	return func(s *Service) {
		if store != nil {
			s.store = store
		}
	}
}

func WithChallengeVerifier(verifier challengeVerifier) Option {
	return func(s *Service) {
		if verifier != nil {
			s.challengeVerifier = verifier
		}
	}
}

func WithLimits(ipHourlyLimit, deviceHourlyLimit int) Option {
	return func(s *Service) {
		if ipHourlyLimit > 0 {
			s.ipHourlyLimit = ipHourlyLimit
		}
		if deviceHourlyLimit > 0 {
			s.deviceHourlyLimit = deviceHourlyLimit
		}
	}
}

func WithChallengePassTTL(ttl time.Duration) Option {
	return func(s *Service) {
		if ttl > 0 {
			s.challengePassTTL = ttl
		}
	}
}

func WithSessionTTL(ttl time.Duration) Option {
	return func(s *Service) {
		if ttl > 0 {
			s.sessionTTL = ttl
		}
	}
}

func WithKeywordSeedRepository(repo KeywordSeedRepository) Option {
	return func(s *Service) {
		if repo != nil {
			s.keywordSeeds = repo
		}
	}
}

func WithAdminPasswordHash(hash string) Option {
	return func(s *Service) {
		hash = strings.TrimSpace(hash)
		if hash != "" {
			s.adminPasswordHash = []byte(hash)
		}
	}
}

type Service struct {
	buildClient       BuildEngineClient
	jdCollector       JDCollectorClient
	adminUsername     string
	adminPasswordHash []byte
	anonymousLimit    int
	ipHourlyLimit     int
	deviceHourlyLimit int
	cooldownDuration  time.Duration
	cacheTTL          time.Duration
	challengePassTTL  time.Duration
	sessionTTL        time.Duration
	store             securityStore
	challengeVerifier challengeVerifier
	keywordSeeds      KeywordSeedRepository
}

func New(buildClient BuildEngineClient, jdCollector JDCollectorClient, adminUsername, adminPassword string, anonymousLimit int, cooldownDuration time.Duration, opts ...Option) *Service {
	if anonymousLimit <= 0 {
		anonymousLimit = 5
	}
	if cooldownDuration <= 0 {
		cooldownDuration = time.Minute
	}
	var adminPasswordHash []byte
	if strings.TrimSpace(adminPassword) != "" {
		var err error
		adminPasswordHash, err = hashAdminPassword(adminPassword)
		if err != nil {
			panic(fmt.Sprintf("hash admin password: %v", err))
		}
	}

	service := &Service{
		buildClient:       buildClient,
		jdCollector:       jdCollector,
		adminUsername:     adminUsername,
		adminPasswordHash: adminPasswordHash,
		anonymousLimit:    anonymousLimit,
		ipHourlyLimit:     max(anonymousLimit*4, 20),
		deviceHourlyLimit: max(anonymousLimit*2, 12),
		cooldownDuration:  cooldownDuration,
		cacheTTL:          10 * time.Minute,
		challengePassTTL:  15 * time.Minute,
		sessionTTL:        30 * 24 * time.Hour,
		store:             newMemorySecurityStore(),
		challengeVerifier: noopChallengeVerifier{},
	}

	for _, opt := range opts {
		opt(service)
	}
	if len(service.adminPasswordHash) == 0 {
		panic("admin password hash is not configured")
	}
	return service
}

func (s *Service) GetCollectorScheduleConfig(ctx context.Context) (model.CollectorScheduleResponse, error) {
	if s.jdCollector == nil {
		return model.CollectorScheduleResponse{}, fmt.Errorf("jd collector client is not configured")
	}
	return s.jdCollector.GetScheduleConfig(ctx)
}

func (s *Service) UpdateCollectorScheduleConfig(ctx context.Context, req model.CollectorScheduleUpsertRequest) (model.CollectorScheduleResponse, error) {
	if s.jdCollector == nil {
		return model.CollectorScheduleResponse{}, fmt.Errorf("jd collector client is not configured")
	}
	return s.jdCollector.UpdateScheduleConfig(ctx, req)
}

func (s *Service) IssueAnonymousSession(ctx context.Context, meta RequestMeta) (model.AnonymousSessionResponse, error) {
	anonymousID := strings.TrimSpace(meta.AnonymousID)
	if anonymousID == "" {
		anonymousID = "anon_" + randomToken(12)
	}
	meta.AnonymousID = anonymousID

	sessionUsage, _, err := s.loadUsage(ctx, "session", anonymousID)
	if err != nil {
		return model.AnonymousSessionResponse{}, err
	}
	challengePassed, err := s.hasChallengePass(ctx, meta)
	if err != nil {
		return model.AnonymousSessionResponse{}, err
	}
	riskLevel := s.evaluateRiskLevel(meta)
	challengeRequired := !challengePassed && s.challengeVerifier.Available() && riskLevel == riskLevelElevated

	return model.AnonymousSessionResponse{
		AnonymousID:             anonymousID,
		CooldownSeconds:         cooldownSeconds(sessionUsage.CooldownUntil),
		RemainingAIRequests:     max(0, s.anonymousLimit-sessionUsage.Used),
		ChallengeRequired:       challengeRequired,
		ChallengePassed:         challengePassed,
		RiskLevel:               riskLevel,
		SessionExpiresInSeconds: int(s.sessionTTL.Seconds()),
	}, nil
}

func (s *Service) GenerateCatalogRecommendation(ctx context.Context, req model.GenerateBuildRequest, meta RequestMeta) (model.CatalogRecommendationResponse, error) {
	meta.AnonymousID = strings.TrimSpace(meta.AnonymousID)
	if meta.AnonymousID == "" {
		meta.AnonymousID = "anon_" + randomToken(12)
	}

	key, err := recommendationKey(req)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}

	riskLevel := s.evaluateRiskLevel(meta)
	sessionUsage, sessionExists, err := s.loadUsage(ctx, "session", meta.AnonymousID)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	ipUsage, ipExists, err := s.loadUsage(ctx, "ip", meta.ClientIP)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	deviceUsage, deviceExists, err := s.loadUsage(ctx, "device", s.deviceKey(meta.DeviceFingerprint))
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	challengePassed, err := s.hasChallengePass(ctx, meta)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}

	if blocked := firstCooldown(sessionUsage, ipUsage, deviceUsage); !blocked.IsZero() {
		return model.CatalogRecommendationResponse{
			RequestStatus: s.requestStatus(false, sessionUsage, blocked, challengePassed, riskLevel),
		}, ErrRateLimited{CooldownSeconds: cooldownSeconds(blocked)}
	}

	if cached, ok, err := s.store.LoadRecommendation(ctx, key); err != nil {
		return model.CatalogRecommendationResponse{}, err
	} else if ok && cached.ExpiresAt.After(time.Now()) {
		response := cached.Response
		response.RequestStatus = s.requestStatus(true, sessionUsage, time.Time{}, challengePassed, riskLevel)
		return response, nil
	}

	requiresChallenge := !challengePassed && riskLevel == riskLevelElevated && s.challengeVerifier.Available()
	if requiresChallenge {
		return model.CatalogRecommendationResponse{
			RequestStatus: s.requestStatus(false, sessionUsage, time.Time{}, false, riskLevel),
		}, ErrChallengeRequired{RiskLevel: riskLevel}
	}

	now := time.Now()
	sessionUsage = ensureUsageWindow(sessionUsage, sessionExists, now)
	ipUsage = ensureUsageWindow(ipUsage, ipExists, now)
	deviceUsage = ensureUsageWindow(deviceUsage, deviceExists, now)

	if sessionUsage.Used >= s.anonymousLimit || ipUsage.Used >= s.ipHourlyLimit || deviceUsage.Used >= s.deviceHourlyLimit {
		cooldownUntil := now.Add(s.cooldownDuration)
		sessionUsage.CooldownUntil = maxTime(sessionUsage.CooldownUntil, cooldownUntil)
		ipUsage.CooldownUntil = maxTime(ipUsage.CooldownUntil, cooldownUntil)
		deviceUsage.CooldownUntil = maxTime(deviceUsage.CooldownUntil, cooldownUntil)
		if err := s.saveUsageTriplet(ctx, meta, sessionUsage, ipUsage, deviceUsage); err != nil {
			return model.CatalogRecommendationResponse{}, err
		}
		return model.CatalogRecommendationResponse{
			RequestStatus: s.requestStatus(false, sessionUsage, cooldownUntil, challengePassed, riskLevel),
		}, ErrRateLimited{CooldownSeconds: cooldownSeconds(cooldownUntil)}
	}

	sessionUsage.Used++
	ipUsage.Used++
	deviceUsage.Used++
	if err := s.saveUsageTriplet(ctx, meta, sessionUsage, ipUsage, deviceUsage); err != nil {
		return model.CatalogRecommendationResponse{}, err
	}

	catalog, err := s.buildClient.GetPriceCatalog(ctx, req)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	advice, err := s.buildClient.GenerateCatalogAdvice(ctx, req, catalog)
	if err != nil {
		return model.CatalogRecommendationResponse{}, err
	}

	response := model.CatalogRecommendationResponse{
		RequestStatus:    s.requestStatus(false, sessionUsage, time.Time{}, challengePassed, riskLevel),
		CatalogItemCount: len(catalog.Items),
		CatalogWarnings:  catalog.Warnings,
		Selection:        advice.Selection,
		Advice:           &advice.Advisory,
	}
	response.RequestStatus.RemainingAIRequests = max(0, s.anonymousLimit-sessionUsage.Used)

	if err := s.store.SaveRecommendation(ctx, key, storedRecommendation{
		Response:  response,
		ExpiresAt: now.Add(s.cacheTTL),
	}, s.cacheTTL); err != nil {
		return model.CatalogRecommendationResponse{}, err
	}
	return response, nil
}

func (s *Service) ListKeywordSeeds(ctx context.Context, filter model.KeywordSeedFilter) (model.KeywordSeedListResponse, error) {
	if s.keywordSeeds == nil {
		return model.KeywordSeedListResponse{}, fmt.Errorf("keyword seed repository is not configured")
	}
	return s.keywordSeeds.ListKeywordSeeds(ctx, filter)
}

func (s *Service) GetKeywordSeed(ctx context.Context, id string) (model.KeywordSeed, bool, error) {
	if s.keywordSeeds == nil {
		return model.KeywordSeed{}, false, fmt.Errorf("keyword seed repository is not configured")
	}
	return s.keywordSeeds.GetKeywordSeed(ctx, id)
}

func (s *Service) CreateKeywordSeed(ctx context.Context, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	if err := validateSeed(req); err != nil {
		return model.KeywordSeed{}, err
	}
	if s.keywordSeeds == nil {
		return model.KeywordSeed{}, fmt.Errorf("keyword seed repository is not configured")
	}
	return s.keywordSeeds.CreateKeywordSeed(ctx, req)
}

func (s *Service) UpdateKeywordSeed(ctx context.Context, id string, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	if err := validateSeed(req); err != nil {
		return model.KeywordSeed{}, err
	}
	if s.keywordSeeds == nil {
		return model.KeywordSeed{}, fmt.Errorf("keyword seed repository is not configured")
	}
	return s.keywordSeeds.UpdateKeywordSeed(ctx, id, req)
}

func (s *Service) SetKeywordSeedEnabled(ctx context.Context, id string, enabled bool) (model.KeywordSeed, error) {
	if s.keywordSeeds == nil {
		return model.KeywordSeed{}, fmt.Errorf("keyword seed repository is not configured")
	}
	return s.keywordSeeds.SetKeywordSeedEnabled(ctx, id, enabled)
}

func (s *Service) ExportKeywordSeedsExcel(ctx context.Context) ([]byte, error) {
	list, err := s.ListKeywordSeeds(ctx, model.KeywordSeedFilter{Page: 1, PageSize: 10000})
	if err != nil {
		return nil, err
	}
	items := append([]model.KeywordSeed(nil), list.Items...)
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

func (s *Service) ImportKeywordSeeds(ctx context.Context, file multipart.File) (model.KeywordSeedImportResponse, error) {
	content, err := io.ReadAll(file)
	if err != nil {
		return model.KeywordSeedImportResponse{}, fmt.Errorf("read upload: %w", err)
	}
	return s.importWorkbookBytes(ctx, content)
}

func (s *Service) importWorkbookBytes(ctx context.Context, content []byte) (model.KeywordSeedImportResponse, error) {
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
		if _, err := s.CreateKeywordSeed(ctx, req); err != nil {
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

const (
	riskLevelNormal   = "normal"
	riskLevelElevated = "elevated"
	usageTTL          = 2 * time.Hour
)

func (s *Service) VerifyChallenge(ctx context.Context, meta RequestMeta, token string) (model.ChallengeVerifyResponse, error) {
	if !s.challengeVerifier.Available() {
		return model.ChallengeVerifyResponse{}, fmt.Errorf("challenge verifier is not configured")
	}
	if err := s.challengeVerifier.Verify(ctx, token, meta.ClientIP); err != nil {
		return model.ChallengeVerifyResponse{}, err
	}
	if err := s.markChallengePass(ctx, meta); err != nil {
		return model.ChallengeVerifyResponse{}, err
	}
	return model.ChallengeVerifyResponse{
		Verified:             true,
		PassExpiresInSeconds: int(s.challengePassTTL.Seconds()),
		RiskLevel:            riskLevelNormal,
	}, nil
}

type ErrRateLimited struct {
	CooldownSeconds int
}

func (e ErrRateLimited) Error() string {
	return "rate limited"
}

type ErrChallengeRequired struct {
	RiskLevel string
}

func (e ErrChallengeRequired) Error() string {
	return "challenge required"
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

func firstCooldown(states ...usageState) time.Time {
	for _, state := range states {
		if state.CooldownUntil.After(time.Now()) {
			return state.CooldownUntil
		}
	}
	return time.Time{}
}

func ensureUsageWindow(usage usageState, exists bool, now time.Time) usageState {
	if !exists || now.Sub(usage.WindowStarted) >= time.Hour {
		return usageState{WindowStarted: now}
	}
	return usage
}

func (s *Service) requestStatus(cacheHit bool, session usageState, cooldownUntil time.Time, challengePassed bool, riskLevel string) model.RequestStatus {
	return model.RequestStatus{
		CacheHit:            cacheHit,
		RemainingAIRequests: max(0, s.anonymousLimit-session.Used),
		CooldownSeconds:     cooldownSeconds(cooldownUntil),
		ChallengeRequired:   !challengePassed && riskLevel == riskLevelElevated && s.challengeVerifier.Available(),
		ChallengePassed:     challengePassed,
		RiskLevel:           riskLevel,
	}
}

func (s *Service) evaluateRiskLevel(meta RequestMeta) string {
	if strings.TrimSpace(meta.DeviceFingerprint) == "" {
		return riskLevelElevated
	}
	return riskLevelNormal
}

func (s *Service) loadUsage(ctx context.Context, scope, key string) (usageState, bool, error) {
	if strings.TrimSpace(key) == "" {
		return usageState{}, false, nil
	}
	return s.store.LoadUsage(ctx, scope, key)
}

func (s *Service) saveUsageTriplet(ctx context.Context, meta RequestMeta, session usageState, ip usageState, device usageState) error {
	if err := s.store.SaveUsage(ctx, "session", meta.AnonymousID, session, usageTTL); err != nil {
		return err
	}
	if strings.TrimSpace(meta.ClientIP) != "" {
		if err := s.store.SaveUsage(ctx, "ip", meta.ClientIP, ip, usageTTL); err != nil {
			return err
		}
	}
	if err := s.store.SaveUsage(ctx, "device", s.deviceKey(meta.DeviceFingerprint), device, usageTTL); err != nil {
		return err
	}
	return nil
}

func (s *Service) hasChallengePass(ctx context.Context, meta RequestMeta) (bool, error) {
	keys := []string{
		"session:" + meta.AnonymousID,
		"ip:" + strings.TrimSpace(meta.ClientIP),
		"device:" + s.deviceKey(meta.DeviceFingerprint),
	}
	for _, key := range keys {
		if strings.HasSuffix(key, ":") {
			continue
		}
		ok, err := s.store.HasChallengePass(ctx, key)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) markChallengePass(ctx context.Context, meta RequestMeta) error {
	keys := []string{
		"session:" + meta.AnonymousID,
		"ip:" + strings.TrimSpace(meta.ClientIP),
		"device:" + s.deviceKey(meta.DeviceFingerprint),
	}
	for _, key := range keys {
		if strings.HasSuffix(key, ":") {
			continue
		}
		if err := s.store.SetChallengePass(ctx, key, s.challengePassTTL); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) deviceKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "missing"
	}
	return value
}

func maxTime(values ...time.Time) time.Time {
	var current time.Time
	for _, value := range values {
		if value.After(current) {
			current = value
		}
	}
	return current
}
