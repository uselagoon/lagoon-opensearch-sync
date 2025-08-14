package sync_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
	"go.uber.org/zap"
)

func TestGenerateIndexPermissionPatterns(t *testing.T) {
	type generateIndexPermissionPatternsInput struct {
		pids         []int
		projectNames map[int]string
	}
	var testCases = map[string]struct {
		input  generateIndexPermissionPatternsInput
		expect []string
	}{
		"project group": {
			input: generateIndexPermissionPatternsInput{
				pids: []int{33},
				projectNames: map[int]string{
					4:  "baz",
					33: "foo",
					34: "bar",
				},
			},
			expect: []string{
				`/^(application|container|lagoon|router)-logs-foo-_-.+/`,
			},
		},
		"regular group": {
			input: generateIndexPermissionPatternsInput{
				pids: []int{33, 34, 35},
				projectNames: map[int]string{
					4:  "baz",
					33: "foo",
					34: "bar",
				},
			},
			expect: []string{
				`/^(application|container|lagoon|router)-logs-foo-_-.+/`,
				`/^(application|container|lagoon|router)-logs-bar-_-.+/`,
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			indexPatterns := sync.GenerateIndexPermissionPatterns(log, tc.input.pids,
				tc.input.projectNames)
			assert.Equal(tt, tc.expect, indexPatterns, "indexPatterns")
		})
	}
}

func TestGenerateRoles(t *testing.T) {
	type generateRolesInput struct {
		groups           []keycloak.Group
		projectNames     map[int]string
		groupProjectsMap map[string][]int
	}
	type generateRolesOutput struct {
		roles map[string]opensearch.Role
	}
	var testCases = map[string]struct {
		input  generateRolesInput
		expect generateRolesOutput
	}{
		"generate roles for regular group and projects": {
			input: generateRolesInput{
				groups: []keycloak.Group{
					{
						ID: "f6697da3-016a-43cd-ba9f-3f5b91b45302",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "drupal-example",
						},
					},
				},
				projectNames: map[int]string{
					31: "drupal9-base",
					34: "somelongerprojectname",
					35: "drupal10-prerelease",
					36: "delta-backend",
				},
				groupProjectsMap: map[string][]int{
					"f6697da3-016a-43cd-ba9f-3f5b91b45302": {31, 36, 34, 25, 35},
				},
			},
			expect: generateRolesOutput{
				roles: map[string]opensearch.Role{
					"drupal-example": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
										"/^(application|container|lagoon|router)-logs-delta-backend-_-.+/",
										"/^(application|container|lagoon|router)-logs-somelongerprojectname-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"drupal-example"},
								},
							},
						},
					},
					"p31": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
									},
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
					"p34": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-somelongerprojectname-_-.+/",
									},
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
					"p35": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
									},
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
					"p36": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-delta-backend-_-.+/",
									},
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
		},
		"generate roles for projects ignoring project group": {
			input: generateRolesInput{
				groups: []keycloak.Group{
					{
						ID: "3fc60c90-b72d-4704-8a57-80438adac98d",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "project-beta-ui",
							Attributes: map[string][]string{
								"type": {`project-default-group`},
							},
						},
					},
				},
				projectNames: map[int]string{
					26: "abc",
					27: "beta-ui",
					48: "somelongprojectname",
				},
				groupProjectsMap: map[string][]int{
					"3fc60c90-b72d-4704-8a57-80438adac98d": {27},
				},
			},
			expect: generateRolesOutput{
				roles: map[string]opensearch.Role{
					"p26": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-abc-_-.+/",
									},
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
					"p27": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-beta-ui-_-.+/",
									},
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
					"p48": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-somelongprojectname-_-.+/",
									},
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
		},
		"generate roles for multi-project project group": {
			input: generateRolesInput{
				groups: []keycloak.Group{
					{
						ID: "3fc60c90-b72d-4704-8a57-80438adac98d",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "project-beta-ui",
							Attributes: map[string][]string{
								"type": {`project-default-group`},
							},
						},
					},
				},
				projectNames: map[int]string{
					26: "abc",
					27: "beta-ui",
					48: "somelongprojectname",
				},
				groupProjectsMap: map[string][]int{
					"3fc60c90-b72d-4704-8a57-80438adac98d": {48, 27, 26},
				},
			},
			expect: generateRolesOutput{
				roles: map[string]opensearch.Role{
					"p26": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-abc-_-.+/",
									},
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
					"p27": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-beta-ui-_-.+/",
									},
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
					"p48": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-somelongprojectname-_-.+/",
									},
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
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			roles := sync.GenerateRoles(
				log, tc.input.groups, tc.input.projectNames, tc.input.groupProjectsMap)
			assert.Equal(tt, tc.expect.roles, roles, "roles")
		})
	}
}

