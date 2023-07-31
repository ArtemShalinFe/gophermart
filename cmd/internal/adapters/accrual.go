package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ArtemShalinFe/gophermart/cmd/internal/config"
	"github.com/ArtemShalinFe/gophermart/cmd/internal/models"
)

type Accrual struct {
	httpClient *http.Client
	host       string
}

type AccrualErr struct {
	error
	timeout int
}

var errOrderNotRegistered = errors.New("the order is not registered in the payment system")
var errTooManyRequests = errors.New("too many requests")

func newAccrualErr(err error, timeout int) *AccrualErr {
	return &AccrualErr{
		error:   err,
		timeout: timeout,
	}
}

func (ae *AccrualErr) TimeoutSec() (int, bool) {
	return ae.timeout, ae.timeout > 0
}

func (ae *AccrualErr) IsOrderNotRegistered() bool {
	return errors.Is(ae.error, errOrderNotRegistered)
}

func (ae *AccrualErr) IsTooManyRequests() bool {
	return errors.Is(ae.error, errTooManyRequests)
}

func NewAccrualClient(cfg config.Config) *Accrual {
	return &Accrual{
		host:       cfg.Accrual,
		httpClient: &http.Client{},
	}
}

func (a *Accrual) GetOrderAccrual(ctx context.Context, order *models.Order) (*models.OrderAccrual, *AccrualErr) {
	req, err := a.request(ctx, order)
	if err != nil {
		return nil, newAccrualErr(fmt.Errorf("failed prepare accrual request err: %w", err), 0)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, newAccrualErr(fmt.Errorf("failed exec accrual request err: %w", err), 0)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil, newAccrualErr(errOrderNotRegistered, 0)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfterString := resp.Header.Get("Retry-After")
		retryAfter, err := strconv.Atoi(retryAfterString)
		if err != nil {
			return nil, newAccrualErr(fmt.Errorf("parse value Retry-After err: %w", err), 0)
		}
		return nil, newAccrualErr(errTooManyRequests, retryAfter)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, newAccrualErr(fmt.Errorf("failed reading response body %s err: %w", string(res), err), 0)
	}

	var oa models.OrderAccrual
	if err := json.Unmarshal(res, &oa); err != nil {
		return nil, newAccrualErr(fmt.Errorf("failed unmarshal response body %s err: %w", string(res), err), 0)
	}

	return &oa, nil
}

func (a *Accrual) request(ctx context.Context, order *models.Order) (*http.Request, error) {
	url, err := url.JoinPath(a.host, "/api/orders/", order.Number)
	if err != nil {
		return nil, fmt.Errorf("failed build url err: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed build accrual request err: %w", err)
	}

	return req, nil
}
