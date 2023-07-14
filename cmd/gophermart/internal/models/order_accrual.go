package models

type OrderAccrual struct {
	OrderNumber string `json:"order"`
	Status      string `json:"status"`
	Accrual     int64  `json:"accrual"`
}
