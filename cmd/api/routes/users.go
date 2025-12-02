package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"imagine/internal/auth"
	"imagine/internal/crypto"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func AccountsRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

	router.Post("/", func(res http.ResponseWriter, req *http.Request) {
		var create dto.UserCreate

		err := render.DecodeJSON(req.Body, &create)
		if err != nil {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
			return
		}

		if create.Name == "" || create.Password == "" || string(create.Email) == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "required fields are missing"})
			return
		}

		var existingUser entities.User
		tx := db.Where("email = ?", string(create.Email)).First(&existingUser)
		switch tx.Error {
		case nil:
			render.Status(req, http.StatusConflict)
			render.JSON(res, req, dto.ErrorResponse{Error: "user already exists"})
			return
		case gorm.ErrRecordNotFound:
			logger.Info("accounts.create: no existing user", slog.String("email", string(create.Email)))
		default:
			logger.Info("accounts.create: db error when checking existing user", slog.Any("error", tx.Error))
			libhttp.ServerError(res, req, tx.Error, logger, nil,
				"Failed to check existing user",
				"Something went wrong, please try again later",
			)
			return
		}

		id, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to generate user ID",
				"Something went wrong, please try again later",
			)
			return
		}

		userEnt := entities.User{
			Uid:       id,
			Email:     string(create.Email),
			Username:  create.Name,
			FirstName: "",
			LastName:  "",
		}

		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		salt := argon.GenerateSalt()
		hashedPass, _ := argon.Hash([]byte(create.Password), salt)
		hashed := hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

		// Create the user and password in a single operation using the
		// non-generated wrapper type so the DTOs remain unchanged.
		uwp := entities.FromUser(userEnt, &hashed)
		txErr := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&uwp).Error; err != nil {
				return err
			}

			return nil
		})

		if txErr != nil {
			libhttp.ServerError(res, req, txErr, logger, nil,
				"Failed to create user",
				"Something went wrong, please try again later",
			)
			return
		}

		render.Status(req, http.StatusCreated)
		render.JSON(res, req, userEnt.DTO())
	})

	router.Get("/{uid}", func(res http.ResponseWriter, req *http.Request) {
		userID := chi.URLParam(req, "uid")

		var user entities.User

		err := db.Where("uid = ?", userID).First(&user).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "user not found"})
				return
			}

			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to get user",
				"Something went wrong, please try again later",
			)
			return
		}

		render.JSON(res, req, user.DTO())
	})

	// Authenticated routes
	router.Group(func(r chi.Router) {
		r.Use(libhttp.AuthMiddleware(db, logger))

		r.Route("/me", func(r chi.Router) {
			r.Get("/", func(res http.ResponseWriter, req *http.Request) {
				user, ok := libhttp.UserFromContext(req)
				if !ok || user == nil {
					render.Status(req, http.StatusUnauthorized)
					render.JSON(res, req, dto.ErrorResponse{Error: "not authenticated"})
					return
				}

				// Compute ETag based on user's UpdatedAt and UID and support conditional
				// requests for bandwidth savings.
				etag := fmt.Sprintf("W/\"%d-%s\"", user.UpdatedAt.UnixNano(), user.Uid)
				res.Header().Set("Cache-Control", "private, max-age=60, must-revalidate")
				res.Header().Set("ETag", etag)

				if match := req.Header.Get("If-None-Match"); match == etag {
					res.WriteHeader(http.StatusNotModified)
					return
				}

				render.JSON(res, req, user.DTO())
			})

			r.Route("/settings", func(r chi.Router) {
				r.Group(func(r chi.Router) {
					r.Use(libhttp.ScopeMiddleware([]auth.Scope{auth.UserSettingsReadScope}))
					r.Get("/", func(res http.ResponseWriter, req *http.Request) {
						user, ok := libhttp.UserFromContext(req)
						if !ok || user == nil {
							render.Status(req, http.StatusUnauthorized)
							render.JSON(res, req, dto.ErrorResponse{Error: "unauthenticated"})
							return
						}

						var defaults []entities.SettingDefault
						if err := db.Find(&defaults).Error; err != nil {
							logger.Error("failed to fetch setting defaults", slog.Any("error", err))
							render.Status(req, http.StatusInternalServerError)
							render.JSON(res, req, dto.ErrorResponse{Error: "internal server error"})
							return
						}

						var overrides []entities.SettingOverride
						if err := db.Where("user_id = ?", user.Uid).Find(&overrides).Error; err != nil {
							logger.Error("failed to fetch user setting overrides", slog.Any("error", err))
							render.Status(req, http.StatusInternalServerError)
							render.JSON(res, req, dto.ErrorResponse{Error: "internal server error"})
							return
						}

						overrideMap := make(map[string]entities.SettingOverride)
						for _, o := range overrides {
							overrideMap[o.Name] = o
						}

						userSettings := make([]dto.UserSetting, 0, len(defaults))
						for _, def := range defaults {
							isEditable := def.IsUserEditable
							userSetting := dto.UserSetting{
								Name:           def.Name,
								DefaultValue:   def.Value,
								ValueType:      string(def.ValueType),
								AllowedValues:  def.AllowedValues,
								IsUserEditable: &isEditable,
								Description:    def.Description,
								Group:          def.Group,
							}

							if override, ok := overrideMap[def.Name]; ok {
								userSetting.Value = override.Value
							} else {
								userSetting.Value = def.Value
							}
							userSettings = append(userSettings, userSetting)
						}

						render.JSON(res, req, userSettings)
					})
				})

				r.Group(func(r chi.Router) {
					r.Use(libhttp.ScopeMiddleware([]auth.Scope{auth.UserSettingsUpdateScope}))
					r.Patch("/", func(res http.ResponseWriter, req *http.Request) {
						user, ok := libhttp.UserFromContext(req)
						if !ok || user == nil {
							render.Status(req, http.StatusUnauthorized)
							render.JSON(res, req, dto.ErrorResponse{Error: "unauthenticated"})
							return
						}

						settingName := req.URL.Query().Get("name")
						if settingName == "" {
							render.Status(req, http.StatusBadRequest)
							render.JSON(res, req, dto.ErrorResponse{Error: "setting name is required"})
							return
						}

						var reqBody struct {
							Value string `json:"value"`
						}
						
						if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
							render.Status(req, http.StatusBadRequest)
							render.JSON(res, req, dto.ErrorResponse{Error: "invalid request body"})
							return
						}

						var userSettingDefaults entities.SettingDefault
						if err := db.Where("name = ?", settingName).First(&userSettingDefaults).Error; err != nil {
							if err == gorm.ErrRecordNotFound {
								render.Status(req, http.StatusNotFound)
								render.JSON(res, req, dto.ErrorResponse{Error: "setting not found"})
								return
							}
							logger.Error("failed to fetch setting definition", slog.Any("error", err))
							render.Status(req, http.StatusInternalServerError)
							render.JSON(res, req, dto.ErrorResponse{Error: "internal server error"})
							return
						}

						if !userSettingDefaults.IsUserEditable {
							render.Status(req, http.StatusForbidden)
							render.JSON(res, req, dto.ErrorResponse{Error: "this setting is not user editable"})
							return
						}

						// Validate the new value based on definition
						if err := validateSettingValue(reqBody.Value, userSettingDefaults); err != nil {
							render.Status(req, http.StatusBadRequest)
							render.JSON(res, req, dto.ErrorResponse{Error: err.Error()})
							return
						}

						override := entities.SettingOverride{
							UserId: user.Uid,
							Name:   settingName,
							Value:  reqBody.Value,
						}

						// Use Upsert to create or update the override
						if err := db.Clauses(clause.OnConflict{
							Columns:   []clause.Column{{Name: "user_id"}, {Name: "name"}},
							DoUpdates: clause.AssignmentColumns([]string{"value"}),
						}).Create(&override).Error; err != nil {
							logger.Error("failed to save setting override", slog.Any("error", err))
							render.Status(req, http.StatusInternalServerError)
							render.JSON(res, req, dto.ErrorResponse{Error: "internal server error"})
							return
						}

						// Return the merged setting for the updated one
						isEditable := userSettingDefaults.IsUserEditable
						userSetting := dto.UserSetting{
							Name:           userSettingDefaults.Name,
							DefaultValue:   userSettingDefaults.Value,
							ValueType:      string(userSettingDefaults.ValueType),
							AllowedValues:  userSettingDefaults.AllowedValues,
							IsUserEditable: &isEditable,
							Description:    userSettingDefaults.Description,
							Group:          userSettingDefaults.Group,
							Value:          override.Value, // The newly set override value
						}

						render.JSON(res, req, userSetting)
					})
				})
			})
		})
	})

	return router
}

// validateSettingValue checks if the provided value conforms to the setting definition.
func validateSettingValue(value string, def entities.SettingDefault) error {
	switch dto.SettingDefaultValueType(def.ValueType) {
	case dto.Boolean:
		if !(strings.EqualFold(value, "true") || strings.EqualFold(value, "false")) {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
	case dto.Integer:
		_, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
	case dto.Enum:
		if def.AllowedValues != nil && len(*def.AllowedValues) > 0 {
			found := false
			for _, v := range *def.AllowedValues {
				if v == value {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value '%s' is not in allowed values: %v", value, *def.AllowedValues)
			}
		}
	case dto.Json:
		var js json.RawMessage
		if err := json.Unmarshal([]byte(value), &js); err != nil {
			return fmt.Errorf("invalid JSON value: %s", value)
		}
	case dto.String:
		// Any string is valid
	default:
		return fmt.Errorf("unknown setting value type: %s", def.ValueType)
	}
	return nil
}
