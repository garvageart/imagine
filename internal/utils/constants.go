package utils

import (
	"testing"
)

const (
	AppName = "imagine"
)

var (
	IsProduction  = IsEnvironment("production")
	IsTest        = testing.Testing() || IsEnvironment("test")
	IsDevelopment = !IsProduction && !IsTest

	Environment = func() string {
		switch true {
		case IsProduction:
			return "production"
		case IsTest:
			return "test"
		default:
			return "development"
		}
	}()
)

const (
	AuthTokenCookie    = "imag-auth-token"
	StateCookie        = "imag-state"
	RedirectCookie     = "imag-redirect-state"
	RefreshTokenCookie = "imag-refresh-token"
)
