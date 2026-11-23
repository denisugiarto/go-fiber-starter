package usecase

import (
	"context"
	"strings"

	"github.com/mofortech/go-fiber-starter/internal/domain"
)

// Note adalah contoh CRUD: setiap catatan hanya bisa diakses pemiliknya.
type Note struct{ repo domain.NoteRepository }

func NewNote(r domain.NoteRepository) *Note { return &Note{repo: r} }

func (s *Note) Create(ctx context.Context, userID, title, body string) (*domain.Note, error) {
	n := &domain.Note{UserID: userID, Title: strings.TrimSpace(title), Body: body}
	if err := s.repo.Create(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Note) List(ctx context.Context, userID string) ([]domain.Note, error) {
	return s.repo.List(ctx, userID)
}

func (s *Note) Get(ctx context.Context, userID, id string) (*domain.Note, error) {
	return s.repo.Get(ctx, userID, id)
}

func (s *Note) Update(ctx context.Context, userID, id, title, body string) (*domain.Note, error) {
	n, err := s.repo.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	n.Title, n.Body = strings.TrimSpace(title), body
	if err := s.repo.Update(ctx, n); err != nil {
		return nil, err
	}
	return n, nil
}

func (s *Note) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}
