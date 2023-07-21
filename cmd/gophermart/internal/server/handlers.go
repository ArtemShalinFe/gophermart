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
	GetCurrentBalance(ctx context.Context, userID string) (float64, error)
	GetWithdrawals(ctx context.Context, userID string) (float64, error)
	GetWithdrawalList(ctx context.Context, userID string) ([]*models.UserWithdrawalsHistory, error)
	AddWithdrawn(ctx context.Context, userID string, orderNumber string, sum float64) error
	GetUploadedOrders(ctx context.Context, order *models.User) ([]*models.Order, error)
	GetOrdersForAccrual(ctx context.Context) ([]*models.Order, error)
	AddOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
	GetOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
	UpdateOrder(ctx context.Context, order *models.Order) error
}

const authHeaderName = "Authorization"
const ContentTypeJSON = "application/json"

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
		h.log.Errorf("failed to get user from JWT in the GetOrders request err: %w ", err)
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

	w.Header().Set("Content-Type", ContentTypeJSON)

	if _, err = w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetMetric error: %w", err)
		return
	}
}

func (h *Handlers) GetBalance(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user from JWT in the GetBalance request err: %w ", err)
		return
	}

	c, err := u.GetUserBalance(ctx, h.store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("failed to get user current balance in the GetBalance request err: %w ", err)
		return
	}

	wn, err := u.GetWithdrawals(ctx, h.store)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("failed to get user withdrawals in the GetBalance request err: %w ", err)
		return
	}

	resp := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   c,
		Withdrawn: wn,
	}

	b, err := json.Marshal(&resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetBalance marshal to json err: %w", err)
		return
	}

	w.Header().Set("Content-Type", ContentTypeJSON)

	if _, err = w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetBalance error: %w", err)
		return
	}
}

func (h *Handlers) AddBalanceWithdrawn(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user from JWT in the AddBalanceWithdrawn request err: %w ", err)
		return
	}

	req := struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}{}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("failed get order and accrual from body err: %w", err)
		return
	}

	if err := json.Unmarshal(b, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed unmarhsal order and accrual err: %w", err)
		return
	}

	if err := u.AddWithdrawn(ctx, h.store, req.Order, req.Sum); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to add withdrawal err: %w", err)
		return
	}
}

func (h *Handlers) GetBalanceMovementHistory(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user from JWT in the GetBalanceMovementHistory request err: %w ", err)
		return
	}

	ub, err := u.GetWithdrawalList(ctx, h.store)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Errorf("failed to get user balance history in the GetBalanceMovementHistory request err: %w ", err)
		return
	}

	if len(ub) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	b, err := json.Marshal(&ub)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetBalanceMovementHistory marshal to json err: %w", err)
		return
	}

	w.Header().Set("Content-Type", ContentTypeJSON)

	if _, err = w.Write(b); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Errorf("GetBalanceMovementHistory error: %w", err)
		return
	}
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
