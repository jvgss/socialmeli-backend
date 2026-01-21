package http

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToken_MakeAndParse_OK(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	tok, err := MakeToken(123, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	c, err := ParseToken(tok)
	require.NoError(t, err)
	require.Equal(t, 123, c.Sub)
	require.Greater(t, c.Exp, c.Iat)
}

func TestToken_Parse_InvalidAndExpired(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")

	// formato inválido
	_, err := ParseToken("abc")
	require.Error(t, err)

	// assinatura inválida
	tok, err := MakeToken(1, time.Hour)
	require.NoError(t, err)
	require.NotEmpty(t, tok)
	// mexe no token para quebrar assinatura
	bad := tok + "x"
	_, err = ParseToken(bad)
	require.Error(t, err)

	// expirado
	tok, err = MakeToken(1, -1*time.Second)
	require.NoError(t, err)
	_, err = ParseToken(tok)
	require.Error(t, err)
}

func TestTokenSecret_DefaultDevSecret(t *testing.T) {
	// garante que sem JWT_SECRET o secret cai no default
	old := os.Getenv("JWT_SECRET")
	t.Cleanup(func() { _ = os.Setenv("JWT_SECRET", old) })
	_ = os.Unsetenv("JWT_SECRET")
	require.Equal(t, "dev-secret", string(tokenSecret()))
}
