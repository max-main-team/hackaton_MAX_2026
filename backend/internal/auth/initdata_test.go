package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// known-answer: хеш посчитан независимо (python hmac) для
// botToken="test-bot-token" и launch_params, полученных из initData ниже.
const testBotToken = "test-bot-token"

const validInitData = "auth_date=1771409719&query_id=4c0ab423&user=%7B%22id%22%3A7%7D&hash=ac715c94969e7b3214b610ffe16202714f265c10dbe358cdaa3a42c2166be7b1"

func TestVerifyValidSignature(t *testing.T) {
	assert.NoError(t, Verify(validInitData, testBotToken))
}

func TestVerifyTamperedPayload(t *testing.T) {
	tampered := "auth_date=1771409719&query_id=4c0ab423&user=%7B%22id%22%3A8%7D&hash=ac715c94969e7b3214b610ffe16202714f265c10dbe358cdaa3a42c2166be7b1"
	assert.Error(t, Verify(tampered, testBotToken))
}

func TestVerifyWrongToken(t *testing.T) {
	assert.Error(t, Verify(validInitData, "other-bot-token"))
}

func TestVerifyMissingHash(t *testing.T) {
	assert.Error(t, Verify("auth_date=1771409719&user=%7B%22id%22%3A7%7D", testBotToken))
}

func TestParseInitData(t *testing.T) {
	raw := "auth_date=1771409719&user=%7B%22id%22%3A42%2C%22first_name%22%3A%22Ivan%22%2C%22language_code%22%3A%22ru%22%7D"

	user, authDate, err := ParseInitData(raw)
	require.NoError(t, err)
	assert.Equal(t, int64(42), user.ID)
	assert.Equal(t, "Ivan", user.FirstName)
	assert.Equal(t, "ru", user.LanguageCode)
	assert.Equal(t, int64(1771409719), authDate.Unix())
}

func TestParseInitDataWithoutUser(t *testing.T) {
	_, _, err := ParseInitData("auth_date=1771409719&query_id=abc")
	require.Error(t, err)
	assert.ErrorContains(t, err, "user is missing")
}
