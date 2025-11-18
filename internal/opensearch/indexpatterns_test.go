package opensearch_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/alecthomas/assert/v2"
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

func TestParseIndexPatterns(t *testing.T) {
	type parseIndexPatternsResponse struct {
		indexPatterns map[string]map[string][]string
		length        int
		lastSortField []int
	}

	var testCases = map[string]struct {
		input  string
		expect parseIndexPatternsResponse
	}{
		"unmarshal indexpatterns no migration": {
			input: "testdata/indexpatterns.json",
			expect: parseIndexPatternsResponse{
				indexPatterns: map[string]map[string][]string{
					"global_tenant": {
						"container-logs-*":   []string{"container-logs-*"},
						"router-logs-*":      []string{"router-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
						"application-logs-*": []string{"application-logs-*"},
					},
					"-152937574_admintenant": {
						"application-logs-*": []string{"6d21de70-dbc1-11ec-b2f3-8b83afd03d97"},
						"lagoon-logs-*":      []string{"3828f1c0-e6d6-11ec-b2f3-8b83afd03d97"},
						"container-logs-*":   []string{"43e0cae0-d661-11ec-99f5-f1a2c20fac86"},
						"router-logs-*":      []string{"5060dda0-d661-11ec-99f5-f1a2c20fac86"},
					},
					"-79010609_internaltest": {
						"router-logs-drupal9-solr-*":               []string{"router-logs-drupal9-solr-*"},
						"router-logs-*":                            []string{"router-logs-*"},
						"container-logs-react-example-*":           []string{"container-logs-react-example-*"},
						"container-logs-drupal9-solr-*":            []string{"container-logs-drupal9-solr-*"},
						"container-logs-*":                         []string{"container-logs-*"},
						"lagoon-logs-drupal9-solr-*":               []string{"lagoon-logs-drupal9-solr-*"},
						"lagoon-logs-*":                            []string{"lagoon-logs-*"},
						"application-logs-react-example-*":         []string{"application-logs-react-example-*"},
						"application-logs-drupal9-solr-*":          []string{"application-logs-drupal9-solr-*"},
						"application-logs-*":                       []string{"application-logs-*"},
						"router-logs-drupal-example-simple-*":      []string{"router-logs-drupal-example-simple-*"},
						"router-logs-drupal-example-*":             []string{"router-logs-drupal-example-*"},
						"application-logs-drupal-example-simple-*": []string{"application-logs-drupal-example-simple-*"},
						"application-logs-drupal-example-*":        []string{"application-logs-drupal-example-*"},
						"container-logs-drupal-example-simple-*":   []string{"container-logs-drupal-example-simple-*"},
						"container-logs-drupal10-prerelease-*":     []string{"container-logs-drupal10-prerelease-*"},
						"container-logs-drupal-example-*":          []string{"container-logs-drupal-example-*"},
						"lagoon-logs-drupal10-prerelease-*":        []string{"lagoon-logs-drupal10-prerelease-*"},
						"lagoon-logs-drupal-example-*":             []string{"lagoon-logs-drupal-example-*"},
						"router-logs-react-example-*":              []string{"router-logs-react-example-*"},
						"router-logs-as-demo-*":                    []string{"router-logs-as-demo-*"},
						"router-logs-drupal10-prerelease-*":        []string{"router-logs-drupal10-prerelease-*"},
						"container-logs-as-demo-*":                 []string{"container-logs-as-demo-*"},
						"lagoon-logs-react-example-*":              []string{"lagoon-logs-react-example-*"},
						"lagoon-logs-as-demo-*":                    []string{"lagoon-logs-as-demo-*"},
						"application-logs-as-demo-*":               []string{"application-logs-as-demo-*"},
						"application-logs-drupal10-prerelease-*":   []string{"application-logs-drupal10-prerelease-*"},
						"lagoon-logs-drupal-example-simple-*":      []string{"lagoon-logs-drupal-example-simple-*"},
					},
					"-1014420205_lagoonadmin": {
						"application-logs-*": []string{"application-logs-*"},
						"router-logs-*":      []string{"router-logs-*"},
						"container-logs-*":   []string{"container-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
					},
					"1589690574_amazeeiointernal": {
						"application-logs-*": []string{"application-logs-*"},
						"router-logs-*":      []string{"router-logs-*"},
						"container-logs-*":   []string{"container-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
					},
					"698816049_drupalexample": {
						"router-logs-drupal9-base-*":               []string{"router-logs-drupal9-base-*"},
						"router-logs-*":                            []string{"router-logs-*"},
						"container-logs-drupal-example-simple-*":   []string{"container-logs-drupal-example-simple-*"},
						"container-logs-drupal9-base-*":            []string{"container-logs-drupal9-base-*"},
						"container-logs-*":                         []string{"container-logs-*"},
						"lagoon-logs-drupal-example-simple-*":      []string{"lagoon-logs-drupal-example-simple-*"},
						"lagoon-logs-*":                            []string{"lagoon-logs-*"},
						"application-logs-drupal-example-simple-*": []string{"application-logs-drupal-example-simple-*"},
						"application-logs-drupal9-base-*":          []string{"application-logs-drupal9-base-*"},
						"application-logs-*":                       []string{"application-logs-*"},
						"router-logs-drupal9-solr-*":               []string{"router-logs-drupal9-solr-*"},
						"router-logs-as-demo-*":                    []string{"router-logs-as-demo-*"},
						"container-logs-drupal9-solr-*":            []string{"container-logs-drupal9-solr-*"},
						"container-logs-as-demo-*":                 []string{"container-logs-as-demo-*"},
						"lagoon-logs-drupal9-solr-*":               []string{"lagoon-logs-drupal9-solr-*"},
						"lagoon-logs-as-demo-*":                    []string{"lagoon-logs-as-demo-*"},
						"application-logs-as-demo-test1-*":         []string{"application-logs-as-demo-test1-*"},
						"application-logs-drupal9-solr-*":          []string{"application-logs-drupal9-solr-*"},
						"application-logs-as-demo-*":               []string{"application-logs-as-demo-*"},
						"router-logs-drupal-example-simple-*":      []string{"router-logs-drupal-example-simple-*"},
						"router-logs-drupal10-prerelease-*":        []string{"router-logs-drupal10-prerelease-*"},
						"router-logs-drupal9-base-gitlab-*":        []string{"router-logs-drupal9-base-gitlab-*"},
						"container-logs-drupal10-prerelease-*":     []string{"container-logs-drupal10-prerelease-*"},
						"container-logs-drupal9-base-gitlab-*":     []string{"container-logs-drupal9-base-gitlab-*"},
						"lagoon-logs-drupal10-prerelease-*":        []string{"lagoon-logs-drupal10-prerelease-*"},
						"lagoon-logs-drupal9-base-*":               []string{"lagoon-logs-drupal9-base-*"},
						"lagoon-logs-drupal9-base-gitlab-*":        []string{"lagoon-logs-drupal9-base-gitlab-*"},
						"application-logs-drupal10-prerelease-*":   []string{"application-logs-drupal10-prerelease-*"},
						"application-logs-drupal9-base-gitlab-*":   []string{"application-logs-drupal9-base-gitlab-*"},
						"router-logs-as-demo-test1-*":              []string{"router-logs-as-demo-test1-*"},
						"container-logs-as-demo-test1-*":           []string{"container-logs-as-demo-test1-*"},
						"lagoon-logs-as-demo-test1-*":              []string{"lagoon-logs-as-demo-test1-*"},
					},
				},
				length:        76,
				lastSortField: []int{1652845960628},
			},
		},
		"unmarshal indexpatterns post migration": {
			input: "testdata/indexpatterns2.json",
			expect: parseIndexPatternsResponse{
				indexPatterns: map[string]map[string][]string{
					"1589690574_amazeeiointernal": {
						"application-logs-*": []string{"application-logs-*"},
						"router-logs-*":      []string{"router-logs-*"},
						"container-logs-*":   []string{"container-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
					},
					"698816049_drupalexample": {
						"application-logs-*":                     []string{"application-logs-*"},
						"router-logs-*":                          []string{"router-logs-*"},
						"container-logs-*":                       []string{"container-logs-*"},
						"lagoon-logs-*":                          []string{"lagoon-logs-*"},
						"application-logs-drupal9-base-*":        []string{"application-logs-drupal9-base-*"},
						"router-logs-drupal9-base-*":             []string{"router-logs-drupal9-base-*"},
						"container-logs-drupal9-base-*":          []string{"container-logs-drupal9-base-*"},
						"lagoon-logs-drupal9-base-*":             []string{"lagoon-logs-drupal9-base-*"},
						"application-logs-drupal10-prerelease-*": []string{"application-logs-drupal10-prerelease-*"},
						"router-logs-drupal10-prerelease-*":      []string{"router-logs-drupal10-prerelease-*"},
						"container-logs-drupal10-prerelease-*":   []string{"container-logs-drupal10-prerelease-*"},
						"lagoon-logs-drupal10-prerelease-*":      []string{"lagoon-logs-drupal10-prerelease-*"},
						"application-logs-drupal9-solr-*":        []string{"application-logs-drupal9-solr-*"},
						"router-logs-drupal9-solr-*":             []string{"router-logs-drupal9-solr-*"},
						"container-logs-drupal9-solr-*":          []string{"container-logs-drupal9-solr-*"},
						"lagoon-logs-drupal9-solr-*":             []string{"lagoon-logs-drupal9-solr-*"},
						"application-logs-as-demo-*":             []string{"application-logs-as-demo-*"},
						"router-logs-as-demo-*":                  []string{"router-logs-as-demo-*"},
						"container-logs-as-demo-*":               []string{"container-logs-as-demo-*"},
						"lagoon-logs-as-demo-*":                  []string{"lagoon-logs-as-demo-*"},
						"application-logs-as-demo-test1-*":       []string{"application-logs-as-demo-test1-*"},
						"router-logs-as-demo-test1-*":            []string{"router-logs-as-demo-test1-*"},
						"container-logs-as-demo-test1-*":         []string{"container-logs-as-demo-test1-*"},
						"lagoon-logs-as-demo-test1-*":            []string{"lagoon-logs-as-demo-test1-*"},
					},
					"-79010609_internaltest": {
						"router-logs-*":                                  []string{"router-logs-*"},
						"container-logs-*":                               []string{"container-logs-*"},
						"lagoon-logs-*":                                  []string{"lagoon-logs-*"},
						"application-logs-drupal9-solr-*":                []string{"application-logs-drupal9-solr-*"},
						"router-logs-drupal9-solr-*":                     []string{"router-logs-drupal9-solr-*"},
						"container-logs-drupal9-solr-*":                  []string{"container-logs-drupal9-solr-*"},
						"lagoon-logs-drupal9-solr-*":                     []string{"lagoon-logs-drupal9-solr-*"},
						"application-logs-react-example-*":               []string{"application-logs-react-example-*"},
						"router-logs-react-example-*":                    []string{"router-logs-react-example-*"},
						"container-logs-react-example-*":                 []string{"container-logs-react-example-*"},
						"lagoon-logs-react-example-*":                    []string{"lagoon-logs-react-example-*"},
						"application-logs-as-demo-*":                     []string{"application-logs-as-demo-*"},
						"router-logs-as-demo-*":                          []string{"router-logs-as-demo-*"},
						"container-logs-as-demo-*":                       []string{"container-logs-as-demo-*"},
						"lagoon-logs-as-demo-*":                          []string{"lagoon-logs-as-demo-*"},
						"application-logs-drupal10-prerelease-*":         []string{"application-logs-drupal10-prerelease-*"},
						"router-logs-drupal10-prerelease-*":              []string{"router-logs-drupal10-prerelease-*"},
						"container-logs-drupal10-prerelease-*":           []string{"container-logs-drupal10-prerelease-*"},
						"lagoon-logs-drupal10-prerelease-*":              []string{"lagoon-logs-drupal10-prerelease-*"},
						"application-logs-drupal-example-*":              []string{"application-logs-drupal-example-*"},
						"router-logs-drupal-example-*":                   []string{"router-logs-drupal-example-*"},
						"container-logs-drupal-example-*":                []string{"container-logs-drupal-example-*"},
						"lagoon-logs-drupal-example-*":                   []string{"lagoon-logs-drupal-example-*"},
						"application-logs-*":                             []string{"application-logs-*"},
						"application-logs-test6-drupal-example-simple-*": []string{"application-logs-test6-drupal-example-simple-*"},
						"router-logs-test6-drupal-example-simple-*":      []string{"router-logs-test6-drupal-example-simple-*"},
						"container-logs-test6-drupal-example-simple-*":   []string{"container-logs-test6-drupal-example-simple-*"},
						"lagoon-logs-test6-drupal-example-simple-*":      []string{"lagoon-logs-test6-drupal-example-simple-*"},
						"application-logs-drupal9-base-*":                []string{"application-logs-drupal9-base-*"},
						"router-logs-drupal9-base-*":                     []string{"router-logs-drupal9-base-*"},
						"container-logs-drupal9-base-*":                  []string{"container-logs-drupal9-base-*"},
						"lagoon-logs-drupal9-base-*":                     []string{"lagoon-logs-drupal9-base-*"},
						"application-logs-drupalcon-demo-*":              []string{"application-logs-drupalcon-demo-*"},
						"router-logs-drupalcon-demo-*":                   []string{"router-logs-drupalcon-demo-*"},
						"container-logs-drupalcon-demo-*":                []string{"container-logs-drupalcon-demo-*"},
						"lagoon-logs-drupalcon-demo-*":                   []string{"lagoon-logs-drupalcon-demo-*"},
						"application-logs-lagoon-ui-*":                   []string{"application-logs-lagoon-ui-*"},
						"router-logs-lagoon-ui-*":                        []string{"router-logs-lagoon-ui-*"},
						"container-logs-lagoon-ui-*":                     []string{"container-logs-lagoon-ui-*"},
						"lagoon-logs-lagoon-ui-*":                        []string{"lagoon-logs-lagoon-ui-*"},
					},
					"global_tenant": {
						"container-logs-*":   []string{"container-logs-*"},
						"router-logs-*":      []string{"router-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
						"application-logs-*": []string{"application-logs-*"},
					},
					"-152937574_admintenant": {
						"application-logs-*": []string{"application-logs-*"},
						"lagoon-logs-*":      []string{"3828f1c0-e6d6-11ec-b2f3-8b83afd03d97"},
						"container-logs-*":   []string{"43e0cae0-d661-11ec-99f5-f1a2c20fac86"},
						"router-logs-*":      []string{"5060dda0-d661-11ec-99f5-f1a2c20fac86"},
					},
				},
				length:        152,
				lastSortField: []int{1670001511585},
			},
		},
		"handle multiple kibana indices": {
			input: "testdata/indexpatterns3.json",
			expect: parseIndexPatternsResponse{
				indexPatterns: map[string]map[string][]string{
					"global_tenant": {
						"router-logs-*":      []string{"router-logs-*"},
						"lagoon-logs-*":      []string{"lagoon-logs-*"},
						"application-logs-*": []string{"9b7da830-d427-11ed-b326-3348256dd0e8"},
					},
					"-152937574_admintenant": {
						"lagoon-logs-*": []string{"lagoon-logs-*"},
					},
				},
				length:        9,
				lastSortField: []int{1683014064736},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			data, err := os.ReadFile(tc.input)
			if err != nil {
				tt.Fatal(err)
			}

			var s opensearch.SearchResult
			assert.NoError(tt, json.Unmarshal(data, &s), "json unmarshal")

			assert.Equal(tt, len(s.Hits.Hits), tc.expect.length, "index pattern length")
			assert.Equal(
				tt,
				s.Hits.Hits[len(s.Hits.Hits)-1].Sort,
				tc.expect.lastSortField,
				"last sort field")

			indexPatterns, err := opensearch.ParseIndexPatterns(s.Hits.Hits)
			assert.NoError(tt, err, "parseIndexPatterns error")
			assert.Equal(tt, indexPatterns, tc.expect.indexPatterns, "index patterns")
		})
	}
}

func TestSplitMigration(t *testing.T) {
	var testCases = map[string]struct {
		input  []string
		expect map[string]map[string][]string
	}{
		"migration split across two pages": {
			input: []string{
				"testdata/indexpatternsSplitMigration0.json",
				"testdata/indexpatternsSplitMigration1.json",
			},
			expect: map[string]map[string][]string{
				"1589690574_amazeeiointernal": {
					"application-logs-*": {"application-logs-*"},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			// configure router to return the input data in sequence
			respIdx := 0
			mux := http.NewServeMux()
			mux.HandleFunc("/.kibana*/_search",
				func(w http.ResponseWriter, r *http.Request) {
					data, err := os.ReadFile(tc.input[respIdx])
					assert.NoError(tt, err, name)
					_, err = io.Copy(w, bytes.NewBuffer(data))
					assert.NoError(tt, err, name)
					respIdx++
				})
			ts := httptest.NewServer(mux)
			// init opensearch client with a searchSize of 3 to match the testdata
			client, err := opensearch.NewTestClient(ts.URL, 3)
			assert.NoError(tt, err, name)
			// run test and check results
			indexPatterns, err := client.IndexPatterns(t.Context())
			assert.NoError(tt, err, name)
			assert.Equal(tt, tc.expect, indexPatterns, name)
		})
	}
}

func TestIdenticalSortValue(t *testing.T) {
	var testCases = map[string]struct {
		input []struct {
			expectSearchAfter []int
			responseDataPath  string
		}
		expect map[string]map[string][]string
	}{
		"identical sort values split across pages": {
			input: []struct {
				expectSearchAfter []int
				responseDataPath  string
			}{
				{
					responseDataPath: "testdata/indexpatternsIdenticalTimestamps0.json",
				},
				{
					expectSearchAfter: []int{1735263963223},
					responseDataPath:  "testdata/indexpatternsIdenticalTimestamps1.json",
				},
			},
			expect: map[string]map[string][]string{
				"-1878760932_bartenant": {
					"router-logs-foosite-au-_-*":      {"router-logs-foosite-au-_-*"},
					"application-logs-foosite-au-_-*": {"application-logs-foosite-au-_-*"},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			// configure router to return the input data in sequence
			respIdx := 0
			mux := http.NewServeMux()
			mux.HandleFunc("/.kibana*/_search",
				func(w http.ResponseWriter, r *http.Request) {
					// inspect request body
					body, err := io.ReadAll(r.Body)
					assert.NoError(tt, err, name)
					var query opensearch.SearchBody
					err = json.Unmarshal(body, &query)
					assert.NoError(tt, err, name)
					if !slices.Equal(tc.input[respIdx].expectSearchAfter, query.SearchAfter) {
						http.Error(w,
							fmt.Sprintf("unexpected searchAfter value: expected %v, got %v",
								tc.input[respIdx].expectSearchAfter,
								query.SearchAfter),
							http.StatusBadRequest)
						return
					}
					// construct response
					data, err := os.ReadFile(tc.input[respIdx].responseDataPath)
					assert.NoError(tt, err, name)
					_, err = io.Copy(w, bytes.NewBuffer(data))
					assert.NoError(tt, err, name)
					respIdx++
				})
			ts := httptest.NewServer(mux)
			// init opensearch client with a searchSize of 3 to match the testdata
			client, err := opensearch.NewTestClient(ts.URL, 3)
			assert.NoError(tt, err, name)
			// run test and check results
			indexPatterns, err := client.IndexPatterns(t.Context())
			assert.NoError(tt, err, name)
			assert.Equal(tt, tc.expect, indexPatterns, name)
		})
	}
}
