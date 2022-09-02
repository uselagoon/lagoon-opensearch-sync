package sync

import (
	"context"

	"github.com/uselagoon/lagoon-opensearch-sync/internal/keycloak"
	"github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"
	"go.uber.org/zap"
)

// tenantsEqual checks the fields Lagoon cares about for functional equality.
func tenantsEqual(a, b opensearch.Tenant) bool {
	if a.Description != b.Description {
		return false
	}
	if a.Hidden != b.Hidden {
		return false
	}
	return true
}

// calculateTenantDiff returns a map of opensearch tenants which should be
// created, and a slice of tenant names which should be deleted, in order to
// reconcile existing with required.
func calculateTenantDiff(existing, required map[string]opensearch.Tenant) (
	map[string]opensearch.Tenant, []string) {
	// calculate tenants to create
	toCreate := map[string]opensearch.Tenant{}
	for name, rTenant := range required {
		eTenant, ok := existing[name]
		if !ok || !tenantsEqual(eTenant, rTenant) {
			toCreate[name] = rTenant
		}
	}
	// calculate tenants to delete
	var toDelete []string
	for name, eTenant := range existing {
		rTenant, ok := required[name]
		if !ok || !tenantsEqual(rTenant, eTenant) {
			// don't delete unnecessarily. create action in opensearch is actually
			// create/replace.
			// https://opensearch.org/docs/2.2/security-plugin/access-control
			// 	/api#create-tenant
			if _, ok := toCreate[name]; !ok {
				toDelete = append(toDelete, name)
			}
		}
	}
	return toCreate, toDelete
}

// generateTenants returns a slice of tenants generated from the given slice of
// keycloak Groups.
func generateTenants(log *zap.Logger,
	groups []keycloak.Group) map[string]opensearch.Tenant {
	tenants := map[string]opensearch.Tenant{}
	for _, group := range groups {
		// we only need tenants for regular groups, not project groups
		if isProjectGroup(log, group) {
			continue
		}
		tenants[group.Name] = opensearch.Tenant{
			Hidden:   false,
			Reserved: false,
			Static:   false,
			TenantDescription: opensearch.TenantDescription{
				Description: group.Name,
			},
		}
	}
	return tenants
}

// given a map of opensearch tenants, return those that are not static,
// reserved, or named admin_tenant.
func filterTenants(
	tenants map[string]opensearch.Tenant) map[string]opensearch.Tenant {
	valid := map[string]opensearch.Tenant{}
	for name, tenant := range tenants {
		if tenant.Static || tenant.Reserved || name == "admin_tenant" {
			continue
		}
		valid[name] = tenant
	}
	return valid
}

// syncTenants reconciles Opensearch tenants with Lagoon keycloak groups.
func syncTenants(ctx context.Context, log *zap.Logger, groups []keycloak.Group,
	o OpensearchService, dryRun bool) {
	// get tenants from Opensearch
	existing, err := o.Tenants(ctx)
	if err != nil {
		log.Error("couldn't get tenants from Opensearch", zap.Error(err))
		return
	}
	// ignore non-lagoon tenants
	existing = filterTenants(existing)
	// generate the tenants required by Lagoon
	required := generateTenants(log, groups)
	// calculate tenants to add/remove
	toCreate, toDelete := calculateTenantDiff(existing, required)
	for _, name := range toDelete {
		if dryRun {
			log.Info("dry run mode: not deleting tenant", zap.String("name", name))
			continue
		}
		err = o.DeleteTenant(ctx, name)
		if err != nil {
			log.Warn("couldn't delete tenant", zap.Error(err))
			continue
		}
		log.Info("deleted tenant", zap.String("name", name))
	}
	for name, tenant := range toCreate {
		if dryRun {
			log.Info("dry run mode: not creating tenant", zap.String("name", name))
			continue
		}
		err = o.CreateTenant(ctx, name, &tenant)
		if err != nil {
			log.Warn("couldn't create tenant", zap.Error(err))
			continue
		}
		log.Info("created tenant", zap.String("name", name))
	}
}
