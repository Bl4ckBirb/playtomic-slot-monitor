package models

// AuthMethods lists how a given email can authenticate, from /v3/auth/methods.
type AuthMethods struct {
	AvailableMethods []AuthMethod `json:"available_methods"`
	ActionRequired   bool         `json:"action_required"`
}

// AuthMethod is one way in, such as a password or a Google account.
type AuthMethod struct {
	Type string `json:"type"`
}
