// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package provider

import (
	"context"
	"testing"

	"github.com/Hewlettpackard/hpegl-provider-lib/pkg/registration"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func testResource() *schema.Resource {
	return &schema.Resource{}
}

type Registration struct {
	serviceName string
	resources   map[string]*schema.Resource
	datasources map[string]*schema.Resource
}

func (r Registration) Name() string {
	return r.serviceName
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

func providerConfigure(p *schema.Provider) schema.ConfigureContextFunc { // nolint staticcheck
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return nil, nil
	}
}

func TestNewProviderFunc(t *testing.T) {
	t.Parallel()
	testcases := []struct {
		name     string
		regs     []Registration
		panicMsg string
	}{
		{
			name: "success",
			regs: []Registration{
				{
					serviceName: "test-service",
					resources: map[string]*schema.Resource{
						"test-resource": testResource(),
					},
					datasources: map[string]*schema.Resource{
						"test-datasource": testResource(),
					},
				},
			},
		},
		{
			name: "success two services",
			regs: []Registration{
				{
					serviceName: "test-service",
					resources: map[string]*schema.Resource{
						"test-resource": testResource(),
					},
					datasources: map[string]*schema.Resource{
						"test-datasource": testResource(),
					},
				},
				{
					serviceName: "test-service2",
					resources: map[string]*schema.Resource{
						"test-resource2": testResource(),
					},
					datasources: map[string]*schema.Resource{
						"test-datasource2": testResource(),
					},
				},
			},
		},
		{
			name: "duplicate resource",
			regs: []Registration{
				{
					serviceName: "test-service",
					resources: map[string]*schema.Resource{
						"test-resource": testResource(),
					},
				},
				{
					serviceName: "test-service2",
					resources: map[string]*schema.Resource{
						"test-resource": testResource(),
					},
				},
			},
			panicMsg: "resource name test-resource is repeated in service test-service2",
		},
		{
			name: "duplicate data source",
			regs: []Registration{
				{
					serviceName: "test-service",
					datasources: map[string]*schema.Resource{
						"test-datasource": testResource(),
					},
				},
				{
					serviceName: "test-service2",
					datasources: map[string]*schema.Resource{
						"test-datasource": testResource(),
					},
				},
			},
			panicMsg: "data-source name test-datasource is repeated in service test-service2",
		},
		{
			name: "duplicate service name",
			regs: []Registration{
				{
					serviceName: "test-service",
				},
				{
					serviceName: "test-service",
				},
			},
			panicMsg: "service name test-service is repeated",
		},
	}

	for _, testcase := range testcases {
		tc := testcase
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var regs []registration.ServiceRegistration

			if len(tc.regs) == 1 {
				regs = ServiceRegistrationSlice(tc.regs[0])
			} else {
				regs = make([]registration.ServiceRegistration, len(tc.regs))
				for i, reg := range tc.regs {
					regs[i] = reg
				}
			}

			defer func() {
				r := recover()
				if r != nil {
					if tc.panicMsg != "" {
						assert.Equal(t, tc.panicMsg, r)
					} else {
						assert.Equal(t, nil, r)
					}
				}
			}()

			NewProviderFunc(regs, providerConfigure)()
		})
	}
}
