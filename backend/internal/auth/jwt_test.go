package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret"

func TestJWTRoundtrip(t *testing.T) {
	now := time.Now()

	token, err := Issue(42, testSecret, now)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}

	got, err := Parse(token, testSecret)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got != 42 {
		t.Fatalf("want user id 42, got %d", got)
	}
}

func TestJWTParseWrongSecret(t *testing.T) {
	token, err := Issue(42, testSecret, time.Now())
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := Parse(token, "other-secret"); err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestJWTParseExpired(t *testing.T) {
	token, err := Issue(42, testSecret, time.Now().Add(-8*24*time.Hour))
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if _, err := Parse(token, testSecret); err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestJWTParseGarbage(t *testing.T) {
	if _, err := Parse("not-a-token", testSecret); err == nil {
		t.Fatal("expected error for garbage token, got nil")
	}
}
