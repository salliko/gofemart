package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/salliko/gofemart/config"
	"github.com/salliko/gofemart/internal/databases"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var TestCookie *http.Cookie

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	if TestCookie != nil {
		req.AddCookie(TestCookie)
	}

	if cookie, err := req.Cookie("user_id"); err == nil {
		log.Print("Читаем куку до запроса")
		log.Print(cookie.Value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)

	if len(resp.Cookies()) != 0 {
		TestCookie = resp.Cookies()[0]
	}

	require.NoError(t, err)

	return resp
}

func TestRouter(t *testing.T) {
	cfg := config.Config{
		RunAddress:  "localhost:8080",
		DatabaseURL: "postgres://postgres:postgres@localhost:5432",
	}

	db, err := databases.NewPostgresqlDatabase(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	type want struct {
		status int
		body   string
		header string
	}

	tests := []struct {
		name   string
		path   string
		method string
		body   string
		want   want
	}{
		{
			name:   "POST user login неверная пара логин/пароль",
			path:   "/api/user/login",
			method: http.MethodPost,
			body:   `{"login":"test", "password": "1234"}`,
			want: want{
				status: http.StatusUnauthorized,
				header: `application/json; charset=UTF-8`,
			},
		},
		{
			name:   "POST user register",
			path:   "/api/user/register",
			method: http.MethodPost,
			body:   `{"login":"test", "password": "1234"}`,
			want: want{
				status: http.StatusOK,
				header: `application/json; charset=UTF-8`,
			},
		},
		{
			name:   "POST user register StatusConflict",
			path:   "/api/user/register",
			method: http.MethodPost,
			body:   `{"login":"test", "password": "1234"}`,
			want: want{
				status: http.StatusConflict,
				header: `application/json; charset=UTF-8`,
			},
		},
		{
			name:   "POST user login",
			path:   "/api/user/login",
			method: http.MethodPost,
			body:   `{"login":"test", "password": "1234"}`,
			want: want{
				status: http.StatusOK,
				header: `application/json; charset=UTF-8`,
			},
		},
		{
			name:   "POST user orders",
			path:   "/api/user/orders",
			method: http.MethodPost,
			body:   "5567876",
			want: want{
				status: http.StatusAccepted,
			},
		},
		{
			name:   "POST user orders",
			path:   "/api/user/orders",
			method: http.MethodPost,
			body:   "5567876",
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

	r := NewRouter(cfg, db)
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
