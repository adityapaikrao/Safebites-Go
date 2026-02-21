package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/safebites/backend-go/internal/model"
)

type favoriteQuerier interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
}

type favoriteRepo struct {
	q favoriteQuerier
}

func NewFavoriteRepository(db *DB) FavoriteRepository {
	return &favoriteRepo{q: db.Pool}
}

func (r *favoriteRepo) ListByUser(ctx context.Context, userID string) ([]model.Favorite, error) {
	const query = `
		SELECT id, user_id, product_name, COALESCE(brand, ''), safety_score, COALESCE(image, ''), added_at
		FROM favorites
		WHERE user_id = $1
		ORDER BY added_at DESC`

	rows, err := r.q.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list favorites by user: %w", err)
	}
	defer rows.Close()

	favorites := make([]model.Favorite, 0)
	for rows.Next() {
		var favorite model.Favorite
		var safetyScore sql.NullInt32
		if err := rows.Scan(
			&favorite.ID,
			&favorite.UserID,
			&favorite.ProductName,
			&favorite.Brand,
			&safetyScore,
			&favorite.Image,
			&favorite.AddedAt,
		); err != nil {
			return nil, fmt.Errorf("scan favorite row: %w", err)
		}
		if safetyScore.Valid {
			value := int(safetyScore.Int32)
			favorite.SafetyScore = &value
		}
		favorites = append(favorites, favorite)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate favorites: %w", err)
	}

	return favorites, nil
}

func (r *favoriteRepo) Create(ctx context.Context, favorite *model.Favorite) (*model.Favorite, error) {
	const query = `
		INSERT INTO favorites (user_id, product_name, brand, safety_score, image)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, product_name)
		DO UPDATE SET
			brand = EXCLUDED.brand,
			safety_score = EXCLUDED.safety_score,
			image = EXCLUDED.image
		RETURNING id, user_id, product_name, COALESCE(brand, ''), safety_score, COALESCE(image, ''), added_at`

	var created model.Favorite
	var safetyScore sql.NullInt32
	err := r.q.QueryRow(
		ctx,
		query,
		favorite.UserID,
		favorite.ProductName,
		favorite.Brand,
		favorite.SafetyScore,
		favorite.Image,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.ProductName,
		&created.Brand,
		&safetyScore,
		&created.Image,
		&created.AddedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create favorite: %w", err)
	}
	if safetyScore.Valid {
		value := int(safetyScore.Int32)
		created.SafetyScore = &value
	}

	return &created, nil
}

func (r *favoriteRepo) Delete(ctx context.Context, userID string, favoriteID int) error {
	const query = `DELETE FROM favorites WHERE user_id = $1 AND id = $2`

	cmdTag, err := r.q.Exec(ctx, query, userID, favoriteID)
	if err != nil {
		return fmt.Errorf("delete favorite: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *favoriteRepo) Exists(ctx context.Context, userID, productName string) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = $1 AND product_name = $2)`

	var exists bool
	err := r.q.QueryRow(ctx, query, userID, productName).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check favorite exists: %w", err)
	}

	return exists, nil
}
