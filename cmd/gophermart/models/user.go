package models

import (
	"context"
	"encoding/base64"
	"errors"
)

type UserDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	Id             string `json:"uuid"`
	Login          string `json:"login"`
	PasswordBase64 string `json:"password"`
}

var ErrLoginIsBusy = errors.New("login is busy")
var ErrUnknowUser = errors.New("unknow user")

type UserStorage interface {
	AddUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUser(ctx context.Context, us *UserDTO) (*User, error)
	GetUploadedOrders(ctx context.Context, us *User) ([]*Order, error)
}

func (u *UserDTO) AddUser(ctx context.Context, db UserStorage) (*User, error) {
	return db.AddUser(ctx, u)
}

func (u *UserDTO) GetUser(ctx context.Context, db UserStorage) (*User, error) {

	if u.Login == "" {
		return nil, ErrUnknowUser
	}

	return db.GetUser(ctx, u)
}

func (o *User) GetUploadedOrders(ctx context.Context, db UserStorage) ([]*Order, error) {
	return db.GetUploadedOrders(ctx, o)
}

func EncodePassword(password string) string {
	return base64.StdEncoding.EncodeToString([]byte(password))
}
