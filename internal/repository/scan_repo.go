package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/safebites/backend-go/internal/model"
)

type scanQuerier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

type scanRepo struct {
	q scanQuerier
}

func NewScanRepository(db *DB) ScanRepository {
	return &scanRepo{q: db.Pool}
}

func (r *scanRepo) ListByUser(ctx context.Context, userID string, limit int) ([]model.Scan, error) {
	if limit <= 0 {
		limit = 20
	}

	const query = `
		SELECT id, user_id, product_name, COALESCE(brand, ''), COALESCE(image, ''), safety_score, is_safe, ingredients, timestamp
		FROM scans
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2`

	rows, err := r.q.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list scans by user: %w", err)
	}
	defer rows.Close()

	scans := make([]model.Scan, 0)
	for rows.Next() {
		var scan model.Scan
		var ingredientsBytes []byte

		if err := rows.Scan(
			&scan.ID,
			&scan.UserID,
			&scan.ProductName,
			&scan.Brand,
			&scan.Image,
			&scan.SafetyScore,
			&scan.IsSafe,
			&ingredientsBytes,
			&scan.Timestamp,
		); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}

		if err := unmarshalIngredients(ingredientsBytes, &scan.Ingredients); err != nil {
			return nil, fmt.Errorf("decode ingredients: %w", err)
		}

		scans = append(scans, scan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate scans: %w", err)
	}

	return scans, nil
}

func (r *scanRepo) Create(ctx context.Context, scan *model.Scan) (*model.Scan, error) {
	ingredientsJSON, err := json.Marshal(scan.Ingredients)
	if err != nil {
		return nil, fmt.Errorf("marshal ingredients: %w", err)
	}

	const query = `
		INSERT INTO scans (id, user_id, product_name, brand, image, safety_score, is_safe, ingredients)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
		RETURNING id, user_id, product_name, COALESCE(brand, ''), COALESCE(image, ''), safety_score, is_safe, ingredients, timestamp`

	var created model.Scan
	var ingredientsBytes []byte

	err = r.q.QueryRow(
		ctx,
		query,
		scan.ID,
		scan.UserID,
		scan.ProductName,
		scan.Brand,
		scan.Image,
		scan.SafetyScore,
		scan.IsSafe,
		ingredientsJSON,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.ProductName,
		&created.Brand,
		&created.Image,
		&created.SafetyScore,
		&created.IsSafe,
		&ingredientsBytes,
		&created.Timestamp,
	)
	if err != nil {
		return nil, fmt.Errorf("create scan: %w", err)
	}

	if err := unmarshalIngredients(ingredientsBytes, &created.Ingredients); err != nil {
		return nil, fmt.Errorf("decode ingredients: %w", err)
	}

	return &created, nil
}

func (r *scanRepo) GetStats(ctx context.Context, userID string) (*model.UserStats, error) {
	const query = `
		SELECT
			COUNT(*) AS total_scans,
			COUNT(*) FILTER (WHERE timestamp::date = CURRENT_DATE) AS today_scans,
			COUNT(*) FILTER (WHERE timestamp::date = CURRENT_DATE AND is_safe = TRUE) AS safe_today,
			COUNT(*) FILTER (WHERE timestamp::date = CURRENT_DATE AND is_safe = FALSE) AS risky_today,
			COALESCE(AVG(safety_score)::float8, 0) AS average_score
		FROM scans
		WHERE user_id = $1`

	var stats model.UserStats
	err := r.q.QueryRow(ctx, query, userID).Scan(
		&stats.TotalScans,
		&stats.TodayScans,
		&stats.SafeToday,
		&stats.RiskyToday,
		&stats.AverageScore,
	)
	if err != nil {
		return nil, fmt.Errorf("get user scan stats: %w", err)
	}

	return &stats, nil
}

func unmarshalIngredients(in []byte, out *[]map[string]interface{}) error {
	if len(in) == 0 {
		*out = []map[string]interface{}{}
		return nil
	}
	if err := json.Unmarshal(in, out); err != nil {
		return err
	}
	if *out == nil {
		*out = []map[string]interface{}{}
	}
	return nil
}
