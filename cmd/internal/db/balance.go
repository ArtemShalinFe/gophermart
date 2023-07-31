package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/ArtemShalinFe/gophermart/cmd/internal/models"
)

func (db *DB) GetCurrentBalance(ctx context.Context, userID string) (float64, error) {
	sql := `
	SELECT sum
	FROM currentBalances
	WHERE userId = $1;`

	var b float64
	row := db.pool.QueryRow(ctx, sql, userID)
	if err := row.Scan(&b); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("db GetCurrentBalance err: %w", err)
		}
	}

	return b, nil
}

func (db *DB) AddWithdrawn(ctx context.Context, userID string, orderNumber string, sum float64) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to start AddWithdrawn transaction err: %w", err)
	}

	defer tx.Commit(ctx)

	sql := `
	INSERT INTO withdrawals(date, userid, orderNumber, sum)
	VALUES (CURRENT_TIMESTAMP, $1, $2, $3);`

	if _, err := tx.Exec(ctx, sql, userID, orderNumber, sum); err != nil {
		return fmt.Errorf("db AddWithdrawn err: %w", err)
	}

	if _, err := db.UpdateUserBalance(ctx, tx, userID, -sum); err != nil {
		if errors.Is(err, models.ErrNotEnoughAccruals) {
			if err := tx.Rollback(ctx); err != nil {
				return fmt.Errorf("failed rollback transaction AddWithdrawn err: %w", err)
			}
		}
		return fmt.Errorf("failed update user balance. AddWithdrawn err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed commit transaction AddWithdrawn err: %w", err)
	}

	return nil
}

func (db *DB) GetWithdrawals(ctx context.Context, userID string) (float64, error) {
	sql := `
	SELECT coalesce(sum(sum),0)
	FROM withdrawals
	WHERE userId = $1;`

	var b float64
	row := db.pool.QueryRow(ctx, sql, userID)
	if err := row.Scan(&b); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("db GetWithdrawals err: %w", err)
		}
	}

	return b, nil
}

func (db *DB) GetWithdrawalList(ctx context.Context, userID string) ([]*models.UserWithdrawalsHistory, error) {
	sql := `
	SELECT date, orderNumber, sum
	FROM withdrawals
	WHERE userId =  $1
	ORDER BY date DESC`

	var m []*models.UserWithdrawalsHistory

	rows, err := db.pool.Query(ctx, sql, userID)
	if err != nil {
		return nil, fmt.Errorf("db GetWithdrawalList err: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ub models.UserWithdrawalsHistory
		if err := rows.Scan(&ub.ProcessedAt, &ub.OrderNumber, &ub.Sum); err != nil {
			return nil, fmt.Errorf("db rows scan err GetWithdrawalList err: %w", err)
		}

		m = append(m, &ub)
	}

	return m, nil
}

func (db *DB) UpdateUserBalance(ctx context.Context, tx pgx.Tx, userID string, sum float64) (float64, error) {
	sql := `
	INSERT INTO currentBalances(userid, sum)
		VALUES ($1, $2)
	ON CONFLICT (userid) 
		DO UPDATE SET sum = EXCLUDED.sum + currentBalances.sum
	RETURNING
		sum;`

	var cb float64
	row := tx.QueryRow(ctx, sql, userID, sum)
	if err := row.Scan(&cb); err != nil {
		return 0, fmt.Errorf("db UpdateUserBalance err: %w", err)
	}

	if cb < 0 {
		return 0, models.ErrNotEnoughAccruals
	}

	return cb, nil
}
