package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ArtemShalinFe/gophermart/cmd/internal/models"
)

func (db *DB) AddOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error) {
	sql := `
	INSERT INTO orders(uploaded, number, userid, status, sum)
	VALUES (CURRENT_TIMESTAMP, $1, $2, $3, 0)
	RETURNING 
		id, uploaded, number, sum, userid, status;`

	row := db.pool.QueryRow(ctx, sql, order.Number, order.UserID, models.OrderStatusNew)

	o := models.Order{}
	if err := row.Scan(&o.ID, &o.UploadedAt, &o.Number, &o.Accrual, &o.UserID, &o.Status); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) && pgErr.ConstraintName == "orders_number_key" {
				return nil, models.ErrOrderWasRegisteredEarlier
			}
			return nil, fmt.Errorf("db AddOrder pgerr: %w", err)
		}
		return nil, fmt.Errorf("db AddOrder row scan err: %w", err)
	}

	return &o, nil
}

func (db *DB) GetOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error) {
	sql := `
	SELECT 
		id, uploaded, number, sum, userid, status
	FROM 
		orders
	WHERE 
		number = $1;`

	row := db.pool.QueryRow(ctx, sql, order.Number)

	o := models.Order{}
	if err := row.Scan(&o.ID, &o.UploadedAt, &o.Number, &o.Accrual, &o.UserID, &o.Status); err != nil {
		return nil, fmt.Errorf("db GetOrder err: %w", err)
	}

	return &o, nil
}

func (db *DB) GetOrdersForAccrual(ctx context.Context) ([]*models.Order, error) {
	sql := `
	SELECT id, userid, uploaded, number, sum, status
	FROM orders
	WHERE status IN ($1, $2)
	ORDER BY uploaded DESC
	LIMIT 10;`

	rows, err := db.pool.Query(ctx, sql, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("db GetOrdersForAccrual err: %w", err)
	}
	defer rows.Close()

	var ors []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.UploadedAt, &o.Number, &o.Accrual, &o.Status); err != nil {
			return nil, fmt.Errorf("db GetOrdersForAccrual row scan err: %w", err)
		}
		ors = append(ors, &o)
	}

	return ors, nil
}

func (db *DB) UpdateOrder(ctx context.Context, order *models.Order) error {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("unable to start UpdateOrder transaction err: %w", err)
	}

	defer func(tx pgx.Tx) {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				db.log.Errorf("failed rollback transaction UpdateOrder err: %w", err)
			}
		}
	}(tx)

	sql := `
	UPDATE orders
	SET
		uploaded = $2, 
		number = $3, 
		userid = $4,
		sum = $5, 
		status = $6
	WHERE
		id = $1;`

	if _, err := tx.Exec(ctx, sql,
		order.ID, order.UploadedAt, order.Number, order.UserID, order.Accrual, order.Status); err != nil {
		return fmt.Errorf("db UpdateOrder err: %w", err)
	}

	if _, err := db.UpdateUserBalance(ctx, tx, order.UserID, order.Accrual); err != nil {
		return fmt.Errorf("failed update user balance. UpdateUserBalance err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed commit transaction UpdateOrder err: %w", err)
	}

	return nil
}

func (db *DB) GetUploadedOrders(ctx context.Context, u *models.User) ([]*models.Order, error) {
	sql := `
	SELECT id, uploaded, number, sum, status
	FROM orders
	WHERE userId = $1
	ORDER BY uploaded DESC;`

	rows, err := db.pool.Query(ctx, sql, u.ID)
	if err != nil {
		return nil, fmt.Errorf("db GetUploadedOrders err: %w", err)
	}
	defer rows.Close()

	var ors []*models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UploadedAt, &o.Number, &o.Accrual, &o.Status); err != nil {
			return nil, fmt.Errorf("db GetUploadedOrders row scan err: %w", err)
		}
		ors = append(ors, &o)
	}

	return ors, nil
}
