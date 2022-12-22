package opensearch_test

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/alecthomas/assert"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
)

func TestSearchBodyMarshal(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect map[string]opensearch.Tenant
	}{
		"unmarshal tenants": {
			input: "testdata/tenants.json",
			expect: map[string]opensearch.Tenant{
				"admin_tenant": {
					TenantDescription: opensearch.TenantDescription{
						Description: "Tenant for admin user",
					},
				},
				"amazee.io internal": {
					TenantDescription: opensearch.TenantDescription{
						Description: "amazee.io internal",
					},
				},
				"drupal-example": {
					TenantDescription: opensearch.TenantDescription{
						Description: "drupal-example",
					},
				},
				"global_tenant": {
					Reserved: true,
					Static:   true,
					TenantDescription: opensearch.TenantDescription{
						Description: "Global tenant",
					},
				},
				"internaltest": {
					TenantDescription: opensearch.TenantDescription{
						Description: "internaltest",
					},
				},
				"lagoonadmin": {
					TenantDescription: opensearch.TenantDescription{
						Description: "lagoonadmin",
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			data, err := os.ReadFile(tc.input)
			if err != nil {
				tt.Fatal(err)
			}
			// check for missing fields in opensearch.Tenant
			var tenants map[string]opensearch.Tenant
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&tenants); err != nil {
				tt.Fatal(err)
			}
			assert.Equal(tt, tc.expect, tenants, "tenants")
		})
	}
}

