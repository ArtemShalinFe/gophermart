package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/models"
	"go.uber.org/zap"
)

type Storage interface {
	AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error)
	GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error)
	GetUploadedOrders(ctx context.Context, order *models.User) ([]*models.Order, error)
	GetOrdersForAccrual(ctx context.Context) ([]*models.Order, error)
	AddOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
	GetOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
}

const authHeaderName = "Authorization"

type Handlers struct {
	log       *zap.SugaredLogger
	store     Storage
	secretKey []byte
}

func NewHandlers(secretKey []byte, db Storage, log *zap.SugaredLogger) (*Handlers, error) {
	return &Handlers{
		store:     db,
		secretKey: secretKey,
		log:       log,
	}, nil
}

func (h *Handlers) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := getLoginPsw(w, r)
	if err != nil {
		h.log.Errorf("failed to read the Register request body err: %w ", err)
		return
	}

	if _, err = u.AddUser(ctx, h.store); err != nil {
		if errors.Is(err, models.ErrLoginIsBusy) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		h.log.Errorf("failed add user in the Register request err: %w ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := NewJWTToken(h.secretKey, u.Login, u.Password)
	if err != nil {
		h.log.Errorf("failed to build JWT token in the Register request err: %w ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(authHeaderName, token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := getLoginPsw(w, r)
	if err != nil {
		h.log.Errorf("failed to read the Login request body err: %w ", err)
		return
	}

	_, err = u.GetUser(ctx, h.store)
	if err != nil {
		if errors.Is(err, models.ErrUnknowUser) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.log.Errorf("failed in the Register request err: %w ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	token, err := NewJWTToken(h.secretKey, u.Login, u.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(authHeaderName, token)
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) AddOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user from JWT in the AddOrder request err: %w ", err)
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("failed to read the AddOrder request body err: %w ", err)
		return
	}

	o := &models.OrderDTO{
		Number: string(b),
		UserID: u.ID,
	}

	if !o.NumberIsCorrect() {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if _, err = o.AddOrder(ctx, h.store); err != nil {
		if !errors.Is(err, models.ErrOrderWasRegisteredEarlier) {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Errorf("failed to add order in the AddOrder request err: %w ", err)
			return
		}

		o, err := o.GetOrder(ctx, h.store)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Errorf("failed to get the order the AddOrder request body err: %w ", err)
			return
		}

		if o.UserID != u.ID {
			w.WriteHeader(http.StatusConflict)
			return
		} else {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handlers) GetOrders(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user from JWT in the AddOrder request err: %w ", err)
		return
	}

	os, err := u.GetUploadedOrders(ctx, h.store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("get uploaded orders err: %w", err)
		return
	}

	b, err := json.Marshal(os)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetOrders marshal to json err: %w", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if _, err = w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetMetric error: %w", err)
		return
	}
}

func (h *Handlers) GetBalance(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.log.Info("GetBalance not implemented")
}

func (h *Handlers) AddBalanceWithdraw(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.log.Info("AddBalanceWithdraw not implemented")
}

func (h *Handlers) GetWithdrawals(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.log.Info("GetWithdrawals not implemented")
}

func getLoginPsw(w http.ResponseWriter, r *http.Request) (*models.UserDTO, error) {
	var u models.UserDTO

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("failed get login and pass from body err: %w", err)
	}

	if err := json.Unmarshal(b, &u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("failed unmarhsal login and pass err: %w", err)
	}

	if u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("bad request")
	}

	u.Password = models.EncodePassword(u.Password)

	return &u, nil
}
