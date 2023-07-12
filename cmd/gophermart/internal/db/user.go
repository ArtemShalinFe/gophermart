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

func (db *DB) AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	sql := `
	INSERT INTO public.users(login, pass)
	VALUES ($1, $2)
	RETURNING id, login, pass
	;`

	row := tx.QueryRow(ctx, sql, us.Login, us.Password)

	u := models.User{}
	if err := row.Scan(&u.Id, &u.Login, &u.PasswordBase64); err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return nil, models.ErrLoginIsBusy
			}
			return nil, fmt.Errorf("db AddUser err: %w", err)
		}

		return nil, fmt.Errorf("db AddUser err: %w", err)

	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction err: %w", err)
	}

	return &u, nil

}

func (db *DB) GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	sql := `
	SELECT id, login, pass
	FROM users
	WHERE login = $1 AND pass = $2;`

	row := tx.QueryRow(ctx, sql, us.Login, us.Password)

	u := models.User{}
	if err := row.Scan(&u.Id, &u.Login, &u.PasswordBase64); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUnknowUser
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("unable to commit transaction err: %w", err)
	}

	return &u, nil

}

func (db *DB) GetUploadedOrders(ctx context.Context, u *models.User) ([]*models.Order, error) {

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start transaction err: %w", err)
	}
	defer tx.Rollback(ctx)

	sql := `
	SELECT id, uploaded, number, sum, status,
	FROM orders
	WHERE userId = $1
	ORDER BY uploaded DESC;`

	rows, err := tx.Query(ctx, sql, u.Id)
	if err != nil {
		return nil, fmt.Errorf("db GetUploadedOrders err: %w", err)
	}

	var ors []*models.Order
	for rows.Next() {

		var o models.Order
		rows.Scan(&o.Id, &o.Uploaded_at, &o.Number, &o.Accrual, &o.Status)
		ors = append(ors, &o)

	}

	return ors, nil

}
