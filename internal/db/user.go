package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/ArtemShalinFe/gophermart/internal/models"
)

func (db *DB) AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	sql := `
	INSERT INTO users(login, pass)
	VALUES ($1, $2)
	RETURNING id, login, pass
	;`

	row := db.pool.QueryRow(ctx, sql, us.Login, us.Password)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) && pgErr.ConstraintName == "users_login_key" {
				return nil, models.ErrLoginIsBusy
			}
			return nil, fmt.Errorf("db AddUser pgerr: %w", err)
		}
		return nil, fmt.Errorf("db AddUser row scan err: %w", err)
	}

	return &u, nil
}

func (db *DB) GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error) {
	sql := `
	SELECT id, login, pass
	FROM users
	WHERE login = $1;`

	row := db.pool.QueryRow(ctx, sql, us.Login)

	u := models.User{}
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUnknowUser
		}
		return nil, fmt.Errorf("db GetUser row scan err: %w", err)
	}

	return &u, nil
}
