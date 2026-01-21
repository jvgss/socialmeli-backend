package service

import (
	"errors"
	"strings"
	"time"

	"socialmeli/internal/domain"
	"socialmeli/internal/store"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("Credenciais inválidas.")
)

type AuthService struct {
	st store.Store
}

func NewAuthService(st store.Store) *AuthService { return &AuthService{st: st} }

type RegisterPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsSeller bool   `json:"is_seller"`
}

type LoginPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *AuthService) Register(p RegisterPayload) (domain.Account, error) {
	name := strings.TrimSpace(p.Name)
	email := strings.TrimSpace(p.Email)
	if err := domain.ValidateTextRequired(name, 40, domain.ErrMaxLen40); err != nil {
		return domain.Account{}, err
	}
	// valida email de forma simples (MVP)
	if email == "" || !strings.Contains(email, "@") {
		return domain.Account{}, errors.New("E-mail inválido.")
	}
	if len([]rune(p.Password)) < 6 {
		return domain.Account{}, errors.New("Senha deve ter pelo menos 6 caracteres.")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(p.Password), bcrypt.DefaultCost)
	if err != nil {
		return domain.Account{}, err
	}
	acc, err := s.st.CreateAccount(name, email, string(hash), p.IsSeller)
	if err != nil {
		return domain.Account{}, err
	}
	// garante created_at em memory
	if acc.CreatedAt.IsZero() {
		acc.CreatedAt = time.Now()
	}
	return acc, nil
}

func (s *AuthService) Login(p LoginPayload) (domain.Account, error) {
	email := strings.TrimSpace(p.Email)
	acc, ok := s.st.GetAccountByEmail(email)
	if !ok {
		return domain.Account{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(p.Password)); err != nil {
		return domain.Account{}, ErrInvalidCredentials
	}
	return acc, nil
}
