// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package atf

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// GetAPIFunc accepts terraform states attribures as params and
// expects response and error as return values
type GetAPIFunc func(attr map[string]string) (interface{}, error)

type (
	ValidateResourceDestroyFunc func(attr map[string]string) error
	Acc                         struct {
		// in PreCheck each service team should provides validation which done
		// beforehand (before running the configurations). For example valildate
		// paritcular env is set or not.
		PreCheck func(t *testing.T)
		// Providers used in testing
		Providers map[string]*schema.Provider
		// GetAPI used to check the truth and do the validation purpose after the
		// configuration applied on the test. GetAPI expected to be an API call
		// and get the specific resource from the infrastructure.
		GetAPI GetAPIFunc
		// Name of the resource/Data source
		ResourceName string
		// Version indicates the version of test case. This should be unique across
		// all the test cases for a specific resource/data source. Version helps to
		// write different and independent test cases of same resource/ data source
		Version                 string
		ValidateResourceDestroy ValidateResourceDestroyFunc
	}
)

// RunResourcePlanTest to run resource plan only test case. This will take first
// config from specific resource.
func (a *Acc) RunResourcePlanTest(t *testing.T) {
	t.Helper()
	checkSkip(t)
	a.runPlanTest(t, true)
}

// RunDataSourceTests to run data source plan only test case. This will take first
// config from specific data source
func (a *Acc) RunDataSourceTests(t *testing.T) {
	t.Helper()
	checkSkip(t)
	r := newReader(t, false, a.ResourceName)
	testSteps := r.getTestCases(a.Version, a.GetAPI)

	resource.ParallelTest(t, resource.TestCase{
		IsUnitTest: false,
		PreCheck:   func() { a.PreCheck(t) },
		Providers:  a.Providers,
		Steps:      testSteps,
	})
}

// RunResourceTests creates test cases and run tests which includes create/update/delete/read
func (a *Acc) RunResourceTests(t *testing.T) {
	t.Helper()
	checkSkip(t)
	// skip resource create/update/delete operation in short mode
	if testing.Short() {
		t.Skipf("Skipping %s resource testing in short mode", a.ResourceName)
	}

	// populate test cases
	r := newReader(t, true, a.ResourceName)
	testSteps := r.getTestCases(a.Version, a.GetAPI)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { a.PreCheck(t) },
		Providers: a.Providers,
		CheckDestroy: resource.ComposeTestCheckFunc(
			a.checkResourceDestroy,
		),
		Steps: testSteps,
	})
}

// checkResourceDestroy checks resource destroy conditions. This will check the resource actually exists
// and then run user validation if any
func (a *Acc) checkResourceDestroy(s *terraform.State) error {
	rs, ok := s.RootModule().Resources[fmt.Sprintf("%s.tf_%s", a.ResourceName, getLocalName(a.ResourceName))]
	if !ok {
		return fmt.Errorf("[Check Destroy] resource %s not found", a.ResourceName)
	}
	// skip destroy validation if developer doesn't specify
	if a.ValidateResourceDestroy == nil {
		return nil
	}

	return a.ValidateResourceDestroy(rs.Primary.Attributes)
}

// runs plan test for resource or data source. only first config from test case
// will considered on plan test
func (a *Acc) runPlanTest(t *testing.T, isResource bool) {
	t.Helper()
	// populate test cases
	r := newReader(t, isResource, a.ResourceName)
	testSteps := r.getTestCases(a.Version, a.GetAPI)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { a.PreCheck(t) },
		Providers: a.Providers,
		Steps: []resource.TestStep{
			{
				Config:             testSteps[0].Config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check:              testSteps[0].Check,
			},
		},
	})
}

// checkSkip ensure to exclude acceptance test while running UTs.
func checkSkip(t *testing.T) {
	t.Helper()
	if strings.ToLower(os.Getenv("TF_ACC")) != "true" && os.Getenv("TF_ACC") != "1" {
		t.Skip("acceptance test is skipped since TF_ACC is not set")
	}
}
