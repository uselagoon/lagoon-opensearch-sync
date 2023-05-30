package sync_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/sync"
	"go.uber.org/zap"
)

type generateIndexPatternsForGroupInput struct {
	group        keycloak.Group
	projectNames map[int]string
}

type generateIndexPatternsForGroupOutput struct {
	indexPatterns []string
	err           error
}

func TestGenerateIndexPatternsForGroup(t *testing.T) {
	var testCases = map[string]struct {
		input  generateIndexPatternsForGroupInput
		expect generateIndexPatternsForGroupOutput
	}{
		"valid group": {
			input: generateIndexPatternsForGroupInput{
				group: keycloak.Group{
					ID: "f6697da3-016a-43cd-ba9f-3f5b91b45302",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "drupal-example",
						Attributes: map[string][]string{
							"group-lagoon-project-ids": {`{"drupal-example":[31,34,35]}`},
							"lagoon-projects":          {`31,34,35`},
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
			expect: generateIndexPatternsForGroupOutput{
				indexPatterns: []string{
					"application-logs-drupal9-base-_-*",
					"container-logs-drupal9-base-_-*",
					"lagoon-logs-drupal9-base-_-*",
					"router-logs-drupal9-base-_-*",
					"application-logs-somelongerprojectname-_-*",
					"container-logs-somelongerprojectname-_-*",
					"lagoon-logs-somelongerprojectname-_-*",
					"router-logs-somelongerprojectname-_-*",
					"application-logs-drupal10-prerelease-_-*",
					"container-logs-drupal10-prerelease-_-*",
					"lagoon-logs-drupal10-prerelease-_-*",
					"router-logs-drupal10-prerelease-_-*",
					"application-logs-*",
					"container-logs-*",
					"lagoon-logs-*",
					"router-logs-*",
				},
			},
		},
		"valid group with unknown pid": {
			input: generateIndexPatternsForGroupInput{
				group: keycloak.Group{
					ID: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "drupal-example2",
						Attributes: map[string][]string{
							"group-lagoon-project-ids": {`{"drupal-example":[31,35,44]}`},
							"lagoon-projects":          {`31,35,44`},
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
			expect: generateIndexPatternsForGroupOutput{
				indexPatterns: []string{
					"application-logs-drupal9-base-_-*",
					"container-logs-drupal9-base-_-*",
					"lagoon-logs-drupal9-base-_-*",
					"router-logs-drupal9-base-_-*",
					"application-logs-drupal10-prerelease-_-*",
					"container-logs-drupal10-prerelease-_-*",
					"lagoon-logs-drupal10-prerelease-_-*",
					"router-logs-drupal10-prerelease-_-*",
					"application-logs-*",
					"container-logs-*",
					"lagoon-logs-*",
					"router-logs-*",
				},
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			indexPatterns, err := sync.GenerateIndexPatternsForGroup(log, tc.input.group,
				tc.input.projectNames, false)
			if (err == nil && tc.expect.err != nil) ||
				(err != nil && tc.expect.err == nil) {
				tt.Fatalf("got err:\n%v\nexpected err:\n%v\n", err, tc.expect.err)
			}
			if !reflect.DeepEqual(indexPatterns, tc.expect.indexPatterns) {
				tt.Fatalf("got:\n%v\nexpected:\n%v\n", indexPatterns,
					tc.expect.indexPatterns)
			}
		})
	}
}

func TestCalculateIndexPatternDiff(t *testing.T) {
	type input struct {
		existing map[string]map[string][]string
		required map[string]map[string]bool
	}
	type output struct {
		toCreate map[string][]string
		toDelete map[string]map[string][]string
	}
	var testCases = map[string]struct {
		input  input
		expect output
	}{
		"no diff": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{""},
						"bar-project":     []string{""},
						"drupal-example2": []string{""},
					},
					sync.HashPrefix("yourgroup"): {
						"baz-project":    []string{""},
						"quux-project":   []string{""},
						"drupal-example": []string{""},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"bar-project":     true,
						"drupal-example2": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{},
				toDelete: map[string]map[string][]string{},
			},
		},
		"create group/tenant": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{""},
						"bar-project":     []string{""},
						"drupal-example2": []string{""},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"bar-project":     true,
						"drupal-example2": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{
					"yourgroup": {
						"baz-project",
						"drupal-example",
						"quux-project",
					},
				},
				toDelete: map[string]map[string][]string{},
			},
		},
		"create project pattern": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{""},
						"drupal-example2": []string{""},
					},
					sync.HashPrefix("yourgroup"): {
						"baz-project":    []string{""},
						"quux-project":   []string{""},
						"drupal-example": []string{""},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"bar-project":     true,
						"drupal-example2": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{
					"mygroup": {
						"bar-project",
					},
				},
				toDelete: map[string]map[string][]string{},
			},
		},
		"delete project": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{"fooID-123"},
						"bar-project":     []string{"barID-123"},
						"drupal-example2": []string{"drupalID-123"},
					},
					sync.HashPrefix("yourgroup"): {
						"baz-project":    []string{"bazID-123"},
						"quux-project":   []string{"quuxID-123"},
						"drupal-example": []string{"drupalID-456"},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"drupal-example2": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{},
				toDelete: map[string]map[string][]string{
					"mygroup": {
						"bar-project": []string{"barID-123"},
					},
				},
			},
		},
		"create and delete": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{"fooID-123"},
						"bar-project":     []string{"barID-123"},
						"drupal-example2": []string{"drupalID-123"},
					},
					sync.HashPrefix("yourgroup"): {
						"baz-project":    []string{"bazID-123"},
						"quux-project":   []string{"quuxID-123"},
						"drupal-example": []string{"drupalID-456"},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"drupal-example2": true,
						"drupal-example3": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{
					"mygroup": {
						"drupal-example3",
					},
				},
				toDelete: map[string]map[string][]string{
					"mygroup": {
						"bar-project": []string{"barID-123"},
					},
				},
			},
		},
		"create and delete duplicate": {
			input: input{
				existing: map[string]map[string][]string{
					sync.HashPrefix("mygroup"): {
						"foo-project":     []string{"fooID-123"},
						"bar-project":     []string{"barID-123"},
						"drupal-example2": []string{"drupalID-123"},
					},
					sync.HashPrefix("yourgroup"): {
						"baz-project":    []string{"bazID-123"},
						"quux-project":   []string{"quuxID-123"},
						"drupal-example": []string{"drupalID-456", "drupalID-789"},
					},
				},
				required: map[string]map[string]bool{
					"mygroup": {
						"foo-project":     true,
						"drupal-example2": true,
						"drupal-example3": true,
					},
					"yourgroup": {
						"baz-project":    true,
						"quux-project":   true,
						"drupal-example": true,
					},
				},
			},
			expect: output{
				toCreate: map[string][]string{
					"mygroup": {
						"drupal-example3",
					},
				},
				toDelete: map[string]map[string][]string{
					"mygroup": {
						"bar-project": []string{"barID-123"},
					},
					"yourgroup": {
						"drupal-example": []string{"drupalID-789"},
					},
				},
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			toCreate, toDelete := sync.CalculateIndexPatternDiff(
				log, tc.input.existing, tc.input.required)
			// Sort slices for accurate comparison. In the case of this test slice
			// order is not important - just that they contain the same set of
			// strings.
			for k := range toCreate {
				sort.Strings(toCreate[k])
			}
			if !reflect.DeepEqual(toCreate, tc.expect.toCreate) {
				tt.Fatalf("got:\n%v\nexpected:\n%v\n", toCreate,
					tc.expect.toCreate)
			}
			if !reflect.DeepEqual(toDelete, tc.expect.toDelete) {
				tt.Fatalf("got:\n%v\nexpected:\n%v\n", toDelete,
					tc.expect.toDelete)
			}
		})
	}
}

