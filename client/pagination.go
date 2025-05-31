package client

import (
	"context"
	"iter"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// DefaultPageSize is what the All iterators ask for when params leave Size unset.
const DefaultPageSize = 100

// paginate walks until an empty page. A short page is not the end: the server
// may cap the size we asked for. An error yields once, then stops.
func paginate[T any](ctx context.Context, fetch func(context.Context, int) ([]T, error)) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T

		for page := 0; ; page++ {
			items, err := fetch(ctx, page)
			if err != nil {
				yield(zero, err)
				return
			}

			if len(items) == 0 {
				return
			}

			for _, item := range items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

// AllClasses iterates every page of a class search.
func (c *Client) AllClasses(ctx context.Context, params *models.SearchClassesParams) iter.Seq2[models.Class, error] {
	var q models.SearchClassesParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Class, error) {
		q.Page = page
		return c.SearchClasses(ctx, &q)
	})
}

// AllLessons iterates every page of a lesson search.
func (c *Client) AllLessons(ctx context.Context, params *models.SearchLessonsParams) iter.Seq2[models.Lesson, error] {
	var q models.SearchLessonsParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Lesson, error) {
		q.Page = page
		return c.SearchLessons(ctx, &q)
	})
}

// AllMatches iterates every page of a match search.
func (c *Client) AllMatches(ctx context.Context, params *models.SearchMatchesParams) iter.Seq2[models.Match, error] {
	var q models.SearchMatchesParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Match, error) {
		q.Page = page
		return c.SearchMatches(ctx, &q)
	})
}

// AllTenants iterates every page of a tenant search.
func (c *Client) AllTenants(ctx context.Context, params *models.SearchTenantsParams) iter.Seq2[models.Tenant, error] {
	var q models.SearchTenantsParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Tenant, error) {
		q.Page = page
		return c.SearchTenants(ctx, &q)
	})
}
