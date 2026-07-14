package presenter

type OAuthCallbackData struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         any    `json:"user"`
	IsNewUser    bool   `json:"is_new_user"`
}
