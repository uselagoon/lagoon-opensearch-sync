package sync_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
)

func TestFilterRolesMapping(t *testing.T) {
	type filterRolesMappingInput struct {
		rolesMappings map[string]opensearch.RoleMapping
		roles         map[string]opensearch.Role
	}
	var testCases = map[string]struct {
		input  filterRolesMappingInput
		expect map[string]opensearch.RoleMapping
	}{
		"filter hidden, reserved and custom roles mapping": {
			input: filterRolesMappingInput{
				rolesMappings: map[string]opensearch.RoleMapping{
					"hidden-role": {
						Hidden: true,
						RoleMappingPermissions: opensearch.RoleMappingPermissions{
							Users: []string{"test-user"},
						},
					},
					"reserved-role": {
						Reserved: true,
						RoleMappingPermissions: opensearch.RoleMappingPermissions{
							Users: []string{"test-user"},
						},
					},
					"custom-role": {
						RoleMappingPermissions: opensearch.RoleMappingPermissions{
							Users: []string{"test-user"},
						},
					},
					"drupal-role": {
						RoleMappingPermissions: opensearch.RoleMappingPermissions{
							Users: []string{"test-user"},
						},
					},
				},
				roles: map[string]opensearch.Role{
					"hidden-role": {
						Static: true,
						RolePermissions: opensearch.RolePermissions{
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-pets-com-_-.+/",
									},
									MaskedFields: []string{},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_read"},
									TenantPatterns: []string{"global_tenant"},
								},
							},
						},
					},
					"reserved-role": {
						Reserved: true,
						RolePermissions: opensearch.RolePermissions{
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-pets-com-_-.+/",
									},
									MaskedFields: []string{},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_read"},
									TenantPatterns: []string{"global_tenant"},
								},
							},
						},
					},
					"custom-role": {
						RolePermissions: opensearch.RolePermissions{
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-pets-com-_-.+/",
									},
									MaskedFields: []string{},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_read"},
									TenantPatterns: []string{"global_tenant"},
								},
							},
						},
					},
					"drupal-role": {
						RolePermissions: opensearch.RolePermissions{
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-pets-com-_-.+/",
									},
									MaskedFields: []string{},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_read"},
									TenantPatterns: []string{"global_tenant"},
								},
							},
						},
					},
				},
			},
			expect: map[string]opensearch.RoleMapping{
				"drupal-role": {
					RoleMappingPermissions: opensearch.RoleMappingPermissions{
						Users: []string{"test-user"},
					}},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			filteredRolesMappings := sync.FilterRolesMapping(tc.input.rolesMappings, tc.input.roles)
			assert.Equal(tt, tc.expect, filteredRolesMappings, "filteredRolesMappings")
		})
	}
}
