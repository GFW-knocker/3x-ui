package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// A probe path answered outside the secret base path lets a scanner confirm
// the panel without knowing that path; every such path must stay a hard 404.
func TestUnauthenticatedProbePathsAre404(t *testing.T) {
	engine := newTestEngine(t)
	for _, path := range []string{
		"/.well-known/appspecific/com.chrome.devtools.json",
	} {
		resp := httptest.NewRecorder()
		engine.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, path, nil))
		if resp.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d %q, want 404", path, resp.Code, resp.Body.String())
		}
	}
}
