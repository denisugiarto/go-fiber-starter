// Package http adalah adapter masuk: Fiber v3 (router, middleware, DTO, validasi, format error).
package http

import (
	"context"
	"errors"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/mofortech/go-fiber-starter/docs"
	"github.com/mofortech/go-fiber-starter/internal/domain"
	"github.com/mofortech/go-fiber-starter/internal/usecase"
)

type Deps struct {
	Auth        *usecase.Auth
	Notes       *usecase.Note
	Tokens      domain.TokenService
	CORSOrigins []string
	Ping        func(context.Context) error // cek kesiapan (mis. database); boleh nil
}

func New(d Deps) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: errorHandler, StructValidator: structValidator{validator.New()}})
	app.Use(recover.New(), requestid.New(), cors.New(cors.Config{AllowOrigins: d.CORSOrigins}), accessLog)

	app.Get("/health", func(c fiber.Ctx) error { return c.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/ready", func(c fiber.Ctx) error {
		if d.Ping != nil {
			if err := d.Ping(c.Context()); err != nil {
				slog.Error("ready gagal", "err", err)
				return fiber.NewError(fiber.StatusServiceUnavailable, "belum siap")
			}
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})
	app.Get("/openapi.yaml", func(c fiber.Ctx) error { c.Type("yaml"); return c.Send(docs.OpenAPI) })
	app.Get("/docs", func(c fiber.Ctx) error { c.Type("html"); return c.SendString(swaggerPage) })

	h := &handler{auth: d.Auth, notes: d.Notes}
	v1 := app.Group("/api/v1")
	v1.Post("/auth/register", h.register)
	v1.Post("/auth/login", h.login)

	p := v1.Group("", requireAuth(d.Tokens))
	p.Get("/me", h.me)
	p.Get("/notes", h.listNotes)
	p.Post("/notes", h.createNote)
	p.Get("/notes/:id", h.getNote)
	p.Put("/notes/:id", h.updateNote)
	p.Delete("/notes/:id", h.deleteNote)
	return app
}

func accessLog(c fiber.Ctx) error {
	err := c.Next()
	slog.Info("request", "id", requestid.FromContext(c), "method", c.Method(), "path", c.Path(), "status", c.Response().StatusCode(), "ip", c.IP())
	return err
}

// requireAuth membaca "Authorization: Bearer <token>" dan menyimpan user ID di Locals.
func requireAuth(t domain.TokenService) fiber.Handler {
	return func(c fiber.Ctx) error {
		const prefix = "Bearer "
		h := c.Get("Authorization")
		if len(h) <= len(prefix) || h[:len(prefix)] != prefix {
			return domain.ErrUnauthorized
		}
		uid, err := t.Verify(h[len(prefix):])
		if err != nil {
			return err
		}
		fiber.Locals(c, userKey{}, uid)
		return c.Next()
	}
}

type userKey struct{}

func uid(c fiber.Ctx) string { return fiber.Locals[string](c, userKey{}) }

// --- validasi & error ---

// bind membaca body JSON dan memvalidasinya. Body rusak/bukan JSON -> 400; gagal validasi -> 422 (validationError).
func bind(c fiber.Ctx, out any) error {
	err := c.Bind().Body(out)
	var ve validationError
	if err == nil || errors.As(err, &ve) {
		return err
	}
	return fiber.NewError(fiber.StatusBadRequest, "body request harus berupa JSON yang valid")
}

type structValidator struct{ v *validator.Validate }

type validationError map[string]string

func (validationError) Error() string { return "data tidak valid" }

func (s structValidator) Validate(out any) error {
	err := s.v.Struct(out)
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err
	}
	m := validationError{}
	for _, f := range ve {
		m[f.Field()] = f.Tag()
	}
	return m
}

type errBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func errorHandler(c fiber.Ctx, err error) error {
	var (
		ve validationError
		fe *fiber.Error
	)
	status, b := fiber.StatusInternalServerError, errBody{Code: "internal", Message: "terjadi kesalahan pada server"}
	switch {
	case errors.As(err, &ve):
		status, b = fiber.StatusUnprocessableEntity, errBody{Code: "validation", Message: ve.Error(), Details: ve}
	case errors.Is(err, domain.ErrNotFound):
		status, b = fiber.StatusNotFound, errBody{Code: "not_found", Message: err.Error()}
	case errors.Is(err, domain.ErrConflict):
		status, b = fiber.StatusConflict, errBody{Code: "conflict", Message: err.Error()}
	case errors.Is(err, domain.ErrInvalidCredentials), errors.Is(err, domain.ErrUnauthorized):
		status, b = fiber.StatusUnauthorized, errBody{Code: "unauthorized", Message: err.Error()}
	case errors.As(err, &fe):
		status, b = fe.Code, errBody{Code: "http_error", Message: fe.Message}
	default:
		slog.Error("kesalahan tak terduga", "err", err, "path", c.Path())
	}
	return c.Status(status).JSON(fiber.Map{"error": b})
}

const swaggerPage = `<!doctype html><html><head><meta charset="utf-8"><title>API docs</title>
<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css"></head>
<body><div id="ui"></div><script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>SwaggerUIBundle({url:"/openapi.yaml",dom_id:"#ui"})</script></body></html>`
