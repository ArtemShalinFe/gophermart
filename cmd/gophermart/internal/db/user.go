package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/models"
)

func (db *DB) AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start AddUser transaction err: %w", err)
	}

	defer tx.Rollback(ctx)

	sql := `
	INSERT INTO public.users(login, pass)
	VALUES ($1, $2)
	RETURNING id, login, pass
	;`

	row := tx.QueryRow(ctx, sql, us.Login, us.Password)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordBase64); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
				return nil, models.ErrLoginIsBusy
			}
			return nil, fmt.Errorf("db AddUser pgerr: %w", err)
		}
		return nil, fmt.Errorf("db AddUser row scan err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed commit transaction AddUser err: %w", err)
	}

	return &u, nil
}

func (db *DB) GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to start GetUser transaction err: %w", err)
	}

	defer tx.Rollback(ctx)

	sql := `
	SELECT id, login, pass
	FROM users
	WHERE login = $1 AND pass = $2;`

	row := tx.QueryRow(ctx, sql, us.Login, us.Password)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordBase64); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUnknowUser
		}
		return nil, fmt.Errorf("db GetUser row scan err: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed commit transaction GetUser err: %w", err)
	}

	return &u, nil
}
