package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateAuthToken(uid string, serverKey string, secret []byte) (string, error) {
	userFingerprint := GenerateRandomBytes(50)
	userFingerprintString := hex.EncodeToString(userFingerprint)

	tokenSHA256 := sha256.New()
	tokenSHA256.Write([]byte(userFingerprintString))
	sha256Sum := tokenSHA256.Sum(nil)
	hashString := hex.EncodeToString(sha256Sum)

	expiryTime := carbon.Now().AddYear().StdTime()

	// TODO: Generate a refresh token as well
	// Note: did I just reimplement oauth lmao????
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// This probably bad and maybe we should generate an ID here instead
		"sub": uid,               // Subject (user identifier)
		"aud": "viz",             // TODO: change this to the actual audience/browser/server URL idk
		"iss": serverKey,         // Issuer
		"iat": time.Now().Unix(), // Issued at
		"exp": expiryTime,        // Expiration time
		"fgp": hashString,        // Fingerprint
	})

	tokenString, err := claims.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func CreateAuthTokenCookie(expireTime time.Time, token string) *http.Cookie {
	return &http.Cookie{
		Name:     "token",
		Value:    token,
		Expires:  expireTime,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}
