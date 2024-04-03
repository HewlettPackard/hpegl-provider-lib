// (C) Copyright 2024 Hewlett Packard Enterprise Development LP

package issuertoken

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/provider"
)

func TestGenerateParamsAndURL(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name       string
		iamVersion string
		hasError   bool
	}{
		{
			name:       "valid IAM version GLCS",
			iamVersion: provider.IAMVersionGLCS,
			hasError:   false,
		},
		{
			name:       "valid IAM version GLP",
			iamVersion: provider.IAMVersionGLP,
			hasError:   false,
		},
		{
			name:       "invalid IAM version",
			iamVersion: "invalid",
			hasError:   true,
		},
	}

	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, _, err := generateParamsAndURL("clientID", "clientSecret", "identityServiceURL", tc.iamVersion)
			if tc.hasError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
