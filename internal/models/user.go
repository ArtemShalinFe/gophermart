package models

import (
	"context"
	"errors"
	"fmt"
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

type UserBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

var ErrLoginIsBusy = errors.New("login is busy")
var ErrUnknowUser = errors.New("unknow user")
var ErrNotEnoughAccruals = errors.New("not enough accruals")

type UserStorage interface {
	AddUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUploadedOrders(ctx context.Context, us *User) ([]*Order, error)
	GetBalance(ctx context.Context, userID string) (*UserBalance, error)
	GetWithdrawalList(ctx context.Context, userID string) ([]*UserWithdrawalsHistory, error)
}

func (u *UserDTO) AddUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.AddUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("add user was failed err: %w", err)
	}
	return us, nil
}

func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {
	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	us, err := db.GetUser(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("get user was failed err: %w", err)
	}
	return us, nil
}

func (u *User) GetUploadedOrders(ctx context.Context, db UserStorage) ([]*Order, error) {
	ors, err := db.GetUploadedOrders(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("get uploaded orders was failed err: %w", err)
	}
	return ors, nil
}

func (u *User) GetBalance(ctx context.Context, db UserStorage) (*UserBalance, error) {
	b, err := db.GetBalance(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("get balance was failed err: %w", err)
	}
	return b, nil
}

func (u *User) GetWithdrawalList(ctx context.Context, db UserStorage) ([]*UserWithdrawalsHistory, error) {
	lws, err := db.GetWithdrawalList(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("get withdrawal list was failed err: %w", err)
	}
	return lws, nil
}

func (u *User) AddWithdrawn(ctx context.Context, db OrderStorage, orderNumber string, sum float64) error {
	if err := db.AddWithdrawn(ctx, u.ID, orderNumber, sum); err != nil {
		return fmt.Errorf("add withdrawn was failed err: %w", err)
	}
	return nil
}
