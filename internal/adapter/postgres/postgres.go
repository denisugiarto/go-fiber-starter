// Package postgres adalah adapter keluar: repository berbasis pgx + runner migrasi sederhana.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mofortech/go-fiber-starter/internal/domain"
)

func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// Migrate menjalankan berkas *.sql yang belum tercatat, urut nama berkas, masing-masing dalam satu transaksi.
func Migrate(ctx context.Context, pool *pgxpool.Pool, files fs.FS) error {
	if _, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version text PRIMARY KEY, applied_at timestamptz NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)
	for _, name := range names {
		var done bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, name).Scan(&done); err != nil {
			return err
		}
		if done {
			continue
		}
		sql, err := fs.ReadFile(files, name)
		if err != nil {
			return err
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migrasi %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}

// notFound mengubah pgx.ErrNoRows dan ID uuid yang salah format menjadi domain.ErrNotFound.
func notFound(err error) error {
	var pe *pgconn.PgError
	if errors.Is(err, pgx.ErrNoRows) || (errors.As(err, &pe) && pe.Code == "22P02") {
		return domain.ErrNotFound
	}
	return err
}

type Users struct{ db *pgxpool.Pool }

func NewUsers(db *pgxpool.Pool) *Users { return &Users{db} }

func (r *Users) Create(ctx context.Context, u *domain.User) error {
	err := r.db.QueryRow(ctx, `INSERT INTO users (email, name, password_hash) VALUES ($1,$2,$3) RETURNING id, created_at`,
		u.Email, u.Name, u.PasswordHash).Scan(&u.ID, &u.CreatedAt)
	var pe *pgconn.PgError
	if errors.As(err, &pe) && pe.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}

func (r *Users) scan(row pgx.Row) (*domain.User, error) {
	var u domain.User
	if err := row.Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.CreatedAt); err != nil {
		return nil, notFound(err)
	}
	return &u, nil
}

const userCols = `SELECT id, email, name, password_hash, created_at FROM users`

func (r *Users) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, userCols+` WHERE email = $1`, email))
}

func (r *Users) ByID(ctx context.Context, id string) (*domain.User, error) {
	return r.scan(r.db.QueryRow(ctx, userCols+` WHERE id = $1`, id))
}

type Notes struct{ db *pgxpool.Pool }

func NewNotes(db *pgxpool.Pool) *Notes { return &Notes{db} }

const noteCols = `SELECT id, user_id, title, body, created_at, updated_at FROM notes`

func scanNote(row pgx.Row) (*domain.Note, error) {
	var n domain.Note
	if err := row.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return nil, notFound(err)
	}
	return &n, nil
}

func (r *Notes) Create(ctx context.Context, n *domain.Note) error {
	return r.db.QueryRow(ctx, `INSERT INTO notes (user_id, title, body) VALUES ($1,$2,$3) RETURNING id, created_at, updated_at`,
		n.UserID, n.Title, n.Body).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt)
}

func (r *Notes) List(ctx context.Context, userID string) ([]domain.Note, error) {
	rows, err := r.db.Query(ctx, noteCols+` WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Note{}
	for rows.Next() {
		n, err := scanNote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *n)
	}
	return out, rows.Err()
}

func (r *Notes) Get(ctx context.Context, userID, id string) (*domain.Note, error) {
	return scanNote(r.db.QueryRow(ctx, noteCols+` WHERE id = $1 AND user_id = $2`, id, userID))
}

func (r *Notes) Update(ctx context.Context, n *domain.Note) error {
	err := r.db.QueryRow(ctx, `UPDATE notes SET title=$3, body=$4, updated_at=now() WHERE id=$1 AND user_id=$2 RETURNING updated_at`,
		n.ID, n.UserID, n.Title, n.Body).Scan(&n.UpdatedAt)
	return notFound(err)
}

func (r *Notes) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM notes WHERE id=$1 AND user_id=$2`, id, userID)
	if err != nil {
		return notFound(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
