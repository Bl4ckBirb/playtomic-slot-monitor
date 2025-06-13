package models

import "testing"

func TestSearchSocialUsersParams(t *testing.T) {
	v := (&SearchSocialUsersParams{
		RequesterUserID: "me",
		UserIDs:         []string{"1", "2"},
	}).ToURLValues()

	if v.Get("requester_user_id") != "me" || v.Get("user_ids") != "1,2" {
		t.Errorf("social params = %v", v)
	}
	// The app sends this explicitly even when false.
	if v.Get("exclude_followed") != "false" {
		t.Errorf("exclude_followed = %q", v.Get("exclude_followed"))
	}
}

func TestSearchSocialUsersParamsNilReceiver(t *testing.T) {
	if (*SearchSocialUsersParams)(nil).ToURLValues() == nil {
		t.Error("nil receiver returned nil")
	}
}