func TestCalculateRoleDiff(t *testing.T) {
	type calculateRoleDiffInput struct {
		existing map[string]opensearch.Role
		required map[string]opensearch.Role
	}
	type calculateRoleDiffOutput struct {
		toCreate map[string]opensearch.Role
		toDelete []string
	}
	var testCases = map[string]struct {
		input  calculateRoleDiffInput
		expect calculateRoleDiffOutput
	}{
		"extra role and missing role": {
			input: calculateRoleDiffInput{
				existing: map[string]opensearch.Role{
					"drupal-example": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"drupal-example"},
								},
							},
						},
					},
					"drupal-example2": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal8-base-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal8-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal7-solr-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"drupal-example"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
				required: map[string]opensearch.Role{
					"drupal-example": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"drupal-example"},
								},
							},
						},
					},
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
					"p23": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-lagoon-website-_-.+/",
									},
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
			expect: calculateRoleDiffOutput{
				toCreate: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p23": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-lagoon-website-_-.+/",
									},
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
				toDelete: []string{"drupal-example2"},
			},
		},
		"index pattern mismatch": {
			input: calculateRoleDiffInput{
				existing: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-nolongerexists-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
				required: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
			expect: calculateRoleDiffOutput{
				toCreate: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
				},
			},
		},
		"update report role": {
			input: calculateRoleDiffInput{
				existing: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-nolongerexists-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
				required: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
			expect: calculateRoleDiffOutput{
				toCreate: map[string]opensearch.Role{
					"internaltest": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{
								"cluster:admin/opendistro/reports/instance/list",
								"cluster:admin/opendistro/reports/instance/get",
								"cluster:admin/opendistro/reports/menu/download",
							},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
										"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
								},
							},
							TenantPermissions: []opensearch.TenantPermission{
								{
									AllowedActions: []string{"kibana_all_write"},
									TenantPatterns: []string{"internaltest"},
								},
							},
						},
					},
					"p11": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{},
							IndexPermissions: []opensearch.IndexPermission{
								{
									AllowedActions: []string{
										"read",
										"indices:monitor/settings/get",
									},
									IndexPatterns: []string{
										"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
									},
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
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			toCreate, toDelete :=
				sync.CalculateRoleDiff(tc.input.existing, tc.input.required)
			assert.Equal(tt, tc.expect.toCreate, toCreate, "toCreate")
			assert.Equal(tt, tc.expect.toDelete, toDelete, "toDelete")
		})
	}
}

func TestGenerateRegularGroupRole(t *testing.T) {
	type generateRoleInput struct {
		group            keycloak.Group
		projectNames     map[int]string
		groupProjectsMap map[string][]int
	}
	var testCases = map[string]struct {
		input  generateRoleInput
		expect opensearch.Role
	}{
		"masked_fields is not null": {
			input: generateRoleInput{
				group: keycloak.Group{
					ID: "49f93046-e326-4d99-92e1-41eee24faf84",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "drooplrox",
					},
				},
				projectNames: map[int]string{
					31: "drupal11-base",
					34: "short",
				},
				groupProjectsMap: map[string][]int{
					"49f93046-e326-4d99-92e1-41eee24faf84": {31, 34},
				},
			},
			expect: opensearch.Role{
				RolePermissions: opensearch.RolePermissions{
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/instance/list",
						"cluster:admin/opendistro/reports/instance/get",
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal11-base-_-.+/",
								"/^(application|container|lagoon|router)-logs-short-_-.+/",
							},
							MaskedFields: []string{},
						},
					},
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"drooplrox"},
						},
					},
				},
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			_, role, err := sync.GenerateRegularGroupRole(
				log,
				tc.input.group,
				tc.input.projectNames,
				tc.input.groupProjectsMap,
			)
			assert.NoError(tt, err, name)
			assert.Equal(tt, tc.expect, *role, name)
			assert.True(tt, role.IndexPermissions[0].MaskedFields != nil, name)
		})
	}
}

func TestGenerateProjectRole(t *testing.T) {
	type generateRoleInput struct {
		id   int
		name string
	}
	var testCases = map[string]struct {
		input  generateRoleInput
		expect opensearch.Role
	}{
		"masked_fields is not null": {
			input: generateRoleInput{
				id:   123,
				name: "pets-com",
			},
			expect: opensearch.Role{
				RolePermissions: opensearch.RolePermissions{
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
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
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			_, role := sync.GenerateProjectRole(tc.input.id, tc.input.name)
			assert.Equal(tt, tc.expect, *role, name)
			assert.True(tt, role.IndexPermissions[0].MaskedFields != nil, name)
		})
	}
}
