package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Shaped like the capture: picture, phone, gender and birth_date come back null
// for an account that has not set them.
const userJSON = `{
  "user_id": "user-1",
  "full_name": "Rafa G",
  "email": "player@example.com",
  "picture": null,
  "phone": null,
  "gender": null,
  "birth_date": null,
  "communications_language": "en",
  "is_validated": true,
  "is_email_verified": true,
  "is_phone_verified": false,
  "user_roles": [{"user_role": "ROLE_CUSTOMER", "tenant_id": "*", "scope_id": "*"}],
  "created_at": "2020-01-02T10:00:00"
}`

func TestGetMe(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/users/me" {
			t.Errorf("path = %s, want /v2/users/me", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer t" {
			t.Errorf("Authorization = %q, want the token", got)
		}
		w.Write([]byte(userJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL), WithToken("t"))
	me, err := c.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if me.UserID != "user-1" || me.Email != "player@example.com" {
		t.Errorf("user = %+v", me)
	}
	if me.CreatedAt.String() != "2020-01-02T10:00:00" {
		t.Errorf("CreatedAt = %q", me.CreatedAt)
	}
	if len(me.UserRoles) != 1 || me.UserRoles[0].UserRole != "ROLE_CUSTOMER" {
		t.Errorf("roles = %+v", me.UserRoles)
	}
	// Null optionals stay nil rather than becoming empty strings.
	if me.Picture != nil || me.Phone != nil || me.Gender != nil || me.BirthDate != nil {
		t.Errorf("expected nil optionals, got picture=%v phone=%v gender=%v birth=%v",
			me.Picture, me.Phone, me.Gender, me.BirthDate)
	}
}

func TestGetMeParsesBirthDate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// A date-only birth_date, which models.Time accepts.
		w.Write([]byte(`{"user_id":"1","birth_date":"1990-05-15"}`))
	}))
	defer srv.Close()

	me, err := NewClient(WithBaseURL(srv.URL)).GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe: %v", err)
	}
	if me.BirthDate == nil || me.BirthDate.String() != "1990-05-15" {
		t.Errorf("BirthDate = %v", me.BirthDate)
	}
}

func TestGetUser(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/users/user-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.Write([]byte(userJSON))
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL))
	u, err := c.GetUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("GetUser: %v", err)
	}
	if u.FullName != "Rafa G" {
		t.Errorf("user = %+v", u)
	}

	if _, err := c.GetUser(context.Background(), ""); !errors.Is(err, ErrMissingID) {
		t.Errorf("empty id gave %v, want ErrMissingID", err)
	}
}
