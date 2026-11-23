package http

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/mofortech/go-fiber-starter/internal/domain"
	"github.com/mofortech/go-fiber-starter/internal/usecase"
)

type handler struct {
	auth  *usecase.Auth
	notes *usecase.Note
}

type userResp struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func toUser(u *domain.User) userResp {
	return userResp{ID: u.ID, Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt}
}

type noteResp struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toNote(n *domain.Note) noteResp {
	return noteResp{ID: n.ID, Title: n.Title, Body: n.Body, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt}
}

// bcrypt hanya memakai 72 byte pertama, jadi password dibatasi 72.
type registerReq struct {
	Email    string `json:"email" validate:"required,email,max=254"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type loginReq struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type noteReq struct {
	Title string `json:"title" validate:"required,max=200"`
	Body  string `json:"body" validate:"max=10000"`
}

func (h *handler) register(c fiber.Ctx) error {
	var r registerReq
	if err := bind(c, &r); err != nil {
		return err
	}
	u, err := h.auth.Register(c.Context(), r.Email, r.Name, r.Password)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(toUser(u))
}

func (h *handler) login(c fiber.Ctx) error {
	var r loginReq
	if err := bind(c, &r); err != nil {
		return err
	}
	tok, u, err := h.auth.Login(c.Context(), r.Email, r.Password)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"access_token": tok, "token_type": "Bearer", "user": toUser(u)})
}

func (h *handler) me(c fiber.Ctx) error {
	u, err := h.auth.Me(c.Context(), uid(c))
	if err != nil {
		return err
	}
	return c.JSON(toUser(u))
}

func (h *handler) listNotes(c fiber.Ctx) error {
	ns, err := h.notes.List(c.Context(), uid(c))
	if err != nil {
		return err
	}
	out := make([]noteResp, len(ns))
	for i := range ns {
		out[i] = toNote(&ns[i])
	}
	return c.JSON(out)
}

func (h *handler) createNote(c fiber.Ctx) error {
	var r noteReq
	if err := bind(c, &r); err != nil {
		return err
	}
	n, err := h.notes.Create(c.Context(), uid(c), r.Title, r.Body)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(toNote(n))
}

func (h *handler) getNote(c fiber.Ctx) error {
	n, err := h.notes.Get(c.Context(), uid(c), c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(toNote(n))
}

func (h *handler) updateNote(c fiber.Ctx) error {
	var r noteReq
	if err := bind(c, &r); err != nil {
		return err
	}
	n, err := h.notes.Update(c.Context(), uid(c), c.Params("id"), r.Title, r.Body)
	if err != nil {
		return err
	}
	return c.JSON(toNote(n))
}

func (h *handler) deleteNote(c fiber.Ctx) error {
	if err := h.notes.Delete(c.Context(), uid(c), c.Params("id")); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}
