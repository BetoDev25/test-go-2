package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	secret := "secret-key"
	userID := uuid.New()

	t.Run("Valid JWT", func(t *testing.T) {
		expiresIn := time.Hour
		token, err := MakeJWT(userID, secret, expiresIn)
		if err != nil {
			t.Fatalf("failed to make JWT: %v", err)
		}

		id, err := ValidateJWT(token, secret)
		if err != nil {
			t.Fatalf("failed to validate JWT: %v", err)
		}

		if id != userID {
			t.Errorf("expected userID %v, got %v", userID, id)
		}
	})

	t.Run("Expired JWT", func(t *testing.T) {
		expiresIn := -time.Hour
		token, err := MakeJWT(userID, secret, expiresIn)
		if err != nil {
			t.Fatalf("failed to make JWT: %v", err)
		}

		_, err = ValidateJWT(token, secret)
		if err == nil {
			t.Error("expected error for expired token, got nil")
		}
	})

	t.Run("Wrong Secret", func(t *testing.T) {
		expiresIn := time.Hour
		token, err := MakeJWT(userID, secret, expiresIn)
		if err != nil {
			t.Fatalf("failed to make JWT: %v", err)
		}

		_, err = ValidateJWT(token, "wrong-secret")
		if err == nil {
			t.Error("expected error for wrong secret, got nil")
		}
	})
}
