package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	"github.com/ArtemShalinFe/gophermart/cmd/internal/models"
)

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	hashc := NewMockHashController(ctrl)

	var test = "test"

	u1Dto := &models.UserDTO{
		Login:    test,
		Password: test,
	}

	u2Dto := &models.UserDTO{
		Login:    "test2",
		Password: "test3",
	}

	u1 := &models.User{
		Login:        test,
		PasswordHash: test,
	}

	hc := hashc.EXPECT()
	hc.HashPassword(u1Dto.Password).Return(u1.PasswordHash, nil)
	hc.HashPassword(u2Dto.Password).Return(u2Dto.Password, nil)

	mr := db.EXPECT()

	addUserCall := mr.AddUser(gomock.Any(), u1Dto)
	addUserCall.Return(u1, nil)

	mr.AddUser(gomock.Any(), u2Dto).After(addUserCall).Return(nil, models.ErrLoginIsBusy)

	h, err := NewHandlers([]byte("keyRegister"), db, zap.L().Sugar(), time.Hour*1, hashc)
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		UserDTO interface{}
		name    string
		url     string
		method  string
		status  int
	}{
		{
			name:   "Test user register",
			url:    "/api/user/register",
			status: 200,
			method: http.MethodPost,
			UserDTO: &models.UserDTO{
				Login:    test,
				Password: test,
			},
		},
		{
			name:   "Test user register with same login",
			url:    "/api/user/register",
			status: 409,
			method: http.MethodPost,
			UserDTO: &models.UserDTO{
				Login:    "test2",
				Password: "test3",
			},
		},
		{
			name:    "Test user register with broken body #1",
			url:     "/api/user/register",
			status:  400,
			method:  http.MethodPost,
			UserDTO: nil,
		},
		{
			name:    "Test user register with broken body #2",
			url:     "/api/user/register",
			status:  400,
			method:  http.MethodPost,
			UserDTO: &models.UserDTO{Login: "", Password: ""},
		},
		{
			name:   "Test user register with broken body #3",
			url:    "/api/user/register",
			status: 400,
			method: http.MethodPost,
			UserDTO: struct {
				name string
				psw  string
			}{name: "test123", psw: "test321"},
		},
	}

	for _, v := range tests {
		v := v

		b, err := json.Marshal(v.UserDTO)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t, testServer, v.method, v.url, "", bytes.NewBuffer(b))
		resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode,
			fmt.Sprintf("TestRegisterHandler: %s URL: %s, want: %d, have: %d",
				v.name, v.url, v.status, resp.StatusCode))
	}
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	hashc := NewMockHashController(ctrl)

	var test = "test"

	u1Dto := &models.UserDTO{
		Login:    test,
		Password: test,
	}

	u2Dto := &models.UserDTO{
		Login:    "test2",
		Password: "broken password 123",
	}

	u1 := &models.User{
		Login:        test,
		PasswordHash: test,
	}

	hc := hashc.EXPECT()
	hc.HashPassword(u1Dto.Password).Return(u1.PasswordHash, nil)
	hc.CheckPasswordHash(u1.PasswordHash, u1Dto.Password).Return(true)

	hc.HashPassword(u2Dto.Password).Return(u2Dto.Password, nil)

	mr := db.EXPECT()

	mr.GetUser(gomock.Any(), u1Dto).Return(u1, nil)
	mr.GetUser(gomock.Any(), u2Dto).Return(nil, models.ErrUnknowUser)

	h, err := NewHandlers([]byte("keyLogin"), db, zap.L().Sugar(), time.Hour*1, hashc)
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		UserDTO any
		name    string
		url     string
		method  string
		status  int
	}{
		{
			name:    "Test user login",
			url:     "/api/user/login",
			status:  200,
			method:  http.MethodPost,
			UserDTO: &models.UserDTO{Login: test, Password: test},
		},
		{
			name:    "Test user login with broken pass",
			url:     "/api/user/login",
			status:  401,
			method:  http.MethodPost,
			UserDTO: &models.UserDTO{Login: "test2", Password: "broken password 123"},
		},
		{
			name:    "Test user login with broken body #1",
			url:     "/api/user/login",
			status:  400,
			method:  http.MethodPost,
			UserDTO: nil,
		},
		{
			name:    "Test user login with broken body #2",
			url:     "/api/user/login",
			status:  400,
			method:  http.MethodPost,
			UserDTO: &models.UserDTO{Login: "", Password: ""},
		},
	}

	for _, v := range tests {
		v := v

		b, err := json.Marshal(v.UserDTO)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t,
			testServer,
			v.method,
			v.url,
			"",
			bytes.NewBuffer(b))

		resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode,
			fmt.Sprintf("TestLoginHandler: %s URL: %s, want: %d, have: %d",
				v.name, v.url, v.status, resp.StatusCode))
	}
}

func TestAddOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	hashc := NewMockHashController(ctrl)

	o1Dto := &models.OrderDTO{
		Number: "49927398716",
		UserID: "1",
	}

	o2Dto := &models.OrderDTO{
		Number: "1234567812345670",
		UserID: "1",
	}

	o3Dto := &models.OrderDTO{
		Number: "4026843483168683",
		UserID: "2",
	}

	o1 := &models.Order{
		ID:         "1",
		UserID:     "1",
		Number:     "49927398716",
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	o2 := &models.Order{
		ID:         "2",
		UserID:     "1",
		Number:     "1234567812345670",
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	o3 := &models.Order{
		ID:         "3",
		UserID:     "1",
		Number:     "4026843483168683",
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	var test = "test"

	u1Dto := &models.UserDTO{
		Login:    test,
		Password: "",
	}

	u2Dto := &models.UserDTO{
		Login:    test,
		Password: "",
	}

	u3Dto := &models.UserDTO{
		Login:    "test2",
		Password: "",
	}

	u4Dto := &models.UserDTO{
		Login:    "test2",
		Password: "",
	}

	u1 := &models.User{
		ID:           "1",
		Login:        test,
		PasswordHash: test,
	}

	u2 := &models.User{
		ID:           "2",
		Login:        "test2",
		PasswordHash: "test2",
	}

	hc := hashc.EXPECT()
	hc.HashPassword(u1Dto.Password).Return(u1.PasswordHash, nil)
	hc.HashPassword(u2Dto.Password).Return(u2.PasswordHash, nil)
	hc.HashPassword(u3Dto.Password).Return(u2.PasswordHash, nil)
	hc.HashPassword(u4Dto.Password).Return(u2.PasswordHash, nil)

	hc.CheckPasswordHash(u1.PasswordHash, u1Dto.Password).Return(true)
	hc.CheckPasswordHash(u2.PasswordHash, u2Dto.Password).Return(true)
	hc.CheckPasswordHash(u2.PasswordHash, u3Dto.Password).Return(true)
	hc.CheckPasswordHash(u2.PasswordHash, u4Dto.Password).Return(true)

	mr := db.EXPECT()

	mr.GetUser(gomock.Any(), u1Dto).AnyTimes().Return(u1, nil)
	mr.GetUser(gomock.Any(), u2Dto).AnyTimes().Return(u1, nil)

	mr.GetUser(gomock.Any(), u3Dto).AnyTimes().Return(u2, nil)
	mr.GetUser(gomock.Any(), u4Dto).AnyTimes().Return(u2, nil)

	mr.AddOrder(gomock.Any(), o1Dto).AnyTimes().Return(o1, nil)
	mr.AddOrder(gomock.Any(), o2Dto).AnyTimes().Return(nil, models.ErrOrderWasRegisteredEarlier)
	mr.AddOrder(gomock.Any(), o3Dto).AnyTimes().Return(nil, models.ErrOrderWasRegisteredEarlier)

	mr.GetOrder(gomock.Any(), o1Dto).AnyTimes().Return(o1, nil)
	mr.GetOrder(gomock.Any(), o2Dto).AnyTimes().Return(o2, nil)
	mr.GetOrder(gomock.Any(), o3Dto).AnyTimes().Return(o3, nil)

	h, err := NewHandlers([]byte("keyAddOrder"), db, zap.L().Sugar(), time.Hour*1, hashc)
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		authReq any
		body    any
		name    string
		url     string
		method  string
		status  int
	}{
		{
			name:    "Add order unauthorized",
			url:     "/api/user/orders",
			status:  401,
			method:  http.MethodPost,
			authReq: nil,
			body:    49927398716,
		},
		{
			name:    "Add first order",
			url:     "/api/user/orders",
			status:  202,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: test, Password: test},
			body:    49927398716,
		},
		{
			name:    "Add same order",
			url:     "/api/user/orders",
			status:  200,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: test, Password: test},
			body:    1234567812345670,
		},
		{
			name:    "Add added order",
			url:     "/api/user/orders",
			status:  409,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: "test2", Password: "test2"},
			body:    4026843483168683,
		},
	}

	for _, v := range tests {
		v := v

		b, err := json.Marshal(v.body)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t,
			testServer,
			v.method,
			v.url,
			GetAuthorizationToken(t, testServer, v.authReq),
			bytes.NewBuffer(b))

		resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode,
			fmt.Sprintf("TestAddOrderHandler: %s URL: %s, want: %d, have: %d",
				v.name, v.url, v.status, resp.StatusCode))
	}
}

func TestGetOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)
	hashc := NewMockHashController(ctrl)

	var test = "test"

	o1 := &models.Order{
		ID:         "1",
		UserID:     "1",
		Number:     "49927398716",
		Status:     "NEW",
		Accrual:    0,
		UploadedAt: time.Now(),
	}

	var ors []*models.Order
	ors = append(ors, o1)

	u1Dto := &models.UserDTO{
		Login:    test,
		Password: test,
	}

	u1 := &models.User{
		ID:           "1",
		Login:        test,
		PasswordHash: test,
	}

	mr := db.EXPECT()

	mr.GetUser(gomock.Any(), u1Dto).AnyTimes().Return(u1, nil)
	mr.GetUploadedOrders(gomock.Any(), u1).AnyTimes().Return(ors, nil)

	h, err := NewHandlers([]byte("keyAddOrder"), db, zap.L().Sugar(), time.Hour*1, hashc)
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		authReq         any
		body            any
		name            string
		url             string
		method          string
		status          int
		wantOrdersCount int
	}{
		{
			name:    "Get orders unauthorized",
			url:     "/api/user/orders",
			status:  401,
			method:  http.MethodGet,
			authReq: nil,
		},
		{
			name:            "Get orders",
			url:             "/api/user/orders",
			status:          200,
			method:          http.MethodGet,
			authReq:         &models.UserDTO{Login: test, Password: test},
			wantOrdersCount: 1,
		},
	}

	for _, v := range tests {
		v := v

		b, err := json.Marshal(v.body)
		if err != nil {
			t.Error(err)
		}

		resp, bytes := testRequest(t,
			testServer,
			v.method,
			v.url,
			GetAuthorizationToken(t, testServer, v.authReq),
			bytes.NewBuffer(b))

		defer resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode,
			fmt.Sprintf("TestGetOrdersHandler status code: %s URL: %s, want: %d, have: %d",
				v.name, v.url, v.status, resp.StatusCode))

		if resp.StatusCode == http.StatusOK {
			var os []*models.Order

			if err := json.Unmarshal(bytes, &os); err != nil {
				t.Error(err)
			}

			require.Equal(t, v.wantOrdersCount, len(os),
				fmt.Sprintf("TestGetOrdersHandler len orders: %s URL: %s, want: %d, have: %d",
					v.name, v.url, v.status, resp.StatusCode))
		}
	}
}

func GetAuthorizationToken(t *testing.T, ts *httptest.Server, authReq any) string {
	t.Helper()

	if authReq == nil {
		return ""
	}

	b, err := json.Marshal(authReq)
	if err != nil {
		t.Error(err)
	}

	resp, _ := testRequest(t, ts, http.MethodPost, "/api/user/login", "", bytes.NewBuffer(b))
	defer resp.Body.Close()

	return resp.Header.Get(authHeaderName)
}

func testRequest(t *testing.T, ts *httptest.Server,
	method string, path string, jwt string, body io.Reader) (*http.Response, []byte) {
	t.Helper()

	r, err := url.JoinPath(ts.URL, path)
	if err != nil {
		t.Errorf("URL %s test request  error : %v", err, path)
	}
	req, err := http.NewRequest(method, r, body)
	if jwt != "" {
		req.Header.Set(authHeaderName, jwt)
	}
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
