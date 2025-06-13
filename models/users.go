package models

// User is a Playtomic account, as returned by /v2/users and /v2/users/me.
type User struct {
	UserID                 string     `json:"user_id"`
	FullName               string     `json:"full_name"`
	Email                  string     `json:"email"`
	Picture                *string    `json:"picture"`
	Phone                  *string    `json:"phone"`
	Gender                 *string    `json:"gender"`
	Bio                    *string    `json:"bio"`
	BirthDate              *Time      `json:"birth_date"`
	CountryCode            *string    `json:"country_code"`
	CommunicationsLanguage string     `json:"communications_language"`
	IsValidated            bool       `json:"is_validated"`
	IsEmailVerified        bool       `json:"is_email_verified"`
	IsPhoneVerified        bool       `json:"is_phone_verified"`
	UserRoles              []UserRole `json:"user_roles"`
	CreatedAt              Time       `json:"created_at"`
}

// UserRole is a role grant, scoped to a tenant and scope (both "*" for a plain
// customer), from the user_roles array.
type UserRole struct {
	UserRole string `json:"user_role"`
	TenantID string `json:"tenant_id"`
	ScopeID  string `json:"scope_id"`
}
