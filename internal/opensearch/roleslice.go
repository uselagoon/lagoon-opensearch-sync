package opensearch

import (
	"encoding/json"
)

// RoleSlice implements json.Unmarshaler and handles the object returned from
// the Opensearch API roles endpoint.
type RoleSlice []Role

// UnmarshalJSON implements json.Unmarshaler and handles the object returned from the Opensearch API roles endpoint.
func (rs *RoleSlice) UnmarshalJSON(data []byte) error {
	var roles map[string]Role
	if err := json.Unmarshal(data, &roles); err != nil {
		return err
	}
	var roleSlice RoleSlice
	for name, role := range roles {
		role.Name = name
		roleSlice = append(roleSlice, role)
	}
	*rs = roleSlice
	return nil
}

// Implement sort.Interface by Name.
func (rs RoleSlice) Len() int           { return len([]Role(rs)) }
func (rs RoleSlice) Less(i, j int) bool { return rs[i].Name < rs[j].Name }
func (rs RoleSlice) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// MarshalJSON implements json.Marshaler and matches the format of the object
// returned from the Opensearch API roles endpoint.
func (rs RoleSlice) MarshalJSON() ([]byte, error) {
	roles := map[string]Role{}
	for _, role := range rs {
		roles[role.Name] = role
	}
	data, err := json.Marshal(roles)
	if err != nil {
		return nil, err
	}
	return data, nil
}
