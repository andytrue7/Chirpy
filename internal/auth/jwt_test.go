package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T){
	userUUID := uuid.New()
	tokenSecret := "secret"
	expiresIn := 1 * time.Microsecond

	tokenString, err := MakeJWT(userUUID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("Error making JWT: %v", err)
	}

	gotUUID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Errorf("Error validating JWT: %v", err)
	}

	if gotUUID != userUUID {
		t.Errorf("Expected UUID %v, got %v", userUUID, gotUUID)
	}
}