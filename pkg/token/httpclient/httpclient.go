// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package httpclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/Hewlettpackard/hpegl-provider-lib/pkg/token/identitytoken"
	"github.com/Hewlettpackard/hpegl-provider-lib/pkg/token/issuertoken"
	tokenutil "github.com/Hewlettpackard/hpegl-provider-lib/pkg/token/token-util"
)

type Client struct {
	passedInToken       string
	identityServiceURL  string
	httpClient          tokenutil.HttpClient
	vendedServiceClient bool
}

// New creates a new identity Client object
func New(identityServiceURL string, vendedServiceClient bool, passedInToken string) *Client {
	client := &http.Client{Timeout: 10 * time.Second}
	identityServiceURL = strings.TrimRight(identityServiceURL, "/")
	return &Client{
		passedInToken:       passedInToken,
		identityServiceURL:  identityServiceURL,
		httpClient:          client,
		vendedServiceClient: vendedServiceClient,
	}
}

func (c *Client) GenerateToken(ctx context.Context, tenantID, clientID, clientSecret string) (string, error) {
	// we don't have a passed-in token, so we need to actually generate a token
	if c.passedInToken == "" {
		if c.vendedServiceClient {
			token, err := issuertoken.GenerateToken(ctx, tenantID, clientID, clientSecret, c.identityServiceURL, c.httpClient)
			return token, err
		} else {
			token, err := identitytoken.GenerateToken(ctx, tenantID, clientID, clientSecret, c.identityServiceURL, c.httpClient)
			return token, err
		}
	}

	// we have a passed-in token, return it
	return c.passedInToken, nil
}
