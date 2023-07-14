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
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/config"
	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/internal/models"
)

type Accrual struct {
	httpClient *retryablehttp.Client
	host       string
}

var ErrOrderNotRegistered = errors.New("the order is not registered in the payment system")

func NewAccrualClient(cfg *config.Config, logger retryablehttp.LeveledLogger) *Accrual {
	retryClient := retryablehttp.NewClient()
	retryClient.CheckRetry = CheckRetry
	retryClient.Backoff = Backoff

	return &Accrual{
		host:       cfg.Accrual,
		httpClient: retryClient,
	}
}

func CheckRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	return false, err
}

func Backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	const defaultTimeout = 120

	if resp.StatusCode != http.StatusTooManyRequests {
		return defaultTimeout * time.Second
	}

	retryAfterString := resp.Header.Get("Retry-After")
	retryAfter, err := strconv.ParseInt(retryAfterString, 10, 64)
	if err != nil {
		return defaultTimeout * time.Second
	}

	return time.Duration(retryAfter) * time.Second
}

func (a *Accrual) GetOrderAccrual(ctx context.Context, order *models.Order) (*models.OrderAccrual, error) {
	url, err := url.JoinPath("/api/orders/", order.Number)
	if err != nil {
		return nil, fmt.Errorf("failed build url err: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed build accrual request err: %w", err)
	}
	req.Close = true

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed exec accrual request err: %w", err)
	}

	defer func() {
		if errS := resp.Body.Close(); errS != nil {
			err = errors.Join(err, errS)
		}
	}()

	if resp.StatusCode == http.StatusNoContent {
		return nil, ErrOrderNotRegistered
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body err: %w", err)
	}

	var oa models.OrderAccrual
	if err := json.Unmarshal(res, &oa); err != nil {
		return nil, fmt.Errorf("failed unmarshal response body err: %w", err)
	}

	return &oa, nil
}
