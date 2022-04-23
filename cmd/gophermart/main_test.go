package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func TestRouter(t *testing.T) {
	// cfg := config.Config{
	// 	RunAddress: "localhost:8080",
	// }

	type want struct {
		status int
		body   string
	}

	tests := []struct {
		name   string
		path   string
		method string
		body   string
		want   want
	}{
		{
			name:   "POST user register",
			path:   "/api/user/register",
			method: http.MethodPost,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "POST user login",
			path:   "/api/user/login",
			method: http.MethodPost,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "POST user orders",
			path:   "/api/user/orders",
			method: http.MethodPost,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "GET user orders",
			path:   "/api/user/orders",
			method: http.MethodGet,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "GET user withdrawals",
			path:   "/api/user/withdrawals",
			method: http.MethodGet,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "GET user balance",
			path:   "/api/user/balance",
			method: http.MethodGet,
			want: want{
				status: http.StatusOK,
			},
		},
		{
			name:   "POST user balance withdraw",
			path:   "/api/user/balance/withdraw",
			method: http.MethodPost,
			want: want{
				status: http.StatusOK,
			},
		},
	}

	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.path, strings.NewReader(tt.body))
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)

			assert.NoError(t, err)
			assert.Equal(t, tt.want.status, resp.StatusCode)
			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}