func TestIndexPatternsUnmarshal(t *testing.T) {

	type parseIndexPatternsResponse struct {
		indexPatterns map[string]map[string]bool
		length        int
		lastUpdatedAt string
	}

	var testCases = map[string]struct {
		input  string
		expect parseIndexPatternsResponse
	}{
		"unmarshal indexpatterns no migration": {
			input: "testdata/indexpatterns.json",
			expect: parseIndexPatternsResponse{
				indexPatterns: map[string]map[string]bool{
					"global_tenant": {
						"container-logs-*":   true,
						"router-logs-*":      true,
						"lagoon-logs-*":      true,
						"application-logs-*": true,
					},
					"-152937574_admintenant": {
						"application-logs-*": true,
						"lagoon-logs-*":      true,
						"container-logs-*":   true,
						"router-logs-*":      true,
					},
					"-79010609_internaltest": {
						"router-logs-drupal9-solr-*":               true,
						"router-logs-*":                            true,
						"container-logs-react-example-*":           true,
						"container-logs-drupal9-solr-*":            true,
						"container-logs-*":                         true,
						"lagoon-logs-drupal9-solr-*":               true,
						"lagoon-logs-*":                            true,
						"application-logs-react-example-*":         true,
						"application-logs-drupal9-solr-*":          true,
						"application-logs-*":                       true,
						"router-logs-drupal-example-simple-*":      true,
						"router-logs-drupal-example-*":             true,
						"application-logs-drupal-example-simple-*": true,
						"application-logs-drupal-example-*":        true,
						"container-logs-drupal-example-simple-*":   true,
						"container-logs-drupal10-prerelease-*":     true,
						"container-logs-drupal-example-*":          true,
						"lagoon-logs-drupal10-prerelease-*":        true,
						"lagoon-logs-drupal-example-*":             true,
						"router-logs-react-example-*":              true,
						"router-logs-as-demo-*":                    true,
						"router-logs-drupal10-prerelease-*":        true,
						"container-logs-as-demo-*":                 true,
						"lagoon-logs-react-example-*":              true,
						"lagoon-logs-as-demo-*":                    true,
						"application-logs-as-demo-*":               true,
						"application-logs-drupal10-prerelease-*":   true,
						"lagoon-logs-drupal-example-simple-*":      true,
					},
					"-1014420205_lagoonadmin": {
						"application-logs-*": true,
						"router-logs-*":      true,
						"container-logs-*":   true,
						"lagoon-logs-*":      true,
					},
					"1589690574_amazeeiointernal": {
						"application-logs-*": true,
						"router-logs-*":      true,
						"container-logs-*":   true,
						"lagoon-logs-*":      true,
					},
					"698816049_drupalexample": {
						"router-logs-drupal9-base-*":               true,
						"router-logs-*":                            true,
						"container-logs-drupal-example-simple-*":   true,
						"container-logs-drupal9-base-*":            true,
						"container-logs-*":                         true,
						"lagoon-logs-drupal-example-simple-*":      true,
						"lagoon-logs-*":                            true,
						"application-logs-drupal-example-simple-*": true,
						"application-logs-drupal9-base-*":          true,
						"application-logs-*":                       true,
						"router-logs-drupal9-solr-*":               true,
						"router-logs-as-demo-*":                    true,
						"container-logs-drupal9-solr-*":            true,
						"container-logs-as-demo-*":                 true,
						"lagoon-logs-drupal9-solr-*":               true,
						"lagoon-logs-as-demo-*":                    true,
						"application-logs-as-demo-test1-*":         true,
						"application-logs-drupal9-solr-*":          true,
						"application-logs-as-demo-*":               true,
						"router-logs-drupal-example-simple-*":      true,
						"router-logs-drupal10-prerelease-*":        true,
						"router-logs-drupal9-base-gitlab-*":        true,
						"container-logs-drupal10-prerelease-*":     true,
						"container-logs-drupal9-base-gitlab-*":     true,
						"lagoon-logs-drupal10-prerelease-*":        true,
						"lagoon-logs-drupal9-base-*":               true,
						"lagoon-logs-drupal9-base-gitlab-*":        true,
						"application-logs-drupal10-prerelease-*":   true,
						"application-logs-drupal9-base-gitlab-*":   true,
						"router-logs-as-demo-test1-*":              true,
						"container-logs-as-demo-test1-*":           true,
						"lagoon-logs-as-demo-test1-*":              true,
					},
				},
				length:        76,
				lastUpdatedAt: "2022-05-18T03:52:40.628Z",
			},
		},
		"unmarshal indexpatterns post migration": {
			input: "testdata/indexpatterns2.json",
			expect: parseIndexPatternsResponse{
				indexPatterns: map[string]map[string]bool{
					"1589690574_amazeeiointernal": {
						"application-logs-*": true,
						"router-logs-*":      true,
						"container-logs-*":   true,
						"lagoon-logs-*":      true,
					},
					"698816049_drupalexample": {
						"application-logs-*":                     true,
						"router-logs-*":                          true,
						"container-logs-*":                       true,
						"lagoon-logs-*":                          true,
						"application-logs-drupal9-base-*":        true,
						"router-logs-drupal9-base-*":             true,
						"container-logs-drupal9-base-*":          true,
						"lagoon-logs-drupal9-base-*":             true,
						"application-logs-drupal10-prerelease-*": true,
						"router-logs-drupal10-prerelease-*":      true,
						"container-logs-drupal10-prerelease-*":   true,
						"lagoon-logs-drupal10-prerelease-*":      true,
						"application-logs-drupal9-solr-*":        true,
						"router-logs-drupal9-solr-*":             true,
						"container-logs-drupal9-solr-*":          true,
						"lagoon-logs-drupal9-solr-*":             true,
						"application-logs-as-demo-*":             true,
						"router-logs-as-demo-*":                  true,
						"container-logs-as-demo-*":               true,
						"lagoon-logs-as-demo-*":                  true,
						"application-logs-as-demo-test1-*":       true,
						"router-logs-as-demo-test1-*":            true,
						"container-logs-as-demo-test1-*":         true,
						"lagoon-logs-as-demo-test1-*":            true,
					},
					"-79010609_internaltest": {
						"router-logs-*":                                  true,
						"container-logs-*":                               true,
						"lagoon-logs-*":                                  true,
						"application-logs-drupal9-solr-*":                true,
						"router-logs-drupal9-solr-*":                     true,
						"container-logs-drupal9-solr-*":                  true,
						"lagoon-logs-drupal9-solr-*":                     true,
						"application-logs-react-example-*":               true,
						"router-logs-react-example-*":                    true,
						"container-logs-react-example-*":                 true,
						"lagoon-logs-react-example-*":                    true,
						"application-logs-as-demo-*":                     true,
						"router-logs-as-demo-*":                          true,
						"container-logs-as-demo-*":                       true,
						"lagoon-logs-as-demo-*":                          true,
						"application-logs-drupal10-prerelease-*":         true,
						"router-logs-drupal10-prerelease-*":              true,
						"container-logs-drupal10-prerelease-*":           true,
						"lagoon-logs-drupal10-prerelease-*":              true,
						"application-logs-drupal-example-*":              true,
						"router-logs-drupal-example-*":                   true,
						"container-logs-drupal-example-*":                true,
						"lagoon-logs-drupal-example-*":                   true,
						"application-logs-*":                             true,
						"application-logs-test6-drupal-example-simple-*": true,
						"router-logs-test6-drupal-example-simple-*":      true,
						"container-logs-test6-drupal-example-simple-*":   true,
						"lagoon-logs-test6-drupal-example-simple-*":      true,
						"application-logs-drupal9-base-*":                true,
						"router-logs-drupal9-base-*":                     true,
						"container-logs-drupal9-base-*":                  true,
						"lagoon-logs-drupal9-base-*":                     true,
						"application-logs-drupalcon-demo-*":              true,
						"router-logs-drupalcon-demo-*":                   true,
						"container-logs-drupalcon-demo-*":                true,
						"lagoon-logs-drupalcon-demo-*":                   true,
						"application-logs-lagoon-ui-*":                   true,
						"router-logs-lagoon-ui-*":                        true,
						"container-logs-lagoon-ui-*":                     true,
						"lagoon-logs-lagoon-ui-*":                        true,
					},
					"global_tenant": {
						"container-logs-*":   true,
						"router-logs-*":      true,
						"lagoon-logs-*":      true,
						"application-logs-*": true,
					},
					"-152937574_admintenant": {
						"application-logs-*": true,
						"lagoon-logs-*":      true,
						"container-logs-*":   true,
						"router-logs-*":      true,
					},
				},
				length:        152,
				lastUpdatedAt: "2022-12-02T17:18:31.585Z",
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			data, err := os.ReadFile(tc.input)
			if err != nil {
				tt.Fatal(err)
			}
			indexPatterns := map[string]map[string]bool{}
			length, lastUpdatedAt, err :=
				opensearch.ParseIndexPatterns(data, indexPatterns)
			assert.Equal(tt, length, tc.expect.length, "index pattern length")
			assert.Equal(tt, lastUpdatedAt, tc.expect.lastUpdatedAt, "last updated at")
			assert.NoError(tt, err, "parseIndexPatterns error")
			assert.Equal(tt, indexPatterns, tc.expect.indexPatterns, "index patterns")
		})
	}
}
