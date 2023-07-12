package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/models"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Login    string
	Password string
}

func NewJWTToken(secretKey []byte, login string, password string) (string, error) {

	const TOKEN_EXP = time.Hour * 1

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
		},
		Login:    login,
		Password: password,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func IsAuthorized(tokenString string, secretKey []byte) (bool, error) {

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return secretKey, nil
		})

	if err != nil {
		return false, fmt.Errorf("token parse err: %w", err)
	}

	if !token.Valid {
		return false, fmt.Errorf("token is invalid: %w", err)
	}

	return true, nil

}

func (h *Handlers) GetUserFromJWTToken(w http.ResponseWriter, r *http.Request) (*models.User, error) {

	authToken := r.Header.Get("Authorization")

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(authToken, claims, func(t *jwt.Token) (interface{}, error) {
		return h.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("GetUserFromJWTToken error: %w", err)
	}

	udto := models.UserDTO{
		Login:    claims.Login,
		Password: claims.Password,
	}

	return udto.GetUser(r.Context(), h.store)

}

func (h *Handlers) JwtMiddleware(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authToken := r.Header.Get("Authorization")

		authorized, err := IsAuthorized(authToken, h.secretKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if authorized {
			hr.ServeHTTP(w, r)
		}

	})
}
