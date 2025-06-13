package models

import (
	"net/url"
	"strconv"
	"strings"
)

// SocialUser is a player as seen through the social graph, from /v1/social/users.
type SocialUser struct {
	UserID    string  `json:"user_id"`
	FullName  string  `json:"full_name"`
	Gender    *string `json:"gender"`
	Picture   *string `json:"picture"`
	IsPremium bool    `json:"is_premium"`
	ProPlayer bool    `json:"pro_player"`
	Following bool    `json:"following"`
	Blocking  bool    `json:"blocking"`
	Blocked   bool    `json:"blocked"`
}

// UserStats is a player's follower counts, from /v1/social/users/{id}/stats.
type UserStats struct {
	Following   int  `json:"following"`
	Followers   int  `json:"followers"`
	IsFollowed  bool `json:"is_followed"`
	IsFollower  bool `json:"is_follower"`
	IsProPlayer bool `json:"is_pro_player"`
}

// SearchSocialUsersParams defines parameters for the social user search.
type SearchSocialUsersParams struct {
	RequesterUserID string
	UserIDs         []string
	ExcludeFollowed bool
}

// ToURLValues converts SearchSocialUsersParams to url.Values.
func (p *SearchSocialUsersParams) ToURLValues() url.Values {
	values := url.Values{}
	if p == nil {
		return values
	}

	if s := strings.TrimSpace(p.RequesterUserID); s != "" {
		values.Set("requester_user_id", s)
	}

	if len(p.UserIDs) > 0 {
		values.Set("user_ids", strings.Join(p.UserIDs, ","))
	}

	values.Set("exclude_followed", strconv.FormatBool(p.ExcludeFollowed))

	return values
}
