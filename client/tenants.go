package client

import (
	"context"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchTenants returns clubs matching params. A nil params applies no filters.
func (c *Client) SearchTenants(ctx context.Context, params *models.SearchTenantsParams) ([]models.Tenant, error) {
	return get[[]models.Tenant](ctx, c, "/v1/tenants", params)
}

// GetTenant returns one club by ID.
func (c *Client) GetTenant(ctx context.Context, id string) (*models.Tenant, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	return get[*models.Tenant](ctx, c, "/v1/tenants/"+url.PathEscape(id), nil)
}
