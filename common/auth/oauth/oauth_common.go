package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/render"
	"golang.org/x/oauth2"

	"imagine/common/crypto"
	libhttp "imagine/common/http"
)

type ImagineOAuth struct {
	config *oauth2.Config
}

func (oauth ImagineOAuth) OAuthHandler(res http.ResponseWriter, req *http.Request, logger *slog.Logger) (*oauth2.Token, error) {
	redirectState := req.FormValue("state")
	cookieState, err := req.Cookie("imag-state")
	redirectedStateHash := crypto.CreateHash([]byte(redirectState))
	redirectStateHashString := base64.URLEncoding.EncodeToString(redirectedStateHash)

	if err == http.ErrNoCookie {
		jsonResponse := map[string]any{"message": "state not found"}
		render.JSON(res, req, jsonResponse)
		return nil, err
	}

	if cookieState.Value != redirectStateHashString {
		jsonResponse := map[string]any{"message": "invalid oauth state"}
		logger.Info("invalid oauth state", slog.Group(
			"state",
			slog.String("state", redirectStateHashString),
			slog.String("cookieState", cookieState.Value),
		))

		render.JSON(res, req, jsonResponse)
		return nil, err
	}

	// For some odd reason, after a user has already been authenticated, when they
	// reauthenticate Google returns a URL-encoded code so just incase, make sure
	// decoded first
	// https://stackoverflow.com/a/68917936
	code, _ := url.QueryUnescape(req.FormValue("code"))

	if code == "" {
		res.Write([]byte("Code not found to provide access token"))
		reason := req.FormValue("error_reason")
		if reason == "user_denied" {
			res.Write([]byte("User has denied Permission.."))
		}
		return nil, err
	}

	token, err := oauth.config.Exchange(req.Context(), code)
	if err != nil {
		logger.Error("oauth exchange failed" + err.Error() + "\n")
		return nil, err
	}

	return token, nil
}

func FetchUserDataFromProvider(res http.ResponseWriter, req *http.Request, logger *slog.Logger, apiUrl string) any {
	resp, err := http.Get(apiUrl)
	var userResponse []byte
	var userData any

	if err != nil {
		libhttp.ServerError(res, req, err, logger, nil,
			"Failed to get user info from provider",
			"",
		)
	}

	userResponse, err = io.ReadAll(resp.Body)
	if err != nil {
		libhttp.ServerError(res, req, err, logger, nil,
			"Failed to read user info from body",
			"",
		)
	}

	err = json.Unmarshal(userResponse, &userData)
	if err != nil {
		libhttp.ServerError(res, req, err, logger, nil,
			"Failed to decode JSON",
			"",
		)
	}

	resp.Body.Close()
	return userData
}

func SetupOAuthURL(res http.ResponseWriter, req *http.Request, oauthConfig *oauth2.Config, provider string, state string) (string, error) {
	URL, err := url.Parse(oauthConfig.Endpoint.AuthURL)
	if err != nil {
		return "", errors.New("error parsing oauth url for" + provider)
	}

	parameters := url.Values{}
	parameters.Add("client_id", oauthConfig.ClientID)
	parameters.Add("scope", strings.Join(oauthConfig.Scopes, " "))
	parameters.Add("redirect_uri", oauthConfig.RedirectURL)
	parameters.Add("response_type", "code")
	parameters.Add("state", state)

	URL.RawQuery = parameters.Encode()
	url := URL.String()

	return url, nil
}
