package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/models"
	"go.uber.org/zap"
)

type Storage interface {
	AddUser(ctx context.Context, us *models.UserDTO) (*models.User, error)
	GetUser(ctx context.Context, us *models.UserDTO) (*models.User, error)
	GetUploadedOrders(ctx context.Context, order *models.User) ([]*models.Order, error)
	AddOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
	GetOrder(ctx context.Context, order *models.OrderDTO) (*models.Order, error)
}

type Handlers struct {
	log       *zap.Logger
	store     Storage
	secretKey []byte
}

func NewHandlers(secretKey []byte, db Storage, log *zap.Logger) (*Handlers, error) {

	return &Handlers{
		store:     db,
		secretKey: secretKey,
		log:       log,
	}, nil

}

func (h *Handlers) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	u, err := getLoginPsw(w, r)
	if err != nil {
		h.log.Error("failed to read the Register request body: ", zap.Error(err))
		return
	}

	u.Password = models.EncodePassword(u.Password)

	if _, err = u.AddUser(ctx, h.store); err != nil {

		if errors.Is(err, models.ErrLoginIsBusy) {
			w.WriteHeader(http.StatusConflict)
			return
		} else {
			h.log.Error("failed in the Register request: ", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	}

	token, err := NewJWTToken(h.secretKey, u.Login, u.Password)
	if err != nil {
		h.log.Error("failed to build JWT token in the Register request: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)

	w.WriteHeader(http.StatusOK)

}

func (h *Handlers) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	u, err := getLoginPsw(w, r)
	if err != nil {
		h.log.Error("failed to read the Register request body: ", zap.Error(err))
		return
	}

	user, err := u.GetUser(ctx, h.store)
	if err != nil {
		if errors.Is(err, models.ErrUnknowUser) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.log.Error("failed in the Register request: ", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	u.Password = models.EncodePassword(u.Password)
	if user.PasswordBase64 != u.Password {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := NewJWTToken(h.secretKey, u.Login, u.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Authorization", token)

	w.WriteHeader(http.StatusOK)

}

func (h *Handlers) AddOrder(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	u, err := h.GetUserFromJWTToken(w, r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error("failed to get user from JWT in the AddOrder request: ", zap.Error(err))
		return
	}

	b, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.log.Error("failed to read the AddOrder request body: ", zap.Error(err))
		return
	}

	number, err := strconv.ParseInt(string(b), 0, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.log.Error("failed to convert in the AddOrder request: ", zap.Error(err))
		return
	}

	o := &models.OrderDTO{
		Number: number,
		UserId: u.Id,
	}

	if _, err = o.AddOrder(ctx, h.store); err != nil {

		if !errors.Is(err, models.ErrOrderWasRegisteredEarlier) {
			w.WriteHeader(http.StatusBadRequest)
			h.log.Error("failed to add order in the AddOrder request: ", zap.Error(err))
			return
		}

		o, err := o.GetOrder(ctx, h.store)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.Error("failed to get the order the AddOrder request body: ", zap.Error(err))
			return
		}

		if o.UserId != u.Id {
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
	h.log.Info("GetOrders not implemented")
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
		return nil, err
	}

	if err := json.Unmarshal(b, &u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}

	if u.Login == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("bad request")
	}

	return &u, nil

}
