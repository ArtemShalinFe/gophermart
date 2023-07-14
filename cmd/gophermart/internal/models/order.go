package models

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type OrderDTO struct {
	UserID string `json:"userId"`
	Number string `json:"number"`
}

const OrderStatusNew = "NEW"
const OrderStatusProcessing = "PROCESSING"
const OrderStatusInvalid = "INVALID"
const OrderStatusProcessed = "PROCESSED"

type Order struct {
	UploadedAt time.Time `json:"uploaded_at"`
	ID         string    `json:"uuid"`
	UserID     string    `json:"userId,omitempty"`
	Status     string    `json:"status"`
	Number     string    `json:"number"`
	Accrual    int64     `json:"accrual"`
}

type OrderStorage interface {
	AddOrder(ctx context.Context, order *OrderDTO) (*Order, error)
	GetOrder(ctx context.Context, order *OrderDTO) (*Order, error)
	GetOrdersForAccrual(ctx context.Context) ([]*Order, error)
	UpdateOrder(ctx context.Context, order *Order) error
}

type AccrualService interface {
	GetOrderAccrual(ctx context.Context, order *Order) (*OrderAccrual, error)
}

var ErrOrderWasRegisteredEarlier = errors.New("the order was registered earlier")

func (o *OrderDTO) AddOrder(ctx context.Context, db OrderStorage) (*Order, error) {
	return db.AddOrder(ctx, o)
}

func (o *OrderDTO) GetOrder(ctx context.Context, db OrderStorage) (*Order, error) {
	return db.GetOrder(ctx, o)
}

func (o *OrderDTO) NumberIsCorrect() bool {
	var t = [...]int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}

	odd := len(o.Number) & 1
	var sum int
	for i, c := range o.Number {
		if c < '0' || c > '9' {
			return false
		}
		if i&1 == odd {
			sum += t[c-'0']
		} else {
			sum += int(c - '0')
		}
	}

	return sum%10 == 0
}

func (o *Order) Update(ctx context.Context, db OrderStorage) error {
	return db.UpdateOrder(ctx, o)
}

func RunOrdersAccrual(ctx context.Context, a AccrualService, db OrderStorage) error {
	ors, err := db.GetOrdersForAccrual(ctx)
	if err != nil {
		return fmt.Errorf("get orders for accrual failed err: %w", err)
	}

	for _, o := range ors {
		oa, err := a.GetOrderAccrual(ctx, o)
		if err != nil {
			return fmt.Errorf("get order accrual failed err: %w", err)
		}

		o.Status = oa.Status
		o.Accrual = oa.Accrual

		if err := o.Update(ctx, db); err != nil {
			return fmt.Errorf("update order failed err: %w", err)
		}
	}

	return nil
}
