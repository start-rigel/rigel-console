package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rigel-labs/rigel-console/internal/domain/model"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

type Repository struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string) (*Repository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *Repository) ListKeywordSeeds(ctx context.Context, filter model.KeywordSeedFilter) (model.KeywordSeedListResponse, error) {
	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	conditions := []string{"1 = 1"}
	args := make([]any, 0, 8)
	if category := normalizeCategory(filter.Category); category != "" {
		args = append(args, category)
		conditions = append(conditions, fmt.Sprintf("category = $%d::part_category", len(args)))
	}
	if filter.Brand != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filter.Brand))+"%")
		conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(brand, '')) LIKE $%d", len(args)))
	}
	if filter.Keyword != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filter.Keyword))+"%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(keyword) LIKE $%d OR LOWER(canonical_model) LIKE $%d OR LOWER(COALESCE(brand, '')) LIKE $%d OR EXISTS (SELECT 1 FROM jsonb_array_elements_text(aliases_json) alias WHERE LOWER(alias) LIKE $%d))", len(args), len(args), len(args), len(args)))
	}
	if filter.Enabled != nil {
		args = append(args, *filter.Enabled)
		conditions = append(conditions, fmt.Sprintf("enabled = $%d", len(args)))
	}

	whereClause := strings.Join(conditions, " AND ")
	countQuery := "SELECT COUNT(*) FROM rigel_keyword_seeds WHERE " + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return model.KeywordSeedListResponse{}, fmt.Errorf("count keyword seeds: %w", err)
	}

	args = append(args, pageSize, (page-1)*pageSize)
	query := `
SELECT id::text, category::text, keyword, canonical_model, COALESCE(brand, ''), aliases_json, priority, enabled, COALESCE(notes, ''), created_at, updated_at
FROM rigel_keyword_seeds
WHERE ` + whereClause + `
ORDER BY priority DESC, updated_at DESC
LIMIT $` + fmt.Sprintf("%d", len(args)-1) + ` OFFSET $` + fmt.Sprintf("%d", len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return model.KeywordSeedListResponse{}, fmt.Errorf("list keyword seeds: %w", err)
	}
	defer rows.Close()

	items := make([]model.KeywordSeed, 0, pageSize)
	for rows.Next() {
		item, err := scanKeywordSeed(rows)
		if err != nil {
			return model.KeywordSeedListResponse{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return model.KeywordSeedListResponse{}, fmt.Errorf("iterate keyword seeds: %w", err)
	}

	return model.KeywordSeedListResponse{Items: items, Page: page, PageSize: pageSize, Total: total}, nil
}

func (r *Repository) GetKeywordSeed(ctx context.Context, id string) (model.KeywordSeed, bool, error) {
	query := `
SELECT id::text, category::text, keyword, canonical_model, COALESCE(brand, ''), aliases_json, priority, enabled, COALESCE(notes, ''), created_at, updated_at
FROM rigel_keyword_seeds
WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, strings.TrimSpace(id))
	item, err := scanKeywordSeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.KeywordSeed{}, false, nil
		}
		return model.KeywordSeed{}, false, err
	}
	return item, true, nil
}

func (r *Repository) CreateKeywordSeed(ctx context.Context, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	aliases, err := json.Marshal(normalizeAliases(req.Aliases))
	if err != nil {
		return model.KeywordSeed{}, fmt.Errorf("marshal aliases: %w", err)
	}
	query := `
INSERT INTO rigel_keyword_seeds (category, keyword, canonical_model, brand, aliases_json, priority, enabled, notes)
VALUES ($1::part_category, $2, $3, $4, $5, $6, $7, $8)
RETURNING id::text, category::text, keyword, canonical_model, COALESCE(brand, ''), aliases_json, priority, enabled, COALESCE(notes, ''), created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query,
		normalizeCategory(req.Category),
		strings.TrimSpace(req.Keyword),
		strings.TrimSpace(req.CanonicalModel),
		nullableString(req.Brand),
		aliases,
		normalizePriority(req.Priority),
		req.Enabled,
		nullableString(req.Notes),
	)
	item, err := scanKeywordSeed(row)
	if err != nil {
		return model.KeywordSeed{}, translateConstraintError(err)
	}
	return item, nil
}

