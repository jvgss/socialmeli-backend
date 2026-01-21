package service

import (
	"testing"

	"socialmeli/internal/store"

	"github.com/stretchr/testify/require"
)

func TestAuthService_Register_ValidatesAndCreatesAccount(t *testing.T) {
	st := store.NewMemoryStore()
	s := NewAuthService(st)

	acc, err := s.Register(RegisterPayload{
		Name:     "  João  ",
		Email:    "  Joao@Example.com ",
		Password: "123456",
		IsSeller: true,
	})
	require.NoError(t, err)
	require.NotZero(t, acc.ID)
	require.Equal(t, "João", acc.Name)
	require.Equal(t, "joao@example.com", acc.Email)
	require.True(t, acc.IsSeller)
	require.NotEmpty(t, acc.PasswordHash)
	require.NotEqual(t, "123456", acc.PasswordHash)
	require.False(t, acc.CreatedAt.IsZero())

	// email duplicado
	_, err = s.Register(RegisterPayload{
		Name:     "Maria",
		Email:    "JOAO@example.com",
		Password: "abcdef",
	})
	require.Error(t, err)
	require.ErrorIs(t, err, store.ErrEmailTaken)
}

func TestAuthService_Register_InvalidInputs(t *testing.T) {
	st := store.NewMemoryStore()
	s := NewAuthService(st)

	_, err := s.Register(RegisterPayload{Name: "", Email: "a@b.com", Password: "123456"})
	require.Error(t, err)

	_, err = s.Register(RegisterPayload{Name: "Ok", Email: "", Password: "123456"})
	require.Error(t, err)

	_, err = s.Register(RegisterPayload{Name: "Ok", Email: "invalid", Password: "123456"})
	require.Error(t, err)

	_, err = s.Register(RegisterPayload{Name: "Ok", Email: "a@b.com", Password: "123"})
	require.Error(t, err)
}

func TestAuthService_Login(t *testing.T) {
	st := store.NewMemoryStore()
	s := NewAuthService(st)

	_, err := s.Register(RegisterPayload{Name: "Ana", Email: "ana@ex.com", Password: "123456"})
	require.NoError(t, err)

	acc, err := s.Login(LoginPayload{Email: "  ANA@ex.com ", Password: "123456"})
	require.NoError(t, err)
	require.Equal(t, "ana@ex.com", acc.Email)

	_, err = s.Login(LoginPayload{Email: "ana@ex.com", Password: "wrong"})
	require.ErrorIs(t, err, ErrInvalidCredentials)

	_, err = s.Login(LoginPayload{Email: "missing@ex.com", Password: "123456"})
	require.ErrorIs(t, err, ErrInvalidCredentials)
}
