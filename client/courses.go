package client

import (
	"context"
	"fmt"
	"iter"
	"net/url"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

// SearchCourses returns courses matching params. A nil params applies no filters.
func (c *Client) SearchCourses(ctx context.Context, params *models.SearchCoursesParams) ([]models.Course, error) {
	return get[[]models.Course](ctx, c, "/v1/courses", params)
}

// GetCourse returns one course by ID.
func (c *Client) GetCourse(ctx context.Context, id string) (*models.Course, error) {
	if id == "" {
		return nil, ErrMissingID
	}
	course, err := get[*models.Course](ctx, c, "/v1/courses/"+url.PathEscape(id), nil)
	if err != nil {
		return nil, err
	}
	if course == nil {
		return nil, fmt.Errorf("no course in the response for %q", id)
	}
	return course, nil
}

// AllCourses iterates every page of a course search.
func (c *Client) AllCourses(ctx context.Context, params *models.SearchCoursesParams) iter.Seq2[models.Course, error] {
	var q models.SearchCoursesParams
	if params != nil {
		q = *params
	}
	if q.Size <= 0 {
		q.Size = DefaultPageSize
	}

	return paginate(ctx, func(ctx context.Context, page int) ([]models.Course, error) {
		q.Page = page
		return c.SearchCourses(ctx, &q)
	})
}
