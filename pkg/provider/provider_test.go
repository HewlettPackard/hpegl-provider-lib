package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/common"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/retrieve"
	"github.com/hewlettpackard/hpegl-provider-lib/pkg/token/serviceclient"
)

func testResource() *schema.Resource {
	return &schema.Resource{}
}

type Registration struct {
	resources   map[string]*schema.Resource
	datasources map[string]*schema.Resource
}

func (r Registration) Name() string {
	return "test-service"
}

func (r Registration) SupportedDataSources() map[string]*schema.Resource {
	return r.datasources
}

func (r Registration) SupportedResources() map[string]*schema.Resource {
	return r.resources
}

func (r Registration) ProviderSchemaEntry() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
}

type client struct{}

func (c *client) NewClient(r *schema.ResourceData) (interface{}, error) {
	return nil, nil
}

func (c *client) ServiceName() string {
	return "test-service"
}

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc { // nolint staticcheck
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		c := &client{}
		cli, err := c.NewClient(d)
		if err != nil {
			return nil, diag.Errorf("error in creating client: %s", err)
		}
		// Initialise token handler
		h, err := serviceclient.NewHandler(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		return map[string]interface{}{
			c.ServiceName():                 cli,
			common.TokenRetrieveFunctionKey: retrieve.NewTokenRetrieveFunc(h),
		}, nil
	}
}

func TestNewProviderFunc(t *testing.T) {
	testcases := []struct {
		name        string
		resources   map[string]*schema.Resource
		datasources map[string]*schema.Resource
	}{
		{
			name: "success",
			resources: map[string]*schema.Resource{
				"test-resource": testResource(),
			},
			datasources: map[string]*schema.Resource{
				"test-datasource": testResource(),
			},
		},
	}

	for _, testcase := range testcases {
		tc := testcase

		reg := Registration{
			resources:   tc.resources,
			datasources: tc.datasources,
		}

		provFunc := NewProviderFunc(ServiceRegistrationSlice(reg), providerConfigure)
		provFunc()
	}
}
