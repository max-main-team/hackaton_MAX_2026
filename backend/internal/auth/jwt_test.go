package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret"

func TestJWTRoundtrip(t *testing.T) {
	now := time.Now()

	token, err := Issue(42, testSecret, now)
	require.NoError(t, err)

	got, err := Parse(token, testSecret)
	require.NoError(t, err)
	assert.Equal(t, int64(42), got)
}

func TestJWTParseWrongSecret(t *testing.T) {
	token, err := Issue(42, testSecret, time.Now())
	require.NoError(t, err)

	_, err = Parse(token, "other-secret")
	assert.Error(t, err)
}

func TestJWTParseExpired(t *testing.T) {
	token, err := Issue(42, testSecret, time.Now().Add(-8*24*time.Hour))
	require.NoError(t, err)

	_, err = Parse(token, testSecret)
	assert.Error(t, err)
}

func TestJWTParseGarbage(t *testing.T) {
	_, err := Parse("not-a-token", testSecret)
	assert.Error(t, err)
}
