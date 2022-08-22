package sync

import (
	"context"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// calculateIndexTemplateDiff returns a map of opensearch index templates which
// should be created, and a slice of index template names which should be
// deleted, in order to reconcile existing with required.
func calculateIndexTemplateDiff(existing,
	required map[string]opensearch.IndexTemplate) (
	map[string]opensearch.IndexTemplate, []string) {
	// calculate index template to create
	toCreate := map[string]opensearch.IndexTemplate{}
	for name, rIndexTemplate := range required {
		eIndexTemplate, ok := existing[name]
		if !ok || !indexTemplatesEqual(eIndexTemplate, rIndexTemplate) {
			toCreate[name] = rIndexTemplate
		}
	}
	// calculate index templates to delete
	var toDelete []string
	for name, eIndexTemplate := range existing {
		rIndexTemplate, ok := required[name]
		if !ok || !indexTemplatesEqual(rIndexTemplate, eIndexTemplate) {
			toDelete = append(toDelete, name)
		}
	}
	return toCreate, toDelete
}

// generateIndexTemplates returns a map of index templates required by Lagoon
// logging.
func generateIndexTemplates() map[string]opensearch.IndexTemplate {
	return map[string]opensearch.IndexTemplate{
		"routerlogs": {
			Name: "routerlogs",
			IndexTemplateDefinition: opensearch.IndexTemplateDefinition{
				IndexPatterns: []string{"router-logs-*"},
				Template: opensearch.Template{
					Mappings: &opensearch.Mappings{
						DynamicTemplates: []map[string]opensearch.DynamicTemplate{
							{
								"remote_addr": {
									MatchMappingType: "string",
									Match:            "remote_addr",
									Mapping: opensearch.Mapping{
										Type:            "ip",
										IgnoreMalformed: true,
									},
								},
							},
							{
								"true-client-ip": {
									MatchMappingType: "string",
									Match:            "true-client-ip",
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
	}
}

// syncIndexTemplates reconciles Opensearch index templates with Lagoon logging
// requirements.
func syncIndexTemplates(ctx context.Context, log *zap.Logger,
	o OpensearchService, dryRun bool) {
	// get index templates from Opensearch
	existing, err := o.IndexTemplates(ctx)
	if err != nil {
		log.Error("couldn't get index templates from Opensearch", zap.Error(err))
		return
	}
	// generate the index templates required by Lagoon
	required := generateIndexTemplates()
	// calculate index templates to add/remove
	toCreate, toDelete := calculateIndexTemplateDiff(existing, required)
	for _, name := range toDelete {
		if dryRun {
			log.Info("dry run mode: not deleting index template",
				zap.String("name", name))
			continue
		}
		err = o.DeleteIndexTemplate(ctx, name)
		if err != nil {
			log.Warn("couldn't delete index template", zap.Error(err))
		} else {
			log.Info("deleted index template", zap.String("name", name))
		}
	}
	for name, it := range toCreate {
		if dryRun {
			log.Info("dry run mode: not creating index template",
				zap.String("name", name))
			continue
		}
		err = o.CreateIndexTemplate(ctx, name, &it)
		if err != nil {
			log.Warn("couldn't create index template", zap.Error(err))
		} else {
			log.Info("created index template", zap.String("name", name))
		}
	}
}
