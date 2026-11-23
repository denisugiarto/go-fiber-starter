package http_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	httpadapter "github.com/mofortech/go-fiber-starter/internal/adapter/http"
	jwtadapter "github.com/mofortech/go-fiber-starter/internal/adapter/jwt"
	"github.com/mofortech/go-fiber-starter/internal/adapter/memory"
	"github.com/mofortech/go-fiber-starter/internal/usecase"
)

func newApp() *fiber.App {
	tokens := jwtadapter.New("0123456789abcdef0123456789abcdef", time.Hour)
	return httpadapter.New(httpadapter.Deps{
		Auth:        usecase.NewAuth(memory.NewUsers(), tokens),
		Notes:       usecase.NewNote(memory.NewNotes()),
		Tokens:      tokens,
		CORSOrigins: []string{"http://localhost:3000"},
	})
}

func call(t *testing.T, app *fiber.App, method, path, token, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return res.StatusCode, out
}

func TestFlow(t *testing.T) {
	app := newApp()
	reg := `{"email":"budi@example.com","name":"Budi","password":"rahasia123"}`

	if code, _ := call(t, app, "GET", "/health", "", ""); code != http.StatusOK {
		t.Fatalf("health %d", code)
	}
	if code, _ := call(t, app, "POST", "/api/v1/auth/register", "", reg); code != http.StatusCreated {
		t.Fatalf("register %d", code)
	}
	if code, _ := call(t, app, "POST", "/api/v1/auth/register", "", reg); code != http.StatusConflict {
		t.Fatalf("register ganda %d", code)
	}
	code, body := call(t, app, "POST", "/api/v1/auth/register", "", `{"email":"bukan-email","name":"B","password":"pendek"}`)
	if code != http.StatusUnprocessableEntity || body["error"].(map[string]any)["details"] == nil {
		t.Fatalf("validasi %d %v", code, body)
	}
	if code, _ := call(t, app, "POST", "/api/v1/auth/login", "", `{"email":"budi@example.com","password":"salah"}`); code != http.StatusUnauthorized {
		t.Fatalf("login salah %d", code)
	}
	_, login := call(t, app, "POST", "/api/v1/auth/login", "", `{"email":"budi@example.com","password":"rahasia123"}`)
	tok, _ := login["access_token"].(string)
	if tok == "" {
		t.Fatalf("token kosong: %v", login)
	}

	if code, _ := call(t, app, "GET", "/api/v1/me", "", ""); code != http.StatusUnauthorized {
		t.Fatalf("tanpa token %d", code)
	}
	if code, _ := call(t, app, "GET", "/api/v1/me", "token-ngawur", ""); code != http.StatusUnauthorized {
		t.Fatalf("token ngawur %d", code)
	}
	if code, me := call(t, app, "GET", "/api/v1/me", tok, ""); code != http.StatusOK || me["email"] != "budi@example.com" {
		t.Fatalf("me %d %v", code, me)
	}

	code, n := call(t, app, "POST", "/api/v1/notes", tok, `{"title":"Catatan 1","body":"isi"}`)
	if code != http.StatusCreated {
		t.Fatalf("buat catatan %d", code)
	}
	id := n["id"].(string)
	if code, _ := call(t, app, "POST", "/api/v1/notes", tok, `{"title":""}`); code != http.StatusUnprocessableEntity {
		t.Fatalf("judul kosong %d", code)
	}
	if code, u := call(t, app, "PUT", "/api/v1/notes/"+id, tok, `{"title":"Baru","body":"x"}`); code != http.StatusOK || u["title"] != "Baru" {
		t.Fatalf("ubah %d %v", code, u)
	}

	// pengguna lain tidak boleh menyentuh catatan ini
	call(t, app, "POST", "/api/v1/auth/register", "", `{"email":"ani@example.com","name":"Ani","password":"rahasia123"}`)
	_, l2 := call(t, app, "POST", "/api/v1/auth/login", "", `{"email":"ani@example.com","password":"rahasia123"}`)
	tok2 := l2["access_token"].(string)
	for _, m := range []string{"GET", "DELETE"} {
		if code, _ := call(t, app, m, "/api/v1/notes/"+id, tok2, ""); code != http.StatusNotFound {
			t.Fatalf("%s catatan orang lain harus 404, dapat %d", m, code)
		}
	}

	if code, _ := call(t, app, "DELETE", "/api/v1/notes/"+id, tok, ""); code != http.StatusNoContent {
		t.Fatalf("hapus %d", code)
	}
	if code, _ := call(t, app, "GET", "/api/v1/notes/"+id, tok, ""); code != http.StatusNotFound {
		t.Fatalf("setelah hapus %d", code)
	}
}

func TestBodyBukanJSON(t *testing.T) {
	app := newApp()
	for _, body := range []string{"", "bukan json", `{"email":`} {
		if code, _ := call(t, app, "POST", "/api/v1/auth/login", "", body); code != http.StatusBadRequest {
			t.Fatalf("body %q harus 400, dapat %d", body, code)
		}
	}
}
