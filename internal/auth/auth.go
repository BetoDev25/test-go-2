package auth

import (
	"time"
	"strings"
	"errors"
	"net/http"
	"encoding/hex"
	"crypto/rand"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
        "github.com/alexedwards/argon2id"
)

type MyCustomClaims struct {
        Foo string `json:"foo"`
        jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
        hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
        if err != nil {
                return "", err
        }

        return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
        match, err := argon2id.ComparePasswordAndHash(password, hash)
        if err != nil {
                return false, err
        }

        return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	signingKey := []byte(tokenSecret)

	now := time.Now().UTC()
	expiry := now.Add(expiresIn)

	claims := MyCustomClaims{
		"bar",
		jwt.RegisteredClaims{
			Issuer:   "chirpy-access",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiry),
			Subject:   userID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsStruct := &MyCustomClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claimsStruct,
		func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil

}

var ErrNoAuthHeaderIncluded = errors.New("no auth header included in request")

func GetBearerToken(headers http.Header) (string, error) {
	tokenString := headers.Get("Authorization")
	if tokenString == ""{
		return "", ErrNoAuthHeaderIncluded
	}

	parts := strings.Split(tokenString, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("malformed authorization header")
	}

	return parts[1], nil
}

func MakeRefreshToken() string {
	data := make([]byte, 32)
	_, _ = rand.Read(data)
	return hex.EncodeToString(data)
}
