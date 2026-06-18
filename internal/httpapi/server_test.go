package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestValidBearerToken(t *testing.T) {
	t.Parallel()

	if !validBearerToken("Bearer secret-token", "secret-token") {
		t.Fatal("expected valid bearer token")
	}
	if validBearerToken("Bearer wrong-token", "secret-token") {
		t.Fatal("expected invalid bearer token")
	}
	if validBearerToken("", "secret-token") {
		t.Fatal("expected empty header to be invalid")
	}
}

func TestServeEmbeddedSPAFallsBackWithoutRedirect(t *testing.T) {
	t.Parallel()

	webFS := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`<!doctype html><div id="root"></div>`)},
	}
	req := httptest.NewRequest(http.MethodGet, "/PHA6zQkgJRbQJtKZb9XiG0E3", nil)
	rec := httptest.NewRecorder()

	serveEmbeddedSPA(rec, req, webFS)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if location := rec.Header().Get("Location"); location != "" {
		t.Fatalf("expected no redirect location, got %q", location)
	}
	if body := rec.Body.String(); body == "" {
		t.Fatal("expected index HTML body")
	}
}
