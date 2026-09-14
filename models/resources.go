package models

// Resource represents a court or other resource
type Resource struct {
	ID         string         `json:"id"`
	LockID     string         `json:"lock_id"`
	Name       string         `json:"name"`
	Properties map[string]any `json:"properties"`
}

// TenantResource is a court as returned by GET /v1/tenants/{id}/resources.
// It differs from Resource: the id arrives as "resource_id" and the properties
// are typed (indoor/outdoor, single/double, ...) rather than a free-form map.
type TenantResource struct {
	ResourceID string             `json:"resource_id"`
	Name       string             `json:"name"`
	Properties ResourceProperties `json:"properties"`
}

// IsIndoor reports whether the court's resource_type is "indoor".
func (r TenantResource) IsIndoor() bool {
	return r.Properties.ResourceType == "indoor"
}
