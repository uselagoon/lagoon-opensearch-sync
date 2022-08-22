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

func TestRolesmappingUnmarshal(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect map[string]opensearch.RoleMapping
	}{
		"unmarshal rolesmapping": {
			input: "testdata/rolesmapping.json",
			expect: map[string]opensearch.RoleMapping{
				"all_access": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"admin", "platform-owner"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"amazee.io internal": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"amazee.io internal"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"drupal-example": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"drupal-example"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"internaltest": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"internaltest"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"kibana_server": {
					Reserved: true,
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{},
						Hosts:           []string{},
						Users:           []string{"kibanaserver"},
					},
				},
				"kibana_user": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{},
						Hosts:           []string{},
						Users:           []string{"*"},
					},
				},
				"lagoonadmin": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"lagoonadmin"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p11": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p11"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p12": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p12"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p13": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p13"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p14": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p14"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p15": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p15"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p16": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p16"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p17": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p17"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p18": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p18"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p19": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p19"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p20": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p20"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p21": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p21"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p22": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p22"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p23": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p23"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p24": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p24"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p25": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p25"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p27": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p27"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p29": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p29"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p30": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p30"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p31": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p31"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p33": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p33"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p34": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p34"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p35": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p35"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p36": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p36"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p37": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p37"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p38": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p38"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p39": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p39"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p40": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p40"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"p41": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{"p41"},
						Hosts:           []string{},
						Users:           []string{},
					},
				},
				"prometheus_exporter": {
					Reserved: true,
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						AndBackendRoles: []string{},
						BackendRoles:    []string{},
						Hosts:           []string{},
						Users:           []string{"prometheusexporter"},
					},
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			jb, err := os.ReadFile(tc.input)
			if err != nil {
				tt.Fatal(err)
			}
			// check for missing fields in RoleMapping
			var rm map[string]opensearch.RoleMapping
			decoder := json.NewDecoder(bytes.NewReader(jb))
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&rm); err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expect, rm) {
				tt.Fatalf("expected:\n%s\ngot\n%s\n",
					spew.Sdump(tc.expect), spew.Sdump(rm))
			}
		})
	}
}
