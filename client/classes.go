package client

import (
	"context"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchClasses returns classes matching params. A nil params applies no filters.
func (c *Client) SearchClasses(ctx context.Context, params *models.SearchClassesParams) ([]models.Class, error) {
	return get[[]models.Class](ctx, c, "/v1/classes", params)
}
