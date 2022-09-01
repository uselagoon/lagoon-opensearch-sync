package sync

import "github.com/uselagoon/lagoon-opensearch-sync/internal/opensearch"

func stringSliceEqual(a, b []string) bool {
	// short circuit: we don't care if one is a nil pointer
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if b[i] != a[i] {
			return false
		}
	}
	return true
}

func indexPermissionsEqual(a, b []opensearch.IndexPermission) bool {
	// short circuit: we don't care if one is a nil pointer
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !stringSliceEqual(a[i].AllowedActions, b[i].AllowedActions) {
			return false
		}
		if !stringSliceEqual(a[i].FLS, b[i].FLS) {
			return false
		}
		if !stringSliceEqual(a[i].IndexPatterns, b[i].IndexPatterns) {
			return false
		}
		if !stringSliceEqual(a[i].MaskedFields, b[i].MaskedFields) {
			return false
		}
	}
	return true
}

func tenantPermissionsEqual(a, b []opensearch.TenantPermission) bool {
	// short circuit: we don't care if one is a nil pointer
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !stringSliceEqual(a[i].AllowedActions, b[i].AllowedActions) {
			return false
		}
		if !stringSliceEqual(a[i].TenantPatterns, b[i].TenantPatterns) {
			return false
		}
	}
	return true
}

// rolesEqual checks the fields Lagoon cares about for functional equality
func rolesEqual(a, b opensearch.Role) bool {
	if !stringSliceEqual(a.ClusterPermissions, b.ClusterPermissions) {
		return false
	}
	if a.Hidden != b.Hidden {
		return false
	}
	if !indexPermissionsEqual(a.IndexPermissions, b.IndexPermissions) {
		return false
	}
	if !tenantPermissionsEqual(a.TenantPermissions, b.TenantPermissions) {
		return false
	}
	return true
}
