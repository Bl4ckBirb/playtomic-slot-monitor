package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rafa-garcia/go-playtomic-api/models"
)

func TestSearchSocialUsers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/social/users" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if got := q.Get("requester_user_id"); got != "me" {
			t.Errorf("requester_user_id = %q", got)
		}
		if got := q.Get("user_ids"); got != "1,2" {
			t.Errorf("user_ids = %q", got)
		}
		if got := q.Get("exclude_followed"); got != "false" {
			t.Errorf("exclude_followed = %q", got)
		}
		w.Write([]byte(`[{"user_id":"1","full_name":"A","picture":null,"gender":null,"is_premium":true,"following":true}]`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	users, err := c.SearchSocialUsers(context.Background(), &models.SearchSocialUsersParams{
		RequesterUserID: "me",
		UserIDs:         []string{"1", "2"},
	})
	if err != nil {
		t.Fatalf("SearchSocialUsers: %v", err)
	}
	if len(users) != 1 || users[0].FullName != "A" || !users[0].IsPremium || !users[0].Following {
		t.Errorf("users = %+v", users)
	}
	if users[0].Picture != nil || users[0].Gender != nil {
		t.Errorf("expected nil picture/gender, got %v/%v", users[0].Picture, users[0].Gender)
	}
}

func TestGetUserStats(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/social/users/me/stats" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(`{"following":12,"followers":34,"is_followed":true,"is_follower":false,"is_pro_player":false}`))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	stats, err := c.GetUserStats(context.Background(), "me")
	if err != nil {
		t.Fatalf("GetUserStats: %v", err)
	}
	if stats.Following != 12 || stats.Followers != 34 || !stats.IsFollowed {
		t.Errorf("stats = %+v", stats)
	}
}
