package api

import (
	"testing"
	"time"
)

func TestJWT_roundtrip(t *testing.T) {
	secret := "testsecret"
	token, err := generateJWT("user-123", secret)
	if err != nil {
		t.Fatal(err)
	}
	claims, err := validateJWT(token, secret)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("userID: got %s", claims.UserID)
	}
	if claims.ExpiresAt.Before(time.Now()) {
		t.Fatal("token should not be expired")
	}
}

func TestJWT_wrongSecret(t *testing.T) {
	token, _ := generateJWT("user-123", "secret-a")
	_, err := validateJWT(token, "secret-b")
	if err == nil {
		t.Fatal("expected error with wrong secret")
	}
}
