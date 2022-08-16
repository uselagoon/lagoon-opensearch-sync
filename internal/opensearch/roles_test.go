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

func TestRolesUnmarshal(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect map[string]opensearch.Role
	}{
		"unmarshal roles": {
			input: "testdata/roles.json",
			expect: map[string]opensearch.Role{
				"alerting_crud_alerts": {
					ClusterPermissions: []string{},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{"crud"},
							FLS:            []string{},
							IndexPatterns:  []string{".opendistro-alerting-alert*"},
							MaskedFields:   []string{},
						},
					},
					Reserved:          true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"alerting_full_access": {
					ClusterPermissions: []string{},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{"crud"},
							FLS:            []string{},
							IndexPatterns: []string{
								".opendistro-alerting-config",
								".opendistro-alerting-alert*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"alerting_view_alerts": {
					ClusterPermissions: []string{},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{"read"},
							FLS:            []string{},
							IndexPatterns:  []string{".opendistro-alerting-alert*"},
							MaskedFields:   []string{},
						},
					},
					Reserved:          true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"all_access": {
					ClusterPermissions: []string{"*"},
					Description:        "Allow full access to all indices and all cluster APIs",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{"*"},
							FLS:            []string{},
							IndexPatterns:  []string{"*"},
							MaskedFields:   []string{},
						},
					},
					Reserved: true,
					Static:   true,
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"*"},
						},
					},
				},
				"amazee.io internal": {
					ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS:           []string{},
							IndexPatterns: []string{},
							MaskedFields:  []string{},
						},
					},
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"amazee.io internal"},
						},
					},
				},
				"drupal-example": {
					ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
								"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
								"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
							},
							MaskedFields: []string{},
						},
					},
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"drupal-example"},
						},
					},
				},
				"internaltest": {
					ClusterPermissions: []string{"cluster:admin/opendistro/reports/menu/download"},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
								"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
								"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
								"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
							},
							MaskedFields: []string{},
						},
					},
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"internaltest"},
						},
					},
				},
				"kibana_read_only": {
					ClusterPermissions: []string{},
					IndexPermissions:   []opensearch.IndexPermission{},
					Reserved:           true,
					TenantPermissions:  []opensearch.TenantPermission{},
				},
				"kibana_server": {
					ClusterPermissions: []string{
						"cluster_monitor",
						"cluster_composite_ops",
						"indices:admin/template*",
						"indices:data/read/scroll*",
					},
					Description: "Provide the minimum permissions for the Kibana server",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".kibana",
								".opensearch_dashboards",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".kibana-6",
								".opensearch_dashboards-6",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".kibana_*",
								".opensearch_dashboards_*",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".tasks",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".management-beats*",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices:admin/aliases*",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"kibana_user": {
					ClusterPermissions: []string{"cluster_composite_ops"},
					Description:        "Provide the minimum permissions for a kibana user",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"delete",
								"manage",
								"index",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".kibana",
								".kibana-6",
								".kibana_*",
								".opensearch_dashboards",
								".opensearch_dashboards-6",
								".opensearch_dashboards_*",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								".tasks",
								".management-beats",
								"*:.tasks",
								"*:.management-beats",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"lagoonadmin": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"lagoonadmin-has-no-project",
							},
							MaskedFields: []string{},
						},
					},
					TenantPermissions: []opensearch.TenantPermission{
						{
							AllowedActions: []string{"kibana_all_write"},
							TenantPatterns: []string{"lagoonadmin"},
						},
					},
				},
				"logstash": {
					ClusterPermissions: []string{
						"cluster_monitor",
						"cluster_composite_ops",
						"indices:admin/template/get",
						"indices:admin/template/put",
						"cluster:admin/ingest/pipeline/put",
						"cluster:admin/ingest/pipeline/get",
					},
					Description: "Provide the minimum permissions for logstash and beats",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"crud",
								"create_index",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"logstash-*",
							},
							MaskedFields: []string{},
						},
						{
							AllowedActions: []string{
								"crud",
								"create_index",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*beat*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"manage_snapshots": {
					ClusterPermissions: []string{
						"manage_snapshots",
					},
					Description: "Provide the minimum permissions for managing snapshots",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"indices:data/write/index",
								"indices:admin/create",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"own_index": {
					ClusterPermissions: []string{
						"cluster_composite_ops",
					},
					Description: "Allow all for indices named like the current user",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"indices_all",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"p11": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal-example-_-.+/",
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
				"p23": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-lagoon-website-_-.+/",
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
				"p24": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-ckan-lagoon-_-.+/",
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
				"p27": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-beta-ui-_-.+/",
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
				"p29": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-fastly-controller-testing-_-.+/",
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
				"p31": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal9-base-_-.+/",
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
				"p33": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-react-example-_-.+/",
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
				"p34": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal9-solr-_-.+/",
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
				"p36": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-drupal10-prerelease-_-.+/",
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
				"p37": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-test6-drupal-example-simple-_-.+/",
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
				"p38": {
					ClusterPermissions: []string{
						"cluster:admin/opendistro/reports/menu/download",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
								"indices:monitor/settings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"/^(application|container|lagoon|router)-logs-example-ruby-on-rails-_-.+/",
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
				"prometheus_exporter": {
					ClusterPermissions: []string{
						"cluster_monitor",
						"cluster:admin/snapshot/status",
						"cluster:admin/repository/get",
					},
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"indices_monitor",
								"indices:admin/mappings/get",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"readall": {
					ClusterPermissions: []string{
						"cluster_composite_ops_ro",
					},
					Description: "Provide the minimum permissions for to readall indices",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"readall_and_monitor": {
					ClusterPermissions: []string{
						"cluster_monitor",
						"cluster_composite_ops_ro",
					},
					Description: "Provide the minimum permissions for to readall indices and monitor the cluster",
					IndexPermissions: []opensearch.IndexPermission{
						{
							AllowedActions: []string{
								"read",
							},
							FLS: []string{},
							IndexPatterns: []string{
								"*",
							},
							MaskedFields: []string{},
						},
					},
					Reserved:          true,
					Static:            true,
					TenantPermissions: []opensearch.TenantPermission{},
				},
				"security_rest_api_access": {
					ClusterPermissions: []string{},
					IndexPermissions:   []opensearch.IndexPermission{},
					Reserved:           true,
					TenantPermissions:  []opensearch.TenantPermission{},
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
			// check for missing fields in Role
			var roles map[string]opensearch.Role
			decoder := json.NewDecoder(bytes.NewReader(jb))
			decoder.DisallowUnknownFields()
			if err = decoder.Decode(&roles); err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expect, roles) {
				tt.Fatalf("expected:\n%s\ngot\n%s\n",
					spew.Sdump(tc.expect), spew.Sdump(roles))
			}
		})
	}
}
