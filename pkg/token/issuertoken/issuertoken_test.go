package issuertoken

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// nolint: tparallel
func TestDoRetries(t *testing.T) {
	t.Parallel()
	totalRetries := 0
	testcases := []struct {
		name           string
		call           func() (*http.Response, error)
		responseStatus int
		err            error
	}{
		{
			name: "status 500",
			call: func() (*http.Response, error) {
				totalRetries++

				return &http.Response{StatusCode: http.StatusInternalServerError}, nil
			},
			responseStatus: http.StatusInternalServerError,
		},
		{
			name: "status 429",
			call: func() (*http.Response, error) {
				totalRetries++

				return &http.Response{StatusCode: http.StatusTooManyRequests}, nil
			},
			responseStatus: http.StatusTooManyRequests,
		},
		{
			name: "status 502 no retry",
			call: func() (*http.Response, error) {
				totalRetries++

				return &http.Response{StatusCode: http.StatusBadGateway}, nil
			},
			responseStatus: http.StatusBadGateway,
		},
		{
			name: "no url",
			call: func() (*http.Response, error) {
				return nil, errors.New("http: nil Request.URL")
			},
			err: errors.New("http: nil Request.URL"),
		},
	}

	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			resp, err := doRetries(tc.call, 1) // nolint: bodyclose
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
			} else {
				assert.Equal(t, tc.responseStatus, resp.StatusCode)

				// only 429 and 500 status codes should retry
				if tc.responseStatus == http.StatusBadGateway {
					assert.Equal(t, 1, totalRetries)
				} else {
					assert.Equal(t, 2, totalRetries)
				}

				totalRetries = 0
			}
		})
	}
}

type testHTTPClient struct {
	statusCode int
	body       TokenResponse
}

type bodyReadCloser struct {
	body      []byte
	readIndex int
}

func (b *bodyReadCloser) Close() error {
	return nil
}

func (b *bodyReadCloser) Read(p []byte) (n int, err error) {
	if b.readIndex >= len(b.body) {
		err = io.EOF

		return
	}

	n = copy(p, b.body[b.readIndex:])
	b.readIndex += n

	return n, nil
}

func (h *testHTTPClient) Do(req *http.Request) (*http.Response, error) {
	bytes, err := json.Marshal(h.body)
	if err != nil {
		return nil, err
	}

	body := &bodyReadCloser{body: bytes}

	return &http.Response{StatusCode: h.statusCode, Body: body}, nil
}

func createTestClient(identityServiceURL string, statusCode int, token TokenResponse) *Client {
	c := New(identityServiceURL)
	c.httpClient = &testHTTPClient{
		statusCode: statusCode,
		body:       token,
	}

	return c
}

func TestGenerateToken(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name       string
		ctx        context.Context
		url        string
		statusCode int
		token      TokenResponse
		err        error
	}{
		{
			name:       "success",
			ctx:        context.Background(),
			url:        "https://hpe-greenlake-tenant.okta.com/oauth2/default",
			statusCode: http.StatusOK,
			token: TokenResponse{
				AccessToken: "access-token",
			},
		},
		{
			name:       "status code 404",
			url:        "https://hpe-greenlake-tenant.okta.com/oauth2/default",
			ctx:        context.Background(),
			statusCode: http.StatusNotFound,
			err:        errors.New("Unexpected status code 404"),
		},
		{
			name: "no context",
			url:  "https://hpe-greenlake-tenant.okta.com/oauth2/default",
			ctx:  nil,
			err:  errors.New("net/http: nil Context"),
		},
		{
			name:       "status code 400",
			url:        "https://hpe-greenlake-tenant.okta.com/oauth2/default",
			ctx:        context.Background(),
			statusCode: http.StatusBadRequest,
			err:        errors.New("Bad request: {\"token_type\":\"\",\"expires_in\":0,\"access_token\":\"\",\"scope\":\"\"}"),
		},
		{
			name:       "status code 401",
			url:        "https://hpe-greenlake-tenant.okta.com/oauth2/default",
			ctx:        context.Background(),
			statusCode: http.StatusUnauthorized,
			err:        errors.New("Forbidden: "),
		},
	}

	for _, testcase := range testcases {
		tc := testcase

		c := createTestClient(tc.url, tc.statusCode, tc.token)

		token, err := c.GenerateToken(tc.ctx, "", "", "")
		if tc.err != nil {
			assert.EqualError(t, err, tc.err.Error())
		}

		assert.Equal(t, tc.token.AccessToken, token)
	}
}
