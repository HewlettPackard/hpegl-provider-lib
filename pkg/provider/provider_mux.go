// (C) Copyright 2021-2024 Hewlett Packard Enterprise Development LP

package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/hewlettpackard/hpegl-provider-lib/pkg/registration"
)

// ProviderForMux is a function that returns a list of ProviderServer functions that can be used with the
// Hashicorp mux library.  This function will be called from terraform-provider-hpegl which will be adapted
// to support "legacy" provider code that use SDK v2.0 (i.e. metal, vmaas, caas on PCE) as well as newer provider
// code that uses the new Hashicorp provider "framework".
//
// Note that we will need to add the ProviderSchemaEntry() functions for the newer providers.  This means that
// registration.ServiceRegistration implementations for the newer providers that only contain ProviderSchemaEntry()
// and no SupportedResource() or SupportedDataSources().
func ProviderForMux(reg []registration.ServiceRegistration, pf ConfigureFunc) []func() tfprotov5.ProviderServer {
	providerSchema := generateProviderSchema(reg)
	providerServerList := make([]func() tfprotov5.ProviderServer, 0)
	for _, service := range reg {
		// Only create a provider if it has resources or data sources
		if service.SupportedResources() != nil || service.SupportedDataSources() != nil {
			providerServerList = append(providerServerList, generateProvider(service, pf, providerSchema))
		}
	}

	return providerServerList
}

// generateProviderSchema generates the provider schema from the service registrations.  Note that this schema
// needs to be added to each of the sub-providers.
func generateProviderSchema(reg []registration.ServiceRegistration) map[string]*schema.Schema {
	providerSchema := Schema()
	for _, service := range reg {
		if service.ProviderSchemaEntry() != nil {
			// We panic if the service.Name() key is repeated in providerSchema
			if _, ok := providerSchema[service.Name()]; ok {
				panic(fmt.Sprintf("service name %s is repeated", service.Name()))
			}
			providerSchema[service.Name()] = convertToTypeSet(service.ProviderSchemaEntry())
		}
	}

	return providerSchema
}

// generateProvider will generate a sub-provider for each service that can be used with the Hashicorp mux library.
func generateProvider(
	service registration.ServiceRegistration,
	pf ConfigureFunc,
	providerSchema map[string]*schema.Schema,
) func() tfprotov5.ProviderServer {
	p := schema.Provider{
		Schema:         providerSchema,
		ResourcesMap:   service.SupportedResources(),
		DataSourcesMap: service.SupportedDataSources(),
		// Don't use the following field, experimental
		ProviderMetaSchema: nil,
		TerraformVersion:   "",
	}

	p.ConfigureContextFunc = pf(&p) // nolint staticcheck

	return p.GRPCProvider
}
