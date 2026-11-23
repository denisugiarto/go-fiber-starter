// Package memory berisi repository in-memory. Dipakai untuk tes (tanpa database), bukan untuk produksi.
package memory

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/mofortech/go-fiber-starter/internal/domain"
)

var seq struct {
	sync.Mutex
	n int
}

func nextID() string { seq.Lock(); defer seq.Unlock(); seq.n++; return strconv.Itoa(seq.n) }

type Users struct {
	mu sync.Mutex
	m  map[string]domain.User
}

func NewUsers() *Users { return &Users{m: map[string]domain.User{}} }

func (r *Users) Create(_ context.Context, u *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.m {
		if x.Email == u.Email {
			return domain.ErrConflict
		}
	}
	u.ID, u.CreatedAt = nextID(), time.Now()
	r.m[u.ID] = *u
	return nil
}

func (r *Users) ByEmail(_ context.Context, email string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, x := range r.m {
		if x.Email == email {
			return &x, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (r *Users) ByID(_ context.Context, id string) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x, ok := r.m[id]; ok {
		return &x, nil
	}
	return nil, domain.ErrNotFound
}

type Notes struct {
	mu sync.Mutex
	m  map[string]domain.Note
}

func NewNotes() *Notes { return &Notes{m: map[string]domain.Note{}} }

func (r *Notes) Create(_ context.Context, n *domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	n.ID, n.CreatedAt, n.UpdatedAt = nextID(), time.Now(), time.Now()
	r.m[n.ID] = *n
	return nil
}

func (r *Notes) List(_ context.Context, userID string) ([]domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := []domain.Note{}
	for _, n := range r.m {
		if n.UserID == userID {
			out = append(out, n)
		}
	}
	return out, nil
}

func (r *Notes) Get(_ context.Context, userID, id string) (*domain.Note, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.m[id]; ok && n.UserID == userID {
		return &n, nil
	}
	return nil, domain.ErrNotFound
}

func (r *Notes) Update(_ context.Context, n *domain.Note) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if x, ok := r.m[n.ID]; ok && x.UserID == n.UserID {
		n.UpdatedAt = time.Now()
		r.m[n.ID] = *n
		return nil
	}
	return domain.ErrNotFound
}

func (r *Notes) Delete(_ context.Context, userID, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n, ok := r.m[id]; ok && n.UserID == userID {
		delete(r.m, id)
		return nil
	}
	return domain.ErrNotFound
}
