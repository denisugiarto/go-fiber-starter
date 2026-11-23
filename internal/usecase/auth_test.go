package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mofortech/go-fiber-starter/internal/adapter/memory"
	"github.com/mofortech/go-fiber-starter/internal/domain"
	"github.com/mofortech/go-fiber-starter/internal/usecase"
)

type fakeTokens struct{}

func (fakeTokens) Issue(id string) (string, time.Time, error) { return "tok-" + id, time.Now(), nil }
func (fakeTokens) Verify(string) (string, error)              { return "", domain.ErrUnauthorized }

func TestAuth(t *testing.T) {
	ctx := context.Background()
	a := usecase.NewAuth(memory.NewUsers(), fakeTokens{})

	u, err := a.Register(ctx, "  Budi@Example.com ", "Budi", "rahasia123")
	if err != nil || u.Email != "budi@example.com" || u.PasswordHash == "rahasia123" {
		t.Fatalf("register: %+v %v", u, err)
	}
	if _, err := a.Register(ctx, "budi@example.com", "Budi 2", "rahasia123"); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("email ganda harus ErrConflict, dapat %v", err)
	}
	if tok, _, err := a.Login(ctx, "BUDI@example.com", "rahasia123"); err != nil || tok != "tok-"+u.ID {
		t.Fatalf("login benar: %q %v", tok, err)
	}
	for _, c := range [][2]string{{"budi@example.com", "salah"}, {"tidak@ada.com", "rahasia123"}} {
		if _, _, err := a.Login(ctx, c[0], c[1]); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("login %v harus ErrInvalidCredentials, dapat %v", c, err)
		}
	}
}
