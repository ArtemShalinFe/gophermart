package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/models"
)

func (db *DB) AddOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start transaction err: %w", err)
	}
	defer tx.Commit(ctx)

	o, err := addOrder(ctx, tx, order)
	if err != nil {
		return nil, fmt.Errorf("db AddOrder err: %w", err)
	}

	return o, nil

}

func (db *DB) GetOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start transaction err: %w", err)
	}
	defer tx.Commit(ctx)

	return getOrderByNumber(ctx, tx, order.Number)

}

func getOrderByNumber(ctx context.Context, tx pgx.Tx, number int64) (*models.Order, error) {

	sql := `
	SELECT 
		id, uploaded, number, sum, userid, status
	FROM 
		orders
	WHERE 
		number = $1;`

	row := tx.QueryRow(ctx, sql, number)

	o := models.Order{}
	if err := row.Scan(&o.Id, &o.Uploaded_at, &o.Number, &o.Accrual, &o.UserId, &o.Status); err != nil {
		return nil, err
	}

	return &o, nil

}

func addOrder(ctx context.Context, tx pgx.Tx, order *models.OrderDTO) (*models.Order, error) {

	sql := `
	INSERT INTO orders(uploaded, number, userid, status, sum)
	VALUES (CURRENT_TIMESTAMP, $1, $2, $3, 0)
	RETURNING 
		id, uploaded, number, sum, userid, status
	;`

	row := tx.QueryRow(ctx, sql, order.Number, order.UserId, models.OrderStatusNew)

	o := models.Order{}
	if err := row.Scan(&o.Id, &o.Uploaded_at, &o.Number, &o.Accrual, &o.UserId, &o.Status); err != nil {

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return nil, models.ErrOrderWasRegisteredEarlier
			}
			return nil, fmt.Errorf("db addOrder err: %w", err)
		}

		return nil, fmt.Errorf("db addOrder err: %w", err)

	}

	return &o, nil

}
