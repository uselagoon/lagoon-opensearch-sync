package sync

import "github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"

func mappingEqual(a, b opensearch.Mapping) bool {
	return a.Type == b.Type && a.IgnoreMalformed == b.IgnoreMalformed
}

func dynamicTemplateEqual(a, b opensearch.DynamicTemplate) bool {
	if a.MatchPattern != b.MatchPattern {
		return false
	}
	if a.Match != b.Match {
		return false
	}
	return mappingEqual(a.Mapping, b.Mapping)
}

func dynamicTemplateMapEqual(a, b map[string]opensearch.DynamicTemplate) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for k, aValue := range a {
		bValue, ok := b[k]
		if !ok {
			return false
		}
		if !dynamicTemplateEqual(aValue, bValue) {
			return false
		}
	}
	return true
}

func dynamicTemplatesEqual(a, b []map[string]opensearch.DynamicTemplate) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !dynamicTemplateMapEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func mappingsEqual(a, b *opensearch.Mappings) bool {
	// check the top level Mapping object
	if a == nil && b == nil {
		return true
	}
	if (a == nil && b != nil) || (a != nil && b == nil) {
		return false
	}
	return dynamicTemplatesEqual(a.DynamicTemplates, b.DynamicTemplates)
}

// indexTemplatesEqual checks the fields Lagoon cares about for functional equality:
// * IndexPatterns
// * Template.Aliases
// * Template.Settings
// * Template.Mappings
// * ComposedOf
func indexTemplatesEqual(a, b opensearch.IndexTemplate) bool {
	if a.Name != b.Name {
		return false
	}
	if !stringSliceEqual(a.IndexTemplateDefinition.ComposedOf,
		b.IndexTemplateDefinition.ComposedOf) {
		return false
	}
	if !stringSliceEqual(a.IndexTemplateDefinition.IndexPatterns,
		b.IndexTemplateDefinition.IndexPatterns) {
		return false
	}
	if !mappingsEqual(a.IndexTemplateDefinition.Template.Mappings,
		b.IndexTemplateDefinition.Template.Mappings) {
		return false
	}
	return true
}
