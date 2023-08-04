package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/ArtemShalinFe/gophermart/internal/models"
)

type Claims struct {
	jwt.RegisteredClaims
	Login string
}

func NewJWTToken(secretKey []byte, login string, tokenExp time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		Login: login,
	})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", fmt.Errorf("get token signed string err: %w", err)
	}

	return tokenString, nil
}

func isAuthorized(tokenString string, secretKey []byte) (bool, error) {
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

func (h *Handlers) getUserFromJWTToken(r *http.Request) (*models.User, error) {
	authToken := r.Header.Get(authHeaderName)

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(authToken, claims, func(t *jwt.Token) (interface{}, error) {
		return h.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("GetUserFromJWTToken error: %w", err)
	}

	udto := models.UserDTO{
		Login: claims.Login,
	}

	u, err := udto.GetUser(r.Context(), h.store)
	if err != nil {
		return nil, fmt.Errorf("get user from jwt token was failed err: %w", err)
	}
	return u, nil
}

type key int

var userKey key

func userFromContext(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(userKey).(*models.User)
	return u, ok
}

func (h *Handlers) JwtMiddleware(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get(authHeaderName)

		authorized, err := isAuthorized(authToken, h.secretKey)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !authorized {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		u, err := h.getUserFromJWTToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			h.log.Infof("failed to get user from JWT in JwtMiddleware err: %w ", err)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, u)
		hr.ServeHTTP(w, r.WithContext(ctx))
	})
}
