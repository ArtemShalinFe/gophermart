package models

import (
	"context"
	"errors"
	"time"
)

type UserDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID           string `json:"uuid"`
	Login        string `json:"login"`
	PasswordHash string `json:"password"`
}

type UserWithdrawalsHistory struct {
	ProcessedAt time.Time `json:"processed_at"`
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
}

var ErrLoginIsBusy = errors.New("login is busy")
var ErrUnknowUser = errors.New("unknow user")
var ErrNotEnoughAccruals = errors.New("not enough accruals")

type UserStorage interface {
	AddUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUploadedOrders(ctx context.Context, us *User) ([]*Order, error)
	GetCurrentBalance(ctx context.Context, userID string) (float64, error)
	GetWithdrawals(ctx context.Context, userID string) (float64, error)
	GetWithdrawalList(ctx context.Context, userID string) ([]*UserWithdrawalsHistory, error)
}

func (u *UserDTO) AddUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	return db.AddUser(ctx, u)
}

func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	return db.GetUser(ctx, u)
}

func (u *User) GetUploadedOrders(ctx context.Context, db UserStorage) ([]*Order, error) {
	return db.GetUploadedOrders(ctx, u)
}

func (u *User) GetUserBalance(ctx context.Context, db UserStorage) (float64, error) {
	return db.GetCurrentBalance(ctx, u.ID)
}

func (u *User) GetWithdrawals(ctx context.Context, db UserStorage) (float64, error) {
	return db.GetWithdrawals(ctx, u.ID)
}

func (u *User) GetWithdrawalList(ctx context.Context, db UserStorage) ([]*UserWithdrawalsHistory, error) {
	return db.GetWithdrawalList(ctx, u.ID)
}

func (u *User) AddWithdrawn(ctx context.Context, db OrderStorage, orderNumber string, sum float64) error {
	return db.AddWithdrawn(ctx, u.ID, orderNumber, sum)
}
