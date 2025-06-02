package auth

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleapi "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	libhttp "imagine/common/http"
)

type OAuthServerSetup struct {
	*libhttp.ImagineServer
}

var (
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:7777/signin/oauth?provider=google",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	GoogleOAuth = ImagineOAuth{
		config: GoogleOAuthConfig,
	}
)

type GoogleUserData struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	GivenName     string `json:"given_name"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Hd            string `json:"hd"`
}

func GoogleOAuthHandler(res http.ResponseWriter, req *http.Request, logger *slog.Logger) (*googleapi.Userinfo, error) {
	ctx := context.Background()
	token, err := GoogleOAuth.OAuthHandler(res, req, logger)

	if err != nil {
		return nil, err
	}

	oauth2Service, err := googleapi.NewService(ctx, option.WithTokenSource(GoogleOAuthConfig.TokenSource(ctx, token)))

	if err != nil {
		return nil, err
	}

	userInfo, err := oauth2Service.Userinfo.Get().Do()
	if err != nil {
		return nil, err
	}

	return userInfo, nil
}
