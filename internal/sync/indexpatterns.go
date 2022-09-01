package sync

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/hashcode"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"go.uber.org/zap"
)

var (
	indexPatternTemplates = []string{
		`application-logs-%s-*`,
		`container-logs-%s-*`,
		`lagoon-logs-%s-*`,
		`router-logs-%s-*`,
	}
	globalIndexPatterns = []string{
		`application-logs-*`,
		`container-logs-*`,
		`lagoon-logs-*`,
		`router-logs-*`,
	}
	// indexNameInvalid matches characters which cannot appear in Opensearch
	// index names
	indexNameInvalid = regexp.MustCompile(`[^a-z0-9]+`)
	// specialTenants are not associated with a Lagoon group and receive just the
	// globalIndexPatterns
	specialTenants = []string{"global_tenant"}
)

// hashPrefix returns an Opensearch-index-name-sanitized copy of given a string
// s, prefixed with a Java String hashcode.
func hashPrefix(s string) string {
	return fmt.Sprintf("%s_%s", hashcode.String(s),
		// Sanitize s for use in index name the way that the Opensearch security
		// plugin does it.
		//
		//	https://github.com/opensearch-project/security/blob/
		//	f431ec2201e1466b7c12528347a1f54cf64387c9/src/main/java/org/
		//	opensearch/security/rest/TenantInfoAction.java#L198
		indexNameInvalid.ReplaceAllLiteralString(strings.ToLower(s), ""))
}

// calculateIndexPatternDiff returns a map of Opensearch Dashboards index
// patterns which should be created, and a map of index patterns which should
// be deleted, for each tenant, in order to reconcile existing with required.
// existing contains keys which correspond to tenants, but are encoded in
// "index name" form, which is <hashcode>_<sanitized tenant name>.
func calculateIndexPatternDiff(log *zap.Logger,
	existing, required map[string]map[string]bool) (
	map[string][]string, map[string][]string) {
	index2tenant := map[string]string{}
	// calculate index patterns to create
	toCreate := map[string][]string{}
	var index string
	for tenant, patterns := range required {
		if tenant == "global_tenant" {
			index = tenant
		} else {
			index = hashPrefix(tenant)
		}
		// store tenant name for later use in the toDelete loop
		index2tenant[index] = tenant
		for pattern := range patterns {
			if !existing[index][pattern] {
				toCreate[tenant] = append(toCreate[tenant], pattern)
			}
		}
	}
	// calculate index patterns to delete
	toDelete := map[string][]string{}
	for index, patterns := range existing {
		// ignore any custom index patterns created in the admin_tenant
		if index == hashPrefix("admin_tenant") {
			continue
		}
		// Check for the tenant. If it doesn't appear in the required map then we
		// may have a logic bug, or maybe the index hasn't been cleaned up by
		// Opensearch Dashboards?
		tenant, ok := index2tenant[index]
		if !ok {
			log.Warn("unknown index", zap.String("index", index))
			continue
		}
		for pattern := range patterns {
			if !required[tenant][pattern] {
				toDelete[tenant] = append(toDelete[tenant], pattern)
			}
		}
	}
	return toCreate, toDelete
}

// generateIndexPatternsForGroup returns a slice of index patterns for all the
// projects associated with the given group.
func generateIndexPatternsForGroup(log *zap.Logger, group keycloak.Group,
	projectNames map[int]string) ([]string, error) {
	pids, err := projectIDsForGroup(group)
	if err != nil {
		return nil, fmt.Errorf("couldn't get project IDs for group: %v", err)
	}
	var indexPatterns []string
	for _, pid := range pids {
		name, ok := projectNames[pid]
		if !ok {
			log.Debug("invalid project ID in lagoon-projects group attribute",
				zap.Int("projectID", pid))
			continue
		}
		for _, tpl := range indexPatternTemplates {
			indexPatterns = append(indexPatterns, fmt.Sprintf(tpl, name))
		}
	}
	indexPatterns = append(indexPatterns, globalIndexPatterns...)
	return indexPatterns, nil
}

// generateIndexPatterns returns a map of index patterns required by Lagoon
// logging.
func generateIndexPatterns(log *zap.Logger, groups []keycloak.Group,
	projectNames map[int]string) map[string]map[string]bool {
	indexPatterns := map[string]map[string]bool{}
	var patterns []string
	var err error
	for _, group := range groups {
		if isProjectGroup(log, group) {
			continue // project groups don't get any index patterns
		}
		patterns, err = generateIndexPatternsForGroup(log, group, projectNames)
		if err != nil {
			log.Warn("couldn't generate index patterns for group",
				zap.String("group", group.Name), zap.Error(err))
		}
		if indexPatterns[group.Name] == nil {
			indexPatterns[group.Name] = map[string]bool{}
		}
		for _, pattern := range patterns {
			indexPatterns[group.Name][pattern] = true
		}
	}
	// add index patterns for "special" tenants, where special means "not
	// associated with a Lagoon group"
	for _, tenant := range specialTenants {
		indexPatterns[tenant] = map[string]bool{}
		for _, pattern := range globalIndexPatterns {
			indexPatterns[tenant][pattern] = true
		}
	}
	return indexPatterns
}

// syncIndexPatterns reconciles Opensearch Dashboards index patterns with
// Lagoon logging requirements.
func syncIndexPatterns(ctx context.Context, log *zap.Logger,
	groups []keycloak.Group, projectNames map[int]string, o OpensearchService,
	d DashboardsService, dryRun bool) {
	// get index patterns from Opensearch
	existing, err := o.IndexPatterns(ctx)
	if err != nil {
		log.Error("couldn't get index patterns from Opensearch", zap.Error(err))
		return
	}
	// generate the index patterns required by Lagoon
	required := generateIndexPatterns(log, groups, projectNames)
	// calculate index templates to add/remove
	toCreate, toDelete := calculateIndexPatternDiff(log, existing, required)
	for tenant, patterns := range toDelete {
		for _, pattern := range patterns {
			if dryRun {
				log.Info("dry run mode: not deleting index pattern",
					zap.String("tenant", tenant), zap.String("pattern", pattern))
				continue
			}
			err = d.DeleteIndexPattern(ctx, tenant, pattern)
			if err != nil {
				log.Warn("couldn't delete index pattern", zap.String("tenant", tenant),
					zap.String("pattern", pattern), zap.Error(err))
				continue
			}
			log.Info("deleted index pattern", zap.String("tenant", tenant),
				zap.String("pattern", pattern))
		}
	}
	for tenant, patterns := range toCreate {
		for _, pattern := range patterns {
			if dryRun {
				log.Info("dry run mode: not creating index pattern",
					zap.String("tenant", tenant), zap.String("pattern", pattern))
				continue
			}
			err = d.CreateIndexPattern(ctx, tenant, pattern)
			if err != nil {
				log.Warn("couldn't create index pattern", zap.String("tenant", tenant),
					zap.String("pattern", pattern), zap.Error(err))
				continue
			}
			log.Info("created index pattern", zap.String("tenant", tenant),
				zap.String("pattern", pattern))
		}
	}
}
