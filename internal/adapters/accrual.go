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
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/internal/config"
	"github.com/ArtemShalinFe/gophermart/internal/models"
)

type Accrual struct {
	httpClient *retryablehttp.Client
	log        *zap.SugaredLogger
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

func NewAccrualClient(cfg config.Config, log *zap.SugaredLogger) *Accrual {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.CheckRetry = checkRetry
	client.Backoff = backoff

	return &Accrual{
		host:       cfg.Accrual,
		log:        log,
		httpClient: client,
	}
}

func checkRetry(ctx context.Context, resp *http.Response, err error) (bool, error) {
	check, err := retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	if err != nil {
		return false, fmt.Errorf("accrual error in default retry policy : %w", err)
	}
	return check, nil
}

func backoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	return retryablehttp.LinearJitterBackoff(min, max, attemptNum, resp)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			a.log.Errorf("closing body was failed err: %w", err)
		}
	}()

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

func (a *Accrual) request(ctx context.Context, order *models.Order) (*retryablehttp.Request, error) {
	url, err := url.JoinPath(a.host, "/api/orders/", order.Number)
	if err != nil {
		return nil, fmt.Errorf("failed build url err: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed build accrual request err: %w", err)
	}

	return req, nil
}
