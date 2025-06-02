package main

type ImagineUser struct {
	UUID        string `json:"uuid"`
	ID          string `json:"id"`
	Username    string `json:"username"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	Hash        string `json:"hash"`
	Salt        string `json:"salt"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	DeletedAt   string `json:"deleted_at"`
	UsedOAuth   bool   `json:"used_oauth"`
	OAuthProvider string `json:"oauth_provider"`
	OAuthState string `json:"oauth_state"`
	UserToken string `json:"user_token"`
}