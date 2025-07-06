package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	githubapi "github.com/google/go-github/github"
	_ "github.com/joho/godotenv/autoload"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var (
	GithubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH2_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH2_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:7777/signin/oauth?provider=GitHub",
		Scopes: []string{
			"read:user",
			"user:email",
		},
		Endpoint: github.Endpoint,
	}

	GithubOAuth = ImagineOAuth{
		config: GithubOAuthConfig,
	}
)

type GitHubUserData struct {
	Login                   string `json:"login"`
	Id                      int    `json:"id"`
	NodeId                  string `json:"node_id"`
	AvatarUrl               string `json:"avatar_url"`
	GravatarUrl             string `json:"gravatar_id"`
	URL                     string `json:"url"`
	HtmlUrl                 string `json:"html_url"`
	FollowersUrl            string `json:"followers_url"`
	FollowingUrl            string `json:"following_url"`
	GistsUrl                string `json:"gists_url"`
	StarredUrl              string `json:"starred_url"`
	SubscriptionsUrl        string `json:"subscriptions_url"`
	OrganizationsUrl        string `json:"organizations_url"`
	ReposUrl                string `json:"repos_url"`
	EventsUrl               string `json:"events_url"`
	ReceivedEventsUrl       string `json:"received_events_url"`
	Type                    string `json:"type"`
	SiteAdmin               bool   `json:"site_admin"`
	Name                    string `json:"name"`
	Company                 string `json:"company"`
	Blog                    string `json:"blog"`
	Location                string `json:"location"`
	Email                   string `json:"email"`
	Hireable                bool   `json:"hireable"`
	Bio                     string `json:"bio"`
	TwitterUsername         string `json:"twitter_username"`
	PublicRepos             int    `json:"public_repos"`
	PublicGists             int    `json:"public_gists"`
	Followers               int    `json:"followers"`
	Following               int    `json:"following"`
	CreatedAt               string `json:"created_at"`
	UpdatedAt               string `json:"updated_at"`
	PrivateGists            int    `json:"private_gists"`
	TotalPrivateRepos       int    `json:"total_private_repos"`
	OwnedPrivateRepos       int    `json:"owned_private_repos"`
	DiskUsage               int    `json:"disk_usage"`
	Collaborators           int    `json:"collaborators"`
	TwoFactorAuthentication bool   `json:"two_factor_authentication"`
	Plan                    struct {
		Name          string `json:"name"`
		Space         int    `json:"space"`
		PrivateRepos  int    `json:"private_repos"`
		Collaborators int    `json:"collaborators"`
	} `json:"plan"`
}

func GithubClient(token *oauth2.Token) *githubapi.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(token)
	tc := oauth2.NewClient(ctx, ts)
	client := githubapi.NewClient(tc)

	return client
}

func GithubOAuthHandler(res http.ResponseWriter, req *http.Request, logger *slog.Logger) (*githubapi.User, error) {
	token, err := GithubOAuth.OAuthHandler(res, req, logger)
	if err != nil {
		fmt.Println(token, err)
		return nil, err
	}
	user, _, err := GithubClient(token).Users.Get(context.Background(), "")

	if err != nil {
		errMsg := "error getting user data from GitHub"
		logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}

	return user, nil
}
