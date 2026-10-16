package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ishubhamsingh2e/bourbon/router"
)

func BenchmarkRouterStaticMatching(b *testing.B) {
	r := router.New()
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouterParametricMatching(b *testing.B) {
	r := router.New()
	r.Get("/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := router.URLParam(r, "id")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(id))
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}

func BenchmarkRouterGroupMatching(b *testing.B) {
	r := router.New()
	r.Route("/api/v1", func(api router.Router) {
		api.Route("/organizations", func(orgs router.Router) {
			orgs.Get("/members/{memberId}", func(w http.ResponseWriter, r *http.Request) {
				memberID := router.URLParam(r, "memberId")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(memberID))
			})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/organizations/members/usr_99", nil)
	w := httptest.NewRecorder()

	b.ReportAllocs()

	for b.Loop() {
		r.ServeHTTP(w, req)
	}
}
