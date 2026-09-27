package auth

import "testing"

// known-answer: хеш посчитан независимо (python hmac) для
// botToken="test-bot-token" и launch_params, полученных из initData ниже.
const testBotToken = "test-bot-token"

const validInitData = "auth_date=1771409719&query_id=4c0ab423&user=%7B%22id%22%3A7%7D&hash=ac715c94969e7b3214b610ffe16202714f265c10dbe358cdaa3a42c2166be7b1"

func TestVerifyValidSignature(t *testing.T) {
	if err := Verify(validInitData, testBotToken); err != nil {
		t.Fatalf("expected valid signature, got error: %v", err)
	}
}

func TestVerifyTamperedPayload(t *testing.T) {
	tampered := "auth_date=1771409719&query_id=4c0ab423&user=%7B%22id%22%3A8%7D&hash=ac715c94969e7b3214b610ffe16202714f265c10dbe358cdaa3a42c2166be7b1"
	if err := Verify(tampered, testBotToken); err == nil {
		t.Fatal("expected error for tampered payload, got nil")
	}
}

func TestVerifyWrongToken(t *testing.T) {
	if err := Verify(validInitData, "other-bot-token"); err == nil {
		t.Fatal("expected error for wrong bot token, got nil")
	}
}

func TestVerifyMissingHash(t *testing.T) {
	if err := Verify("auth_date=1771409719&user=%7B%22id%22%3A7%7D", testBotToken); err == nil {
		t.Fatal("expected error for missing hash, got nil")
	}
}

func TestParseInitData(t *testing.T) {
	raw := "auth_date=1771409719&user=%7B%22id%22%3A42%2C%22first_name%22%3A%22Ivan%22%2C%22language_code%22%3A%22ru%22%7D"

	user, authDate, err := ParseInitData(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if user.ID != 42 || user.FirstName != "Ivan" || user.LanguageCode != "ru" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if authDate.Unix() != 1771409719 {
		t.Fatalf("unexpected auth_date: %d", authDate.Unix())
	}
}