func TestGenerateIndexPatterns(t *testing.T) {
	type generateIndexPatternsInput struct {
		groups          []keycloak.Group
		projectNames    map[int]string
		legacyDelimiter bool
	}
	var testCases = map[string]struct {
		input  generateIndexPatternsInput
		expect map[string]map[string]bool
	}{
		"high-level test default group index patterns": {
			input: generateIndexPatternsInput{
				groups: []keycloak.Group{
					{
						ID: "08fef83d-cde7-43a5-8bd2-a18cf440214a",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "foocorp",
							Attributes: map[string][]string{
								"group-lagoon-project-ids": {`{"foocorp":[3133,34435]}`},
								"lagoon-projects":          {`3133,34435`},
							},
						},
					},
					{
						ID: "9f92af94-a7ee-4759-83bb-2b983bd30142",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "project-drupal12-base",
							Attributes: map[string][]string{
								"group-lagoon-project-ids": {`{"project-drupal12-base":[34435]}`},
								"lagoon-projects":          {`34435`},
								"type":                     {`project-default-group`},
							},
						},
					},
				},
				projectNames: map[int]string{
					34435: "drupal12-base",
				},
				legacyDelimiter: false,
			},
			expect: map[string]map[string]bool{
				"foocorp": {
					`application-logs-drupal12-base-_-*`: true,
					`container-logs-drupal12-base-_-*`:   true,
					`lagoon-logs-drupal12-base-_-*`:      true,
					`router-logs-drupal12-base-_-*`:      true,
					`application-logs-*`:                 true,
					`container-logs-*`:                   true,
					`lagoon-logs-*`:                      true,
					`router-logs-*`:                      true,
				},
				"global_tenant": {
					`application-logs-*`: true,
					`container-logs-*`:   true,
					`lagoon-logs-*`:      true,
					`router-logs-*`:      true,
				},
			},
		},
		"high-level test legacy group index patterns": {
			input: generateIndexPatternsInput{
				groups: []keycloak.Group{
					{
						ID: "08fef83d-cde7-43a5-8bd2-a18cf440214a",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "foocorp",
							Attributes: map[string][]string{
								"group-lagoon-project-ids": {`{"foocorp":[3133,34435]}`},
								"lagoon-projects":          {`3133,34435`},
							},
						},
					},
					{
						ID: "9f92af94-a7ee-4759-83bb-2b983bd30142",
						GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
							Name: "project-drupal12-base",
							Attributes: map[string][]string{
								"group-lagoon-project-ids": {`{"project-drupal12-base":[34435]}`},
								"lagoon-projects":          {`34435`},
								"type":                     {`project-default-group`},
							},
						},
					},
				},
				projectNames: map[int]string{
					34435: "drupal12-base",
				},
				legacyDelimiter: true,
			},
			expect: map[string]map[string]bool{
				"foocorp": {
					`application-logs-drupal12-base-*`: true,
					`container-logs-drupal12-base-*`:   true,
					`lagoon-logs-drupal12-base-*`:      true,
					`router-logs-drupal12-base-*`:      true,
					`application-logs-*`:               true,
					`container-logs-*`:                 true,
					`lagoon-logs-*`:                    true,
					`router-logs-*`:                    true,
				},
				"global_tenant": {
					`application-logs-*`: true,
					`container-logs-*`:   true,
					`lagoon-logs-*`:      true,
					`router-logs-*`:      true,
				},
			},
		},
	}
	log := zap.Must(zap.NewDevelopment(zap.AddStacktrace(zap.ErrorLevel)))
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			indexPatterns := sync.GenerateIndexPatterns(
				log, tc.input.groups, tc.input.projectNames,
				tc.input.legacyDelimiter)
			if !reflect.DeepEqual(indexPatterns, tc.expect) {
				tt.Fatalf("got:\n%v\nexpected:\n%v\n", indexPatterns, tc.expect)
			}
		})
	}
}
