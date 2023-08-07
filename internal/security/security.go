package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type hash struct{}

func NewHashController() (*hash, error) {
	return &hash{}, nil
}

func (h *hash) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("get password hash err: %w", err)
	}
	return string(bytes), nil
}

func (h *hash) CheckPasswordHash(hash string, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
