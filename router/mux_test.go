package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type mockResourceController struct{}

func (m *mockResourceController) Index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"action": "index"})
}

func (m *mockResourceController) Show(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"action": "show", "id": URLParam(r, "id")})
}

func (m *mockResourceController) Create(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"action": "create"})
}

func (m *mockResourceController) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"action": "update", "id": URLParam(r, "id")})
}

func (m *mockResourceController) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"action": "delete", "id": URLParam(r, "id")})
}

func TestChiStyleRoutingAndPatterns(t *testing.T) {
	r := New()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("welcome"))
	})

	r.Route("/articles", func(r Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("list-articles"))
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("create-article"))
		})

		r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("search-articles"))
		})

		// Regex parameter pattern: slug must be lowercase letters and dashes
		r.Route("/{articleSlug:[a-z-]+}", func(r Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				slug := URLParam(r, "articleSlug")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("slug:" + slug))
			})
		})

		// Numeric parameter pattern: id must be numbers
		r.Route("/{articleID:[0-9]+}", func(r Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				id := URLParam(r, "articleID")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("id:" + id))
			})
			r.Put("/", func(w http.ResponseWriter, r *http.Request) {
				id := URLParam(r, "articleID")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("updated:" + id))
			})
		})
	})

	// Wildcard catchall
	r.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
		path := URLParam(r, "*")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("file:" + path))
	})

	// 1. GET /
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "welcome" {
		t.Errorf("GET / failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 2. GET /articles
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/articles", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "list-articles" {
		t.Errorf("GET /articles failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 3. GET /articles/search
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/articles/search", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "search-articles" {
		t.Errorf("GET /articles/search failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 4. GET /articles/hello-world (matches slug)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/articles/hello-world", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "slug:hello-world" {
		t.Errorf("GET /articles/hello-world failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 5. GET /articles/12345 (matches numeric ID)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/articles/12345", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "id:12345" {
		t.Errorf("GET /articles/12345 failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 6. GET /files/images/banner.png (wildcard match)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/files/images/banner.png", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "file:images/banner.png" {
		t.Errorf("GET /files/* failed: code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestChiStyleWithAndGroupMiddleware(t *testing.T) {
	r := New()

	var logs []string

	globalMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logs = append(logs, "global")
			next.ServeHTTP(w, r)
		})
	}

	inlineMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logs = append(logs, "inline")
			next.ServeHTTP(w, r)
		})
	}

	authMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logs = append(logs, "auth")
			next.ServeHTTP(w, r)
		})
	}

	r.Use(globalMW)

	// Unauthenticated route
	r.Get("/public", func(w http.ResponseWriter, r *http.Request) {
		logs = append(logs, "public_handler")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("public"))
	})

	// Inline With middleware
	r.With(inlineMW).Get("/inline-protected", func(w http.ResponseWriter, r *http.Request) {
		logs = append(logs, "inline_handler")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("inline"))
	})

	// Group with scoped middleware
	r.Group(func(admin Router) {
		admin.Use(authMW)
		admin.Get("/dashboard", func(w http.ResponseWriter, r *http.Request) {
			logs = append(logs, "admin_handler")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("admin"))
		})
	})

	// 1. Test /public -> global, public_handler
	logs = nil
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/public", nil)
	r.ServeHTTP(w, req)
	if strings.Join(logs, ",") != "global,public_handler" {
		t.Errorf("unexpected logs for /public: %v", logs)
	}

	// 2. Test /inline-protected -> global, inline, inline_handler
	logs = nil
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/inline-protected", nil)
	r.ServeHTTP(w, req)
	if strings.Join(logs, ",") != "global,inline,inline_handler" {
		t.Errorf("unexpected logs for /inline-protected: %v", logs)
	}

	// 3. Test /dashboard -> global, auth, admin_handler
	logs = nil
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	r.ServeHTTP(w, req)
	if strings.Join(logs, ",") != "global,auth,admin_handler" {
		t.Errorf("unexpected logs for /dashboard: %v", logs)
	}
}

func TestMountSubRouter(t *testing.T) {
	r := New()

	admin := New()
	admin.Get("/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("admin-users"))
	})

	r.Mount("/admin", admin)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "admin-users" {
		t.Errorf("Mount failed: code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestResourceAndVersioning(t *testing.T) {
	r := New()

	Resource(r, "/items", &mockResourceController{})
	Version(r, "v1", func(v Router) {
		v.Get("/status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("v1-ok"))
		})
	})

	// Test Resource GET /items/99
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/items/99", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"99"`) {
		t.Errorf("Resource show failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// Test Version GET /api/v1/status
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != "v1-ok" {
		t.Errorf("Version failed: code=%d body=%s", w.Code, w.Body.String())
	}
}

func TestCustomNotFoundAndMethodNotAllowed(t *testing.T) {
	r := New()
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "custom 404", http.StatusNotFound)
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "custom 405", http.StatusMethodNotAllowed)
	})

	r.Get("/exist", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 404
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/not-found", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound || strings.TrimSpace(w.Body.String()) != "custom 404" {
		t.Errorf("NotFound failed: code=%d body=%s", w.Code, w.Body.String())
	}

	// 405
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/exist", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed || strings.TrimSpace(w.Body.String()) != "custom 405" {
		t.Errorf("MethodNotAllowed failed: code=%d body=%s", w.Code, w.Body.String())
	}
}
