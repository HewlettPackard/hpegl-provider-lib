// (C) Copyright 2024 Hewlett Packard Enterprise Development LP

package issuertoken

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/provider"
)

func generateExpParams(iamVersion provider.IAMVersion) url.Values {
	expParams := url.Values{}
	expParams.Add("client_id", "clientID")
	expParams.Add("client_secret", "clientSecret")
	expParams.Add("grant_type", "client_credentials")
	if iamVersion == provider.IAMVersionGLCS {
		expParams.Add("scope", "hpe-tenant")
	}

	return expParams
}

func TestGenerateParamsAndURL(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name       string
		iamVersion provider.IAMVersion
		expParams  url.Values
		hasError   bool
	}{
		{
			name:       "valid IAM version GLCS",
			iamVersion: provider.IAMVersionGLCS,
			expParams:  generateExpParams(provider.IAMVersionGLCS),
			hasError:   false,
		},
		{
			name:       "valid IAM version GLP",
			iamVersion: provider.IAMVersionGLP,
			expParams:  generateExpParams(provider.IAMVersionGLP),
			hasError:   false,
		},
		{
			name:       "invalid IAM version",
			iamVersion: "invalid",
			expParams:  nil,
			hasError:   true,
		},
	}

	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			params, _, err := generateParamsAndURL("clientID", "clientSecret", "identityServiceURL", string(tc.iamVersion))
			assert.Equal(t, tc.expParams, params)
			if tc.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
