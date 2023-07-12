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

	"github.com/ArtemShalinFe/gophermart/cmd/gophermart/models"
)

func TestRegisterHandler(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)

	u1_dto := &models.UserDTO{
		Login:    "test",
		Password: models.EncodePassword("test"),
	}

	u2_dto := &models.UserDTO{
		Login:    "test2",
		Password: models.EncodePassword("test3"),
	}

	u1 := &models.User{
		Login:          "test",
		PasswordBase64: models.EncodePassword("test"),
	}

	mr := db.EXPECT()

	addUserCall := mr.AddUser(gomock.Any(), u1_dto)
	addUserCall.Return(u1, nil)

	mr.AddUser(gomock.Any(), u2_dto).After(addUserCall).Return(nil, models.ErrLoginIsBusy)

	h, err := NewHandlers([]byte("key"), db, zap.L())
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		name    string
		url     string
		status  int
		method  string
		UserDTO interface{}
	}{
		{
			name:   "Test user register",
			url:    "/api/user/register",
			status: 200,
			method: http.MethodPost,
			UserDTO: &models.UserDTO{
				Login:    "test",
				Password: "test",
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
			}{name: "test", psw: "test"},
		},
	}

	for _, v := range tests {

		v := v

		b, err := json.Marshal(v.UserDTO)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t, testServer, v.method, v.url, "", bytes.NewBuffer(b))
		defer resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("%s URL: %s, want: %d, have: %d", v.name, v.url, v.status, resp.StatusCode))

	}
}

func TestLoginHandler(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)

	u1_dto := &models.UserDTO{
		Login:    "test",
		Password: "test",
	}

	u2_dto := &models.UserDTO{
		Login:    "test2",
		Password: "broken password 123",
	}

	u1 := &models.User{
		Login:          "test",
		PasswordBase64: models.EncodePassword("test"),
	}

	mr := db.EXPECT()

	mr.GetUser(gomock.Any(), u1_dto).Return(u1, nil)
	mr.GetUser(gomock.Any(), u2_dto).Return(nil, models.ErrUnknowUser)

	h, err := NewHandlers([]byte("key"), db, zap.L())
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		name    string
		url     string
		status  int
		method  string
		UserDTO any
	}{
		{
			name:    "Test user login",
			url:     "/api/user/login",
			status:  200,
			method:  http.MethodPost,
			UserDTO: &models.UserDTO{Login: "test", Password: "test"},
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

		resp, _ := testRequest(t, testServer, v.method, v.url, "", bytes.NewBuffer(b))
		defer resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("%s URL: %s, want: %d, have: %d", v.name, v.url, v.status, resp.StatusCode))

	}
}

func TestAddOrderHandler(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db := NewMockStorage(ctrl)

	o1_dto := &models.OrderDTO{
		Number: 1,
		UserId: "1",
	}

	o2_dto := &models.OrderDTO{
		Number: 2,
		UserId: "1",
	}

	o3_dto := &models.OrderDTO{
		Number: 3,
		UserId: "2",
	}

	o1 := &models.Order{
		Id:          "1",
		UserId:      "1",
		Number:      1,
		Status:      "NEW",
		Accrual:     0,
		Uploaded_at: time.Now(),
	}

	o2 := &models.Order{
		Id:          "2",
		UserId:      "1",
		Number:      1,
		Status:      "NEW",
		Accrual:     0,
		Uploaded_at: time.Now(),
	}

	o3 := &models.Order{
		Id:          "3",
		UserId:      "1",
		Number:      1,
		Status:      "NEW",
		Accrual:     0,
		Uploaded_at: time.Now(),
	}

	u1_dto := &models.UserDTO{
		Login:    "test",
		Password: models.EncodePassword("test"),
	}

	u2_dto := &models.UserDTO{
		Login:    "test",
		Password: "test",
	}

	u3_dto := &models.UserDTO{
		Login:    "test2",
		Password: models.EncodePassword("test2"),
	}

	u4_dto := &models.UserDTO{
		Login:    "test2",
		Password: "test2",
	}

	u1 := &models.User{
		Id:             "1",
		Login:          "test",
		PasswordBase64: models.EncodePassword("test"),
	}

	u2 := &models.User{
		Id:             "2",
		Login:          "test2",
		PasswordBase64: models.EncodePassword("test2"),
	}

	mr := db.EXPECT()

	mr.GetUser(gomock.Any(), u1_dto).AnyTimes().Return(u1, nil)
	mr.GetUser(gomock.Any(), u2_dto).AnyTimes().Return(u1, nil)

	mr.GetUser(gomock.Any(), u3_dto).AnyTimes().Return(u2, nil)
	mr.GetUser(gomock.Any(), u4_dto).AnyTimes().Return(u2, nil)

	mr.AddOrder(gomock.Any(), o1_dto).AnyTimes().Return(o1, nil)
	mr.AddOrder(gomock.Any(), o2_dto).AnyTimes().Return(nil, models.ErrOrderWasRegisteredEarlier)
	mr.AddOrder(gomock.Any(), o3_dto).AnyTimes().Return(nil, models.ErrOrderWasRegisteredEarlier)

	mr.GetOrder(gomock.Any(), o1_dto).AnyTimes().Return(o1, nil)
	mr.GetOrder(gomock.Any(), o2_dto).AnyTimes().Return(o2, nil)
	mr.GetOrder(gomock.Any(), o3_dto).AnyTimes().Return(o3, nil)

	h, err := NewHandlers([]byte("key"), db, zap.L())
	if err != nil {
		t.Error(err)
	}
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		name    string
		url     string
		status  int
		method  string
		authReq any
		body    any
	}{
		{
			name:    "Add order unauthorized",
			url:     "/api/user/orders",
			status:  401,
			method:  http.MethodPost,
			authReq: nil,
			body:    1,
		},
		{
			name:    "Add first order",
			url:     "/api/user/orders",
			status:  202,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: "test", Password: "test"},
			body:    1,
		},
		{
			name:    "Add same order",
			url:     "/api/user/orders",
			status:  200,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: "test", Password: "test"},
			body:    2,
		},
		{
			name:    "Add added order",
			url:     "/api/user/orders",
			status:  409,
			method:  http.MethodPost,
			authReq: &models.UserDTO{Login: "test2", Password: "test2"},
			body:    3,
		},
	}

	for _, v := range tests {

		v := v

		b, err := json.Marshal(v.body)
		if err != nil {
			t.Error(err)
		}

		resp, _ := testRequest(t, testServer, v.method, v.url, GetAuthorizationToken(t, testServer, v.authReq), bytes.NewBuffer(b))
		defer resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode,
			fmt.Sprintf("%s URL: %s, want: %d, have: %d", v.name, v.url, v.status, resp.StatusCode))

	}

}

func GetAuthorizationToken(t *testing.T, ts *httptest.Server, authReq any) string {

	if authReq == nil {
		return ""
	}

	b, err := json.Marshal(authReq)
	if err != nil {
		t.Error(err)
	}

	resp, _ := testRequest(t, ts, http.MethodPost, "/api/user/login", "", bytes.NewBuffer(b))
	defer resp.Body.Close()

	return resp.Header.Get("Authorization")

}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, jwt string, body io.Reader) (*http.Response, []byte) {

	r, err := url.JoinPath(ts.URL, path)
	if err != nil {
		t.Errorf("URL %s test request  error : %v", err, path)
	}
	req, err := http.NewRequest(method, r, body)
	if jwt != "" {
		req.Header.Set("Authorization", jwt)
	}
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
