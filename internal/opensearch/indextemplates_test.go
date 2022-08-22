package opensearch_test

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
)

func TestIndexTemplatesUnmarshal(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect map[string]opensearch.IndexTemplate
	}{
		"unmarshal index templates": {
			input: "testdata/indextemplates.json",
			expect: map[string]opensearch.IndexTemplate{
				"routerlogs": {
					Name: "routerlogs",
					IndexTemplateDefinition: opensearch.IndexTemplateDefinition{
						ComposedOf:    []string{},
						IndexPatterns: []string{"router-logs-*"},
						Template: opensearch.Template{
							Mappings: &opensearch.Mappings{
								DynamicTemplates: []map[string]opensearch.DynamicTemplate{
									{
										"remote_addr": {
											Match:            "remote_addr",
											MatchMappingType: "string",
											Mapping: opensearch.Mapping{
												Type:            "ip",
												IgnoreMalformed: true,
											},
										},
									},
									{
										"true-client-ip": {
											Match:            "true-client-ip",
											MatchMappingType: "string",
											Mapping: opensearch.Mapping{
												Type:            "ip",
												IgnoreMalformed: true,
											},
										},
									},
								},
							},
						},
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
			// check for missing fields
			var its opensearch.IndexTemplatesSlice
			decoder := json.NewDecoder(bytes.NewReader(data))
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&its); err != nil {
				tt.Fatal(err)
			}
			itm := opensearch.IndexTemplatesMap(&its)
			if !reflect.DeepEqual(tc.expect, itm) {
				tt.Fatalf("expected:\n%s\ngot\n%s\n",
					spew.Sdump(tc.expect), spew.Sdump(itm))
			}
		})
	}
}
