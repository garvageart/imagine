package routes

import (
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/oauth2"

	"imagine/internal/auth"
	oauth "imagine/internal/auth/oauth"
	"imagine/internal/crypto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
)

type ImagineAuthCodeFlow struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// Idk what this is or what I meant to put here
type ImagineAuthPasswordFlow struct {
	State string
}

type ImagineAPIKeyData struct {
	gorm.Model
	UID       string    `json:"uid" gorm:"primary_key"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	UserID    string    `json:"user_id"`
	Scopes    []string  `json:"scopes"`
	RevokedAt time.Time `json:"revoked_at"`
	Revoked   bool      `json:"revoked"`
}

func (a ImagineAPIKeyData) TableName() string { return "api_keys" }

func AuthRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Get("/apikey", func(res http.ResponseWriter, req *http.Request) {
		keys, err := auth.GenerateAPIKey()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating api key",
				"There was an error generating your API key",
			)
			return
		}

		consumerKey := keys["consumer_key"]
		apiKeyId, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating id for api key",
				"There was an error generating your API key",
			)
			return
		}

		apiDataDocument := &ImagineAPIKeyData{
			UID: apiKeyId,
			Key: keys["hashed_key"],
		}

		tx := db.Create(apiDataDocument)
		if tx.Error != nil {
			if (tx.Error == gorm.ErrDuplicatedKey) || (tx.Error == gorm.ErrInvalidData) {
				// Return the key
				libhttp.ServerError(res, req, err, logger, nil, "error inserting api key into database", "Something went wrong on our side, please try again later")
			}
			return
		}

		logger.Info("Generated an API key", slog.String("request_id", libhttp.GetRequestID(req)))
		jsonResponse := map[string]any{"consumer_key": consumerKey}

		res.WriteHeader(http.StatusOK)
		render.JSON(res, req, jsonResponse)
	})

	router.Post("/login", func(res http.ResponseWriter, req *http.Request) {
		var user entities.User
		err := render.DecodeJSON(req.Body, &user)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"invalid request body",
				"Something went wrong, please try again later",
			)
			return
		}

		// user.Password gets binded in the db request later
		// so store it here
		inputPass := user.Password

		if user.Email == "" || user.Password == "" {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "Required fields are missing"})
			return
		}

		tx := db.Select("email", "password").Where("email = ?", user.Email).First(&user)
		if tx.Error != nil {
			if tx.Error == gorm.ErrRecordNotFound {
				res.WriteHeader(http.StatusNotFound)
				render.JSON(res, req, map[string]string{"error": "Cannot find user with provided email"})
				return
			}
			return
		}

		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		dbPass := strings.Split(user.Password, ":")

		hashedInputPassword, _ := argon.Hash([]byte(inputPass), []byte(dbPass[0]))
		isValidPass := argon.Verify(hashedInputPassword, dbPass[1])

		if !isValidPass {
			res.WriteHeader(http.StatusUnauthorized)
			render.JSON(res, req, map[string]string{"error": "Invalid password"})
			return
		}

		authToken := auth.GenerateAuthToken()
		expiryTime := carbon.Now().AddYear().StdTime()
		http.SetCookie(res, auth.CreateAuthTokenCookie(expiryTime, authToken))

		logger.Info("user authenticated", slog.String("request_id", libhttp.GetRequestID(req)))
		render.JSON(res, req, map[string]string{"message": "User authenticated"})
	})

	router.Get("/oauth", func(res http.ResponseWriter, req *http.Request) {
		var oauthConfig *oauth2.Config
		provider := req.FormValue("provider")

		switch provider {
		case "google":
			oauthConfig = oauth.GoogleOAuthConfig
		case "github":
			oauthConfig = oauth.GithubOAuthConfig
		default:
			providerErr := errors.New("unsupported provider")
			if provider != "" {
				providerErr = errors.New("no provider... provided")
				libhttp.ServerError(res, req, providerErr, logger, nil,
					"",
					"Error siging you in. Please try again later.",
				)
			} else {
				libhttp.ServerError(res, req, providerErr, logger, nil,
					"",
					"Error siging you in. Please try again later.",
				)
			}
		}

		state, err := gonanoid.New(32)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating oauth state",
				"",
			)
			return
		}

		stateHash := crypto.CreateHash([]byte(state))
		encryptedStateB64 := base64.URLEncoding.EncodeToString(stateHash)

		// 5 minute max window to login using the generated state
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-redirect-state",
			Value:    encryptedStateB64,
			Expires:  carbon.Now().AddMinutes(5).StdTime(),
			Path:     "/",
			Secure:   true,
			HttpOnly: true, // client doesn't use this value, make HttpOnly
			SameSite: http.SameSiteNoneMode,
		})

		oauthUrl, err := oauth.SetupOAuthURL(res, req, oauthConfig, provider, state)
		if err != nil {

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"",
			)
			return
		}

		http.Redirect(res, req, oauthUrl, http.StatusTemporaryRedirect)
	})

	router.Post("/oauth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		provider := strings.ToLower(chi.URLParam(req, "provider"))
		state := req.FormValue("state")
		var actualUserData struct {
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}

		switch provider {
		case "google":
			resp, err := oauth.GoogleOAuthHandler(res, req, logger)
			if err != nil {

				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Google",
					"We encountered an issue while trying to fetch your Google profile. Please try again.",
				)
				return
			}

			actualUserData.Email = resp.Email
			actualUserData.Name = resp.Name
			actualUserData.Picture = resp.Picture
		case "github":
			resp, err := oauth.GithubOAuthHandler(res, req, logger)
			if err != nil {
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Github",
					"We encountered an issue while trying to fetch your Github profile. Please try again.",
				)
				return
			}

			actualUserData.Email = resp.GetEmail()
			actualUserData.Name = resp.GetName()
			actualUserData.Picture = resp.GetAvatarURL()
		default:
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("OAuth provider unsupported"))
			http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect)
		}

		expiryTime := carbon.Now().AddYear().StdTime()

		tokenString := auth.GenerateAuthToken()

		// at this point, the state has been validated to be correct
		// and unmodified to use that
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-state",
			Value:    state,
			Expires:  expiryTime,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode, //FIXME: this needs to change to same site
		})

		// delete the temporary redirect state from the browser
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-redirect-state",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		})

		http.SetCookie(res, auth.CreateAuthTokenCookie(expiryTime, tokenString))

		logger.Info("User logged in with OAuth", slog.String("provider", provider))
		render.JSON(res, req, actualUserData)
	})

	return router
}
