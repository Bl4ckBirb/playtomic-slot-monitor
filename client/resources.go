package client

import (
	"context"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// GetResources returns the courts of a club, including typed properties such as
// indoor/outdoor and single/double. The availability endpoint identifies courts
// only by resource_id, so this is how a caller resolves those ids to names and
// features.
func (c *Client) GetResources(ctx context.Context, tenantID string) ([]models.TenantResource, error) {
	if tenantID == "" {
		return nil, ErrMissingID
	}
	return get[[]models.TenantResource](ctx, c, "/v1/tenants/"+url.PathEscape(tenantID)+"/resources", nil)
}
