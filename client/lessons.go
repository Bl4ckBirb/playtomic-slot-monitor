package client

import (
	"context"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchLessons returns lessons matching params. A nil params applies no filters.
func (c *Client) SearchLessons(ctx context.Context, params *models.SearchLessonsParams) ([]models.Lesson, error) {
	return get[[]models.Lesson](ctx, c, "/v1/lessons", params)
}
