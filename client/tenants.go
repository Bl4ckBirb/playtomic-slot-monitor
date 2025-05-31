package client

import (
	"context"
	"fmt"
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
	tenant, err := get[*models.Tenant](ctx, c, "/v1/tenants/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, fmt.Errorf("no tenant in the response for %q", id)
	}
	return tenant, nil
}
