// Package jwt mengimplementasikan domain.TokenService dengan JWT HS256.
package jwt

import (
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"

	"github.com/mofortech/go-fiber-starter/internal/domain"
)

type Service struct {
	secret []byte
	ttl    time.Duration
}

func New(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl}
}

func (s *Service) Issue(userID string) (string, time.Time, error) {
	exp := time.Now().Add(s.ttl)
	t := gojwt.NewWithClaims(gojwt.SigningMethodHS256, gojwt.RegisteredClaims{Subject: userID, ExpiresAt: gojwt.NewNumericDate(exp), IssuedAt: gojwt.NewNumericDate(time.Now())})
	tok, err := t.SignedString(s.secret)
	return tok, exp, err
}

func (s *Service) Verify(token string) (string, error) {
	var c gojwt.RegisteredClaims
	_, err := gojwt.ParseWithClaims(token, &c, func(*gojwt.Token) (any, error) { return s.secret, nil },
		gojwt.WithValidMethods([]string{"HS256"}), gojwt.WithExpirationRequired())
	if err != nil || c.Subject == "" {
		return "", domain.ErrUnauthorized
	}
	return c.Subject, nil
}
