package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/web/service"
)

const stockNotFoundBody = "404 page not found\n"

// A wrong-path probe must read like a stock net/http miss: the panel's hardening
// headers on a 404 are a fingerprint that survives the secret base path.
func TestWrongPathProbesLookLikeStockGo(t *testing.T) {
	engine := newTestEngine(t)
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodPost} {
		for _, path := range []string{
			"/definitely-not-here",
			"/.well-known/appspecific/com.chrome.devtools.json",
			"/assets/index.js",
		} {
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("Accept-Encoding", "gzip")
			resp := httptest.NewRecorder()
			engine.ServeHTTP(resp, req)
			assertStockNotFound(t, method, path, resp)
		}
	}
}

// The hardening headers must survive on real panel routes; the generic 404 is
// scoped to misses, not a blanket removal.
func TestPanelRoutesKeepSecurityHeaders(t *testing.T) {
	engine := newTestEngine(t)
	basePath, err := (&service.SettingService{}).GetBasePath()
	if err != nil {
		t.Fatal(err)
	}
	target := basePath + "csrf-token"
	resp := httptest.NewRecorder()
	engine.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, target, nil))
	if resp.Code != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", target, resp.Code)
	}
	if resp.Result().Header.Get("Content-Security-Policy") == "" || resp.Result().Header.Get("X-Frame-Options") != "DENY" {
		t.Fatalf("panel route lost its hardening headers: %v", resp.Result().Header)
	}
	if resp.Result().Header.Get("Server") != "" {
		t.Fatalf("panel route should not advertise a Server header, got %q", resp.Result().Header.Get("Server"))
	}
}

func assertStockNotFound(t *testing.T, method, path string, resp *httptest.ResponseRecorder) {
	t.Helper()
	label := method + " " + path
	if resp.Code != http.StatusNotFound {
		t.Errorf("%s = %d %q, want 404", label, resp.Code, resp.Body.String())
		return
	}
	for _, h := range []string{
		"Content-Security-Policy", "X-Frame-Options", "Referrer-Policy",
		"Strict-Transport-Security", "Cache-Control", "Vary", "Content-Encoding",
		"Set-Cookie", "Server",
	} {
		if v := resp.Result().Header.Get(h); v != "" {
			t.Errorf("%s: header %s=%q leaked on a 404", label, h, v)
		}
	}
	if got := resp.Result().Header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("%s: Content-Type=%q, want stock text/plain; charset=utf-8", label, got)
	}
	if got := resp.Result().Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("%s: X-Content-Type-Options=%q, want nosniff (http.Error sets it)", label, got)
	}
	if got := resp.Result().Header.Get("Content-Length"); got != "" {
		t.Errorf("%s: handler set Content-Length=%q; stock Go leaves it to the server", label, got)
	}
	// http.NotFound writes the body for HEAD too; the real server discards it,
	// the recorder keeps it, exactly as a stock ServeMux behaves under test.
	if body := resp.Body.String(); body != stockNotFoundBody {
		t.Errorf("%s: body = %q, want %q", label, body, stockNotFoundBody)
	}
}
