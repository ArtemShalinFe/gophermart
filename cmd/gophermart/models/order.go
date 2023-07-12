package models

import (
	"context"
	"errors"
	"time"
)

type OrderDTO struct {
	Number int64  `json:"number"`
	UserId string `json:"userId"`
}

// Заказ загружен в систему, но не попал в обработку;
const OrderStatusNew = "NEW"

// Вознаграждение за заказ рассчитывается;
const OrderStatusProcessing = "PROCESSING"

// Система расчёта вознаграждений отказала в расчёте;
const OrderStatusInvalid = "INVALID"

// Данные по заказу проверены и информация о расчёте успешно получена.
const OrderStatusProcessed = "PROCESSED"

type Order struct {
	Id          string    `json:"uuid"`
	UserId      string    `json:"userId"`
	Number      int64     `json:"number"`
	Status      string    `json:"status"`
	Accrual     int64     `json:"accrual"`
	Uploaded_at time.Time `json:"uploaded_at"`
}

type OrderStorage interface {
	AddOrder(ctx context.Context, order *OrderDTO) (*Order, error)
	GetOrder(ctx context.Context, order *OrderDTO) (*Order, error)
}

var ErrOrderWasRegisteredEarlier = errors.New("the order was registered earlier")

func (o *OrderDTO) AddOrder(ctx context.Context, db OrderStorage) (*Order, error) {
	return db.AddOrder(ctx, o)
}

func (o *OrderDTO) GetOrder(ctx context.Context, db OrderStorage) (*Order, error) {
	return db.GetOrder(ctx, o)
}
