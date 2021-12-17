// package atf or acceptance-test-framework consists of helper files to parse,
// validate and run acceptance test with vmaas specified terraform acceptance
// test case format
package atf

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/spf13/viper"
)

// accConfig struct holds terraform configuration of a resource/data source
// along with parsed validations
type accConfig struct {
	// terraform configuration
	config      string
	validations []validation
}

// validation holds parsed validation
type validation struct {
	// isJSON denotes whether this validate against Get API or
	// terraform state file
	isJSON bool
	// json field key or state file attribute key
	key string
	// value is a singleton data which we will validate against
	value interface{}
}

// reader encapsulate all the reading and parsing test files
type reader struct {
	t *testing.T
	// denotes whether this is resource or data source
	isResource bool
	// resource/data source name
	name string
	// expectError
	expectError *regexp.Regexp
	// all the parsed variables stored in vars map
	vars map[string]interface{}
}

// newReader retun new reader instance
func newReader(t *testing.T, isResource bool, name string) *reader {
	return &reader{
		t:          t,
		isResource: isResource,
		name:       name,
		vars:       make(map[string]interface{}),
	}
}

func (r *reader) fatalf(format string, v ...interface{}) {
	r.t.Fatalf("[acc-test] test case for resource "+r.name+" failed. "+format, v...)
}

func (r *reader) skipf(format string, v ...interface{}) {
	r.t.Skipf("[acc-test] test case for resource "+r.name+" is skipped. "+format, v...)
}

// parse variable decleration in 'vars:' field
func (r *reader) readVars(v *viper.Viper) {
	vars, ok := v.Get("vars").(map[string]interface{})
	if !ok {
		return
	}

	for key, val := range vars {
		r.vars[key] = parseMeta(fmt.Sprint(val))
	}
}

// getViperConfig read a file using viper. The filepath to the test will determin by following logic
// ${TF_ACC_TEST_PATH}/(resource/data-sources)/<name of the resource/data source without hpegl_service_name tag>
func (r *reader) getViperConfig(version string) *viper.Viper {
	tfName := getLocalName(r.name)
	if path := os.Getenv(AccTestPathKey); path != "" {
		accTestPath = path
	}
	var postfix string
	if version != "" {
		postfix = fmt.Sprintf("-%s", version)
	}
	v := viper.New()
	v.SetConfigFile(fmt.Sprintf("%s/%s/%s%s.yaml", accTestPath, getTag(r.isResource), tfName, postfix))
	err := v.ReadInConfig()
	if err != nil {
		r.skipf("error while reading config, %v", err)
	}

	return v
}

// replaceVar replaces with varaibles with thier definition
func (r *reader) replaceVar(vars map[string]interface{}, config string) string {
	exp := `\$\([a-zA-Z_0-9]+\)`
	reg := regexp.MustCompile(exp)

	// loop through all the variables on the configuration and replace with appropriate
	// values
	matches := reg.FindAllString(config, -1)
	for _, m := range matches {
		varName, ok := vars[m[2:len(m)-1]]
		// check variable definition exists
		if !ok {
			r.fatalf("variable definition for %s not found", varName)
		}
		config = strings.Replace(config, m, fmt.Sprint(varName), 1)
	}

	return config
}

// parseMeta currently supports generating of random string. But in future this
// can enhance to support random string or any other data types.
// %rand_int will generate random number under randMaxLimit
// %rand_int{a,b} will generate random number in between a and b.
func parseMeta(data string) string {
	exp := `%(rand_int)(\{[0-9]+,[0-9]+\})?`
	reg := regexp.MustCompile(exp)

	matches := reg.FindAllString(data, -1)
	var randInt int
	r := newRand()
	for _, m := range matches {
		offReg := regexp.MustCompile(`[0-9]+,[0-9]`)
		numStr := offReg.FindString(m)
		if numStr != "" {
			intSplit := strings.Split(numStr, ",")
			n1 := toInt(intSplit[0])
			n2 := toInt(intSplit[1])
			randInt = r.Intn(n2-n1) + n1
		} else {
			randInt = r.Intn(randMaxLimit)
		}
		data = strings.Replace(data, m, strconv.Itoa(randInt), 1)
	}

	return data
}

// parseValidations take care parsing the validation and populate validation struct
// currently we are only supporting json/tf validation. Here we can't use any
// meta function (such as len() or greatedThan() etc) on validation, but can be included
// in near future
func (r *reader) parseValidations(vip *viper.Viper, i int) []validation {
	vls, ok := vip.Get(fmt.Sprintf("%s.%d.validations", accKey, i)).(map[interface{}]interface{})
	if !ok {
		return nil
	}
	m := make([]validation, 0, len(vls))
	for k, v := range vls {
		kStr := k.(string)
		kSplit := strings.Split(kStr, ".")
		if len(kSplit) > 1 && (kSplit[0] == jsonKey || kSplit[0] == tfKey) {
			isJSON := false
			if kSplit[0] == jsonKey {
				isJSON = true
			}

			m = append(m, validation{
				isJSON: isJSON,
				key:    kStr[len(kSplit[0])+1:],
				value:  v,
			})
		} else {
			r.fatalf("invalid validation format. validation format should be '[json|tf].key1.key2....keyn: value'")
		}
	}

	return m
}

// parseExpectErr converts string parsed by user to regex object
func (r *reader) parseExpectErr(v *viper.Viper, i int) {
	if expectErrStr := v.GetString(path(accKey, i, "expect_error")); expectErrStr != "" {
		var err error
		r.expectError, err = regexp.Compile(expectErrStr)
		if err != nil {
			r.fatalf("error while compiling regex %s, got error %v", expectErrStr, err)
		}
	}
}

// parseConfig populates terraform configuration and parse to accConfig
func (r *reader) parseConfig(v *viper.Viper) []accConfig {
	tfKey := getLocalName(r.name)

	testCases := v.Get(accKey).([]interface{})
	configs := make([]accConfig, len(testCases))
	// loop each test case in a file and append following line
	// <resource/data> <resource/datasource name> tf_<res/ds nam> {
	// At the end of loop we will have a complete configuration along
	// with provider config as well
	for i := range testCases {
		tfConfig := r.replaceVar(r.vars, v.GetString(path(accKey, i, "config")))
		configs[i].config = fmt.Sprintf(`
		%s
		%s "%s" "tf_%s" {
			%s
		}
		`,
			providerStanza, getType(r.isResource), r.name, tfKey, tfConfig,
		)
		// parse expect error and validation as well
		r.parseExpectErr(v, i)
		configs[i].validations = r.parseValidations(v, i)
	}

	return configs
}

// getTestCases populate TestSteps
func (r *reader) getTestCases(version string, getAPI GetAPIFunc) []resource.TestStep {
	v := r.getViperConfig(version)
	// ignore field in test suite can be used to skip specific test file altogather without
	// actually deleting it
	if v.GetBool("ignore") {
		r.t.Skip("ignoring tests for resource ", r.name)
	}

	r.readVars(v)
	configs := r.parseConfig(v)
	testSteps := make([]resource.TestStep, 0, len(configs))

	tag := ""
	if !r.isResource {
		tag = "data."
	}

	for _, c := range configs {
		testSteps = append(testSteps, resource.TestStep{
			Config: c.config,
			Check: resource.ComposeTestCheckFunc(
				validateResource(
					fmt.Sprintf("%s%s.tf_%s", tag, r.name, getLocalName(r.name)),
					c.validations,
					getAPI,
				),
			),
			ExpectError: r.expectError,
		})
	}

	return testSteps
}
