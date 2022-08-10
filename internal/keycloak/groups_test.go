package keycloak_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
)

func TestGroupsUnmarshal(t *testing.T) {
	var testCases = map[string]struct {
		input  string
		expect []keycloak.Group
	}{
		"unmarshal groups": {
			input: "testdata/groups.json",
			expect: []keycloak.Group{
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
				{
					ID: "9772ddcc-01ea-470a-9c6a-9729fb755ea2",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "internaltest",
						Attributes: map[string][]string{
							"group-lagoon-project-ids": {`{"internaltest":[34,33,25,36,11]}`},
							"lagoon-projects":          {`34,33,25,36,11`},
						},
					},
				},
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
				{
					ID: "8fb9508c-a7e6-445b-a8bb-f28bb0b6eb2d",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-drupal-example",
						Attributes: map[string][]string{
							"lagoon-projects": {`11`},
							"type":            {`project-default-group`},
						},
					},
				},
				{
					ID: "7d5f5769-6904-42cd-9418-d01a1daae6b5",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-drupal9-base",
						Attributes: map[string][]string{
							"lagoon-projects": {`31`},
							"type":            {`project-default-group`},
						},
					},
				},
				{
					ID: "372b0aae-40f1-4af2-b9dd-a4af1d21c845",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-drupal9-solr",
						Attributes: map[string][]string{
							"lagoon-projects": {`34`},
							"type":            {`project-default-group`},
						},
					},
				},
				{
					ID: "9e49d864-d78c-4875-ae46-57daa7151ebe",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-example-ruby-on-rails",
						Attributes: map[string][]string{
							"group-lagoon-project-ids": {`{"project-example-ruby-on-rails":[38]}`},
							"lagoon-projects":          {`38`},
							"type":                     {`project-default-group`},
						},
					},
				},
				{
					ID: "0a442bdd-e89d-4871-8552-80fcc386e236",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-lagoon-website",
						Attributes: map[string][]string{
							"lagoon-projects": {`23`},
							"type":            {`project-default-group`},
						},
					},
				},
				{
					ID: "7cd0cca1-ab32-442f-ba85-adc83d6d6d1a",
					GroupUpdateRepresentation: keycloak.GroupUpdateRepresentation{
						Name: "project-react-example",
						Attributes: map[string][]string{
							"lagoon-projects": {`33`},
							"type":            {`project-default-group`},
						},
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
			var groups []keycloak.Group
			if err = json.Unmarshal(jb, &groups); err != nil {
				tt.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expect, groups) {
				tt.Fatalf("expected %v, got %v", tc.expect, groups)
			}
		})
	}
}
