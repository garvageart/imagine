package main

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/oauth2"

	"imagine/common/auth"
	oauth "imagine/common/auth/oauth"
	"imagine/common/crypto"
	"imagine/common/entities"
	libhttp "imagine/common/http"
	"imagine/common/uid"
)

type ImagineAuthServer struct {
	*libhttp.ImagineServer
	oauth.ImagineOAuth
}

type ImagineAuthCodeFlow struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// Idk what this is or what I meant to put here
type ImagineAuthPasswordFlow struct {
	State string
}

type ImagineUser struct {
	gorm.Model
	UID           string  `json:"uid"`
	FirstName     *string `json:"first_name"`
	LastName      *string `json:"last_name"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	Password      string  `json:"password"`
	UsedOAuth     *bool   `json:"used_oauth"`
	OAuthProvider *string `json:"oauth_provider"`
	OAuthState    *string `json:"oauth_state"`
	UserToken     *string `json:"user_token"`
}

func (user ImagineUser) TableName() string { return "users" }

type ImagineAPIKeyData struct {
	gorm.Model
	ID              string `json:"id" gorm:"primary_key"`
	APIKeyHashed    string `json:"api_key_hashed"`
	ApplicationID   string `json:"application_id"`
	ApplicationName string `json:"application_name"`
}

func (a ImagineAPIKeyData) TableName() string { return "api_keys" }

// TODO: Migrate to central API file/route
func (server ImagineAuthServer) AuthRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	// Set up auth
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))

	// Setup DB (lmao I promise this wasn't AI)
	database := server.Database
	client := database.Client

	defer func() {
		err := database.Disconnect(client)
		if err != nil {
			logger.Error("error disconnecting from postgres", slog.Any("error", err))
		}

		logger.Info("Disconnected from Postgres")
	}()

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "You have been PONGED by the auth server. What do you want here? 🤨"}
		logger.Info("server pinged", slog.String("request_id", libhttp.GetRequestID(req)))
		render.JSON(res, req, jsonResponse)
	})

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
			ID:           apiKeyId,
			APIKeyHashed: keys["hashed_key"],
		}

		tx := client.Create(apiDataDocument)
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

		tx := client.Select("email", "password").Where("email = ?", user.Email).First(&user)
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

		authToken, err := auth.GenerateAuthToken(user.UID, "api", jwtSecret)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating auth token",
				"Something went wrong, please try again later",
			)
			return
		}

		expiryTime := carbon.Now().AddYear().StdTime()
		http.SetCookie(res, auth.CreateAuthTokenCookie(expiryTime, authToken))

		logger.Info("user authenticated", slog.String("request_id", libhttp.GetRequestID(req)))
		render.JSON(res, req, map[string]string{"message": "User authenticated"})
	})

	router.Get("/oauth", func(res http.ResponseWriter, req *http.Request) {
		authTokenCookie := req.CookiesNamed("imag-auth_token")

		if len(authTokenCookie) > 0 {
			authTokenVal := authTokenCookie[0].Value
			authToken, err := jwt.Parse(authTokenVal, func(token *jwt.Token) (interface{}, error) {
				return jwtSecret, nil
			}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))

			switch {
			case authToken.Valid:
				logger.Info("user already authenticated", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"message": "user already authenticated"}

				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenMalformed):
				logger.Info("malformed token", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "malformed token"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				logger.Error("invalid token signature", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "invalid token signature"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenExpired):
				logger.Info("token expired", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "token expired"}

				res.WriteHeader(http.StatusForbidden)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenNotValidYet):
				logger.Info("token not valid yet", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "token not valid yet"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenExpired):
				logger.Info("token expired", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "token expired"}

				res.WriteHeader(http.StatusForbidden)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenInvalidClaims):
				logger.Info("token invalid claims", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "token invalid claims"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			default:
				libhttp.ServerError(res, req, err, logger, nil,
					"",
					"Something went wrong on our side, please try again later",
				)
				return
			}
		}

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
		var actualUserData any // To store the user data struct
		var userEmail string

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

			userEmail = resp.Email
			actualUserData = resp
		case "github":
			resp, err := oauth.GithubOAuthHandler(res, req, logger)
			if err != nil {
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Github",
					"We encountered an issue while trying to fetch your Github profile. Please try again.",
				)
				return
			}

			userEmail = resp.GetEmail()
			actualUserData = resp
		default:
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("OAuth provider unsupported"))
			http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect)
		}

		expiryTime := carbon.Now().AddYear().StdTime()

		tokenString, err := auth.GenerateAuthToken(userEmail, "api", jwtSecret)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to sign JWT token",
				"Could not process your login information. Please try again later",
			)
			return
		}

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

	router.Post("/user", func(res http.ResponseWriter, req *http.Request) {
		var createdUser entities.User

		err := render.DecodeJSON(req.Body, &createdUser)
		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "invalid request body"})
			return
		}

		logger.Debug("user on req", slog.Any("user", createdUser))
		if createdUser.Name == "" || createdUser.Password == "" || createdUser.Email == "" {
			res.WriteHeader(http.StatusBadRequest)
			render.JSON(res, req, map[string]string{"error": "required fields are missing"})
			return
		}

		createdUser.UID, err = uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate user ID",
				"Something went wrong, please try again later",
			)
		}

		//todo: fix this mess. get a string from the salt and hash seperately
		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		salt := argon.GenerateSalt()
		hashedPass, _ := argon.Hash([]byte(createdUser.Password), salt)
		createdUser.Password = hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

		logger.Debug("user before db", slog.Any("user", createdUser))

		err = client.Create(&createdUser).Error
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to create user",
				"Something went wrong, please try again later",
			)

			return
		}

		res.WriteHeader(http.StatusCreated)
		render.JSON(res, req, createdUser)
	})

	return router
}
