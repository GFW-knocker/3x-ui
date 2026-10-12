package sub

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const stockNotFoundBody = "404 page not found\n"

// gin's default miss is a near-copy of net/http's that drops the charset and
// nosniff header, a known gin tell; the sub port must match stock Go exactly.
func TestSubWrongPathProbesLookLikeStockGo(t *testing.T) {
	initSubDB(t)
	oldDistFS := distFS
	distFS = testDistFS
	t.Cleanup(func() { distFS = oldDistFS })
	router, err := (&Server{}).initRouter()
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{http.MethodGet, http.MethodHead} {
		for _, path := range []string{"/", "/nope", "/assets", "/assets/"} {
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)
			label := method + " " + path
			if resp.Code != http.StatusNotFound {
				t.Errorf("%s = %d %q, want 404", label, resp.Code, resp.Body.String())
				continue
			}
			if got := resp.Result().Header.Get("Server"); got != "" {
				t.Errorf("%s: Server=%q, stock Go sends none", label, got)
			}
			if got := resp.Result().Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
				t.Errorf("%s: Content-Type=%q, want stock text/plain; charset=utf-8", label, got)
			}
			if got := resp.Result().Header.Get("X-Content-Type-Options"); got != "nosniff" {
				t.Errorf("%s: X-Content-Type-Options=%q, want nosniff", label, got)
			}
			if method == http.MethodGet && resp.Body.String() != stockNotFoundBody {
				t.Errorf("%s: body = %q, want %q", label, resp.Body.String(), stockNotFoundBody)
			}
		}
	}
}
