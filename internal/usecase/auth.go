// Package usecase berisi aturan bisnis. Hanya bergantung pada paket domain.
package usecase

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/mofortech/go-fiber-starter/internal/domain"
)

type Auth struct {
	users  domain.UserRepository
	tokens domain.TokenService
}

func NewAuth(u domain.UserRepository, t domain.TokenService) *Auth { return &Auth{users: u, tokens: t} }

func (a *Auth) Register(ctx context.Context, email, name, password string) (*domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &domain.User{Email: strings.ToLower(strings.TrimSpace(email)), Name: strings.TrimSpace(name), PasswordHash: string(hash)}
	if err := a.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Login memeriksa kredensial. Pesan error sama untuk email tak dikenal dan password salah (tidak membocorkan akun).
func (a *Auth) Login(ctx context.Context, email, password string) (token string, u *domain.User, err error) {
	u, err = a.users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, domain.ErrNotFound) {
		return "", nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return "", nil, domain.ErrInvalidCredentials
	}
	token, _, err = a.tokens.Issue(u.ID)
	return token, u, err
}

func (a *Auth) Me(ctx context.Context, userID string) (*domain.User, error) {
	return a.users.ByID(ctx, userID)
}
