package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {

	h := NewHandlers()
	r := initRouter(h)

	testServer := httptest.NewServer(r)
	defer testServer.Close()

	var tests = []struct {
		name   string
		url    string
		status int
		method string
	}{
		{
			name:   "Test Ping",
			url:    "/ping",
			status: 200,
			method: http.MethodGet,
		},
	}

	for _, v := range tests {

		resp, _ := testRequest(t, testServer, v.method, v.url, bytes.NewBuffer([]byte("")))
		defer resp.Body.Close()

		require.Equal(t, v.status, resp.StatusCode, fmt.Sprintf("%s URL: %s, want: %d, have: %d", v.name, v.url, resp.StatusCode, v.status))

	}
}

func testRequest(t *testing.T, ts *httptest.Server, method string, path string, body io.Reader) (*http.Response, []byte) {

	r, err := url.JoinPath(ts.URL, path)
	if err != nil {
		t.Errorf("URL %s test request  error : %v", err, path)
	}
	req, err := http.NewRequest(method, r, body)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, respBody
}
