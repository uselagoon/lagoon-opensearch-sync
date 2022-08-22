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

func TestTenantsUnmarshal(t *testing.T) {
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
			if !reflect.DeepEqual(tc.expect, tenants) {
				tt.Fatalf("expected:\n%s\ngot\n%s\n",
					spew.Sdump(tc.expect), spew.Sdump(tenants))
			}
		})
	}
}
