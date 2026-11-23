// Package domain berisi entitas, error, dan port (interface) yang dipakai usecase.
// Tidak boleh mengimpor paket lain di proyek ini.
package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound           = errors.New("data tidak ditemukan")
	ErrConflict           = errors.New("data sudah ada")
	ErrInvalidCredentials = errors.New("email atau password salah")
	ErrUnauthorized       = errors.New("token tidak valid atau kedaluwarsa")
)

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
}

type Note struct {
	ID        string
	UserID    string
	Title     string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Port keluar (diimplementasikan adapter).
type UserRepository interface {
	Create(ctx context.Context, u *User) error // ErrConflict bila email sudah dipakai
	ByEmail(ctx context.Context, email string) (*User, error)
	ByID(ctx context.Context, id string) (*User, error)
}

type NoteRepository interface {
	Create(ctx context.Context, n *Note) error
	List(ctx context.Context, userID string) ([]Note, error)
	Get(ctx context.Context, userID, id string) (*Note, error)
	Update(ctx context.Context, n *Note) error
	Delete(ctx context.Context, userID, id string) error
}

// TokenService menerbitkan dan memeriksa token akses; mengembalikan user ID.
type TokenService interface {
	Issue(userID string) (token string, expiresAt time.Time, err error)
	Verify(token string) (userID string, err error)
}
