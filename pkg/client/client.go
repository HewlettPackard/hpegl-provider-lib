// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package client

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Initialisation interface, service Client creation code will have to satisfy this interface
// The hpegl provider will iterate over a slice of these to initialise service clients
type Initialisation interface {
	// NewClient is run by hpegl to initialise the service client
	NewClient(r *schema.ResourceData) (interface{}, error)

	// ServiceName is used by hpegl, it returns the key to be used for the client returned by NewClient
	// in the map[string]interface{} passed-down to provider code by terraform
	ServiceName() string
}

// GetServiceSettingsMap helper function for use by client code in NewClient instances
// This function takes the schema.ResourceData passed in to NewClient, gets the []interface{}
// at the key passed in which we know will have just one element,gets that element and
// converts to map[string]interface{}.  This map holds the settings for the service.
// If the block hasn't been set we return an error.
func GetServiceSettingsMap(key string, r *schema.ResourceData) (map[string]interface{}, error) {
	l := r.Get(key).([]interface{})
	if len(l) == 0 {
		return nil, fmt.Errorf("service %s block not defined in hpegl stanza", key)
	}

	return l[0].(map[string]interface{}), nil
}