func (r *Repository) UpdateKeywordSeed(ctx context.Context, id string, req model.KeywordSeedUpsertRequest) (model.KeywordSeed, error) {
	aliases, err := json.Marshal(normalizeAliases(req.Aliases))
	if err != nil {
		return model.KeywordSeed{}, fmt.Errorf("marshal aliases: %w", err)
	}
	query := `
UPDATE rigel_keyword_seeds
SET category = $2::part_category,
    keyword = $3,
    canonical_model = $4,
    brand = $5,
    aliases_json = $6,
    priority = $7,
    enabled = $8,
    notes = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING id::text, category::text, keyword, canonical_model, COALESCE(brand, ''), aliases_json, priority, enabled, COALESCE(notes, ''), created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query,
		strings.TrimSpace(id),
		normalizeCategory(req.Category),
		strings.TrimSpace(req.Keyword),
		strings.TrimSpace(req.CanonicalModel),
		nullableString(req.Brand),
		aliases,
		normalizePriority(req.Priority),
		req.Enabled,
		nullableString(req.Notes),
	)
	item, err := scanKeywordSeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.KeywordSeed{}, consoleservice.ErrNotFound{Resource: "keyword seed"}
		}
		return model.KeywordSeed{}, translateConstraintError(err)
	}
	return item, nil
}

func (r *Repository) SetKeywordSeedEnabled(ctx context.Context, id string, enabled bool) (model.KeywordSeed, error) {
	query := `
UPDATE rigel_keyword_seeds
SET enabled = $2, updated_at = NOW()
WHERE id = $1
RETURNING id::text, category::text, keyword, canonical_model, COALESCE(brand, ''), aliases_json, priority, enabled, COALESCE(notes, ''), created_at, updated_at`
	row := r.db.QueryRowContext(ctx, query, strings.TrimSpace(id), enabled)
	item, err := scanKeywordSeed(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.KeywordSeed{}, consoleservice.ErrNotFound{Resource: "keyword seed"}
		}
		return model.KeywordSeed{}, fmt.Errorf("set keyword seed enabled: %w", err)
	}
	return item, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanKeywordSeed(scanner rowScanner) (model.KeywordSeed, error) {
	var item model.KeywordSeed
	var aliases []byte
	if err := scanner.Scan(
		&item.ID,
		&item.Category,
		&item.Keyword,
		&item.CanonicalModel,
		&item.Brand,
		&aliases,
		&item.Priority,
		&item.Enabled,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return model.KeywordSeed{}, err
	}
	if err := json.Unmarshal(aliases, &item.Aliases); err != nil {
		return model.KeywordSeed{}, fmt.Errorf("unmarshal aliases: %w", err)
	}
	return item, nil
}

func normalizeCategory(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "cpu":
		return "CPU"
	case "gpu":
		return "GPU"
	case "motherboard":
		return "MB"
	case "ram":
		return "RAM"
	case "ssd":
		return "SSD"
	case "psu":
		return "PSU"
	case "case":
		return "CASE"
	case "cooler":
		return "COOLER"
	default:
		return strings.ToUpper(strings.TrimSpace(value))
	}
}

func normalizeAliases(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
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

func normalizePriority(priority int) int {
	if priority <= 0 {
		return 100
	}
	return priority
}

func nullableString(value string) any {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return value
}

func translateConstraintError(err error) error {
	if strings.Contains(err.Error(), "uq_rigel_keyword_seeds_category_keyword") {
		return fmt.Errorf("keyword already exists in category")
	}
	return fmt.Errorf("persist keyword seed: %w", err)
}
