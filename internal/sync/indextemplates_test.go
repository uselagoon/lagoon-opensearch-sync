package sync

import (
	"reflect"
	"testing"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
)

func TestCalculateIndexTemplateDiff(t *testing.T) {
	type input struct {
		existing map[string]opensearch.IndexTemplate
		required map[string]opensearch.IndexTemplate
	}
	type output struct {
		toCreate map[string]opensearch.IndexTemplate
		toDelete []string
	}
	var testCases = map[string]struct {
		input  input
		expect output
	}{
		"no diff": {
			input: input{
				existing: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
				required: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
			},
			expect: output{
				toCreate: map[string]opensearch.IndexTemplate{},
				toDelete: nil,
			},
		},
		"create index template": {
			input: input{
				existing: map[string]opensearch.IndexTemplate{},
				required: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
			},
			expect: output{
				toCreate: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
				toDelete: nil,
			},
		},
		"keep unknown index template": {
			input: input{
				existing: map[string]opensearch.IndexTemplate{
					"routerlogs":       {Name: "routerlogs"},
					"manually-created": {Name: "manually-created"},
				},
				required: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
			},
			expect: output{
				toCreate: map[string]opensearch.IndexTemplate{},
				toDelete: nil,
			},
		},
		"keep custom index template": {
			input: input{
				existing: map[string]opensearch.IndexTemplate{
					"routerlogs":  {Name: "routerlogs"},
					"custom-logs": {Name: "custom-logs"},
				}, required: map[string]opensearch.IndexTemplate{
					"routerlogs": {Name: "routerlogs"},
				},
			},
			expect: output{
				toCreate: map[string]opensearch.IndexTemplate{},
				toDelete: nil,
			},
		},
		"replace unequal index temlate": {
			input: input{
				existing: map[string]opensearch.IndexTemplate{
					"routerlogs": {
						Name: "routerlogs",
						IndexTemplateDefinition: opensearch.IndexTemplateDefinition{
							IndexPatterns: []string{"foo"},
						},
					},
					"manually-created": {Name: "manually-created"},
				},
				required: map[string]opensearch.IndexTemplate{
					"routerlogs": {
						Name: "routerlogs",
						IndexTemplateDefinition: opensearch.IndexTemplateDefinition{
							IndexPatterns: []string{"bar"},
						},
					},
					"manually-created": {Name: "manually-created"},
				},
			},
			expect: output{
				toCreate: map[string]opensearch.IndexTemplate{
					"routerlogs": {
						Name: "routerlogs",
						IndexTemplateDefinition: opensearch.IndexTemplateDefinition{
							IndexPatterns: []string{"bar"},
						},
					},
				},
				toDelete: []string{
					"routerlogs",
				},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(tt *testing.T) {
			toCreate, toDelete :=
				calculateIndexTemplateDiff(tc.input.existing, tc.input.required)
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
