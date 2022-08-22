package sync_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
	"go.uber.org/zap"
)

type indexPatternInput struct {
	lpa          string
	projectNames map[int]string
}

type indexPatternOutput struct {
	indexPatterns []string
	err           error
}

func TestGenerateIndexPatterns(t *testing.T) {
	var testCases = map[string]struct {
		input  indexPatternInput
		expect indexPatternOutput
	}{
		"project group": {
			input: indexPatternInput{
				lpa: "33",
				projectNames: map[int]string{
					4:  "baz",
					33: "foo",
					34: "bar",
				},
			},
			expect: indexPatternOutput{
				indexPatterns: []string{
					`/^(application|container|lagoon|router)-logs-foo-_-.+/`,
				},
			},
		},
		"regular group": {
			input: indexPatternInput{
				lpa: "33,34,35",
				projectNames: map[int]string{
					4:  "baz",
					33: "foo",
					34: "bar",
				},
			},
			expect: indexPatternOutput{
				indexPatterns: []string{
					`/^(application|container|lagoon|router)-logs-foo-_-.+/`,
					`/^(application|container|lagoon|router)-logs-bar-_-.+/`,
				},
			},
		},
		"bad attribute": {
			input: indexPatternInput{
				lpa: "33,34,,35",
				projectNames: map[int]string{
					4:  "baz",
					33: "foo",
					34: "bar",
				},
			},
			expect: indexPatternOutput{
				err: fmt.Errorf("an error"),
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			indexPatterns, err := sync.GenerateIndexPatterns(log, tc.input.lpa,
				tc.input.projectNames)
			if (err == nil && tc.expect.err != nil) ||
				(err != nil && tc.expect.err == nil) {
				tt.Fatalf("got %v, expected %v", err, tc.expect.err)
			}
			if !reflect.DeepEqual(indexPatterns, tc.expect.indexPatterns) {
				tt.Fatalf("got %v, expected %v", indexPatterns, tc.expect.indexPatterns)
			}
		})
	}
}

type generateRolesInput struct {
	groups       []keycloak.Group
	projectNames map[int]string
}

type generateRolesOutput struct {
	roles map[string]opensearch.Role
}

func TestGenerateRoles(t *testing.T) {
	var testCases = map[string]struct {
		input  generateRolesInput
		expect generateRolesOutput
	}{
		"generate roles for project group": {
			input: generateRolesInput{
				groups: []keycloak.Group{
					{
						ID: "f6697da3-016a-43cd-ba9f-3f5b91b45302",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "drupal-example",
							Attributes: map[string][]string{
								"group-lagoon-project-ids": {`{"drupal-example":[31,36,34,25,35]}`},
								"lagoon-projects":          {`31,36,34,25,35`},
							},
						},
					},
				},
				projectNames: map[int]string{
					31: "drupal9-base",
					34: "somelongerprojectname",
					35: "drupal10-prerelease",
					36: "delta-backend",
				},
			},
			expect: generateRolesOutput{
				roles: map[string]opensearch.Role{
					"drupal-example": {
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
				},
			},
		},
		"generate roles for regular group": {
			input: generateRolesInput{
				groups: []keycloak.Group{
					{
						ID: "3fc60c90-b72d-4704-8a57-80438adac98d",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "project-beta-ui",
							Attributes: map[string][]string{
								"lagoon-projects": {`27`},
								"type":            {`project-default-group`},
							},
						},
					},
				},
				projectNames: map[int]string{
					26: "abc",
					27: "beta-ui",
					48: "somelongprojectname",
				},
			},
			expect: generateRolesOutput{
				roles: map[string]opensearch.Role{
					"p27": {
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
				},
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			roles := sync.GenerateRoles(log, tc.input.groups, tc.input.projectNames)
			if !reflect.DeepEqual(roles, tc.expect.roles) {
				tt.Fatalf("got:\n%v\nexpected:\n%v\n", roles, tc.expect.roles)
			}
		})
	}
}

type calculateRoleDiffInput struct {
	existing map[string]opensearch.Role
	required map[string]opensearch.Role
}

type calculateRoleDiffOutput struct {
	toCreate map[string]opensearch.Role
	toDelete []string
}

func TestCalculateRoleDiff(t *testing.T) {
	var testCases = map[string]struct {
		input  calculateRoleDiffInput
		expect calculateRoleDiffOutput
	}{
		"extra role and missing role": {
			input: calculateRoleDiffInput{
				existing: map[string]opensearch.Role{
					"drupal-example": {
						RolePermissions: opensearch.RolePermissions{
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
							ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
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
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			toCreate, toDelete :=
				sync.CalculateRoleDiff(tc.input.existing, tc.input.required)
			if !((len(toCreate) == 0 && len(tc.expect.toCreate) == 0) ||
				reflect.DeepEqual(toCreate, tc.expect.toCreate)) {
				tt.Fatalf("toCreate got:\n%v\nexpected:\n%v\n",
					spew.Sdump(toCreate), spew.Sdump(tc.expect.toCreate))
			}
			if !((len(toDelete) == 0 && len(tc.expect.toDelete) == 0) ||
				reflect.DeepEqual(toDelete, tc.expect.toDelete)) {
				tt.Fatalf("toDelete got:\n%v\nexpected:\n%v\n",
					toDelete, tc.expect.toDelete)
			}
		})
	}
}
