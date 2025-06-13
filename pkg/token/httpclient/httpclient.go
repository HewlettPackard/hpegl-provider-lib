// (C) Copyright 2021-2025 Hewlett Packard Enterprise Development LP

package httpclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/identitytoken"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/issuertoken"
	tokenutil "github.com/hewlettpackard/hpegl-provider-lib/pkg/token/token-util"
)

type Client struct {
	passedInToken       string
	identityServiceURL  string
	httpClient          tokenutil.HttpClient
	vendedServiceClient bool
}

// New creates a new identity Client object
func New(identityServiceURL string, iamInsecure, vendedServiceClient bool, passedInToken string) *Client {
	identityServiceURL = strings.TrimRight(identityServiceURL, "/")

	return &Client{
		passedInToken:       passedInToken,
		identityServiceURL:  identityServiceURL,
		httpClient:          createHttpClient(iamInsecure),
		vendedServiceClient: vendedServiceClient,
	}
}

func createHttpClient(iamInsecure bool) *http.Client {
	if iamInsecure {
		// If insecure, we need to set the transport to allow insecure connections
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}
		return &http.Client{
			Transport: transport,
			Timeout:   120 * time.Second,
		}
	}

	// Use the default HTTP client with a timeout
	return &http.Client{Timeout: 120 * time.Second}
}

func (c *Client) GenerateToken(ctx context.Context, tenantID, clientID, clientSecret, iamVersion string) (string, error) {
	// we don't have a passed-in token, so we need to actually generate a token
	if c.passedInToken == "" {
		if c.vendedServiceClient {
			token, err := issuertoken.GenerateToken(
				ctx, clientID, clientSecret, c.identityServiceURL, c.httpClient, iamVersion)

			return token, err
		}

		token, err := identitytoken.GenerateToken(ctx, tenantID, clientID, clientSecret, c.identityServiceURL, c.httpClient)

		return token, err
	}

	// we have a passed-in token, return it
	return c.passedInToken, nil
}
