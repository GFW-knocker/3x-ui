package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/web/service"
)

const nginxNotFoundPage = "<html>\r\n<head><title>404 Not Found</title></head>\r\n<body>\r\n<center><h1>404 Not Found</h1></center>\r\n<hr><center>nginx</center>\r\n</body>\r\n</html>\r\n"

// A wrong-path probe must read like a stock nginx: the panel's hardening
// headers on a 404 are a fingerprint that survives the secret base path.
func TestWrongPathProbesLookLikeNginx(t *testing.T) {
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
			assertNginxNotFound(t, method, path, resp)
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

func assertNginxNotFound(t *testing.T, method, path string, resp *httptest.ResponseRecorder) {
	t.Helper()
	label := method + " " + path
	if resp.Code != http.StatusNotFound {
		t.Errorf("%s = %d %q, want 404", label, resp.Code, resp.Body.String())
		return
	}
	for _, h := range []string{
		"Content-Security-Policy", "X-Frame-Options", "X-Content-Type-Options",
		"Referrer-Policy", "Strict-Transport-Security", "Cache-Control", "Vary",
		"Content-Encoding", "Set-Cookie",
	} {
		if v := resp.Result().Header.Get(h); v != "" {
			t.Errorf("%s: header %s=%q leaked on a 404", label, h, v)
		}
	}
	if got := resp.Result().Header.Get("Server"); got != "nginx" {
		t.Errorf("%s: Server=%q, want nginx", label, got)
	}
	if got := resp.Result().Header.Get("Content-Type"); got != "text/html" {
		t.Errorf("%s: Content-Type=%q, want text/html", label, got)
	}
	if got := resp.Result().Header.Get("Content-Length"); got != "146" {
		t.Errorf("%s: Content-Length=%q, want 146 (nginx's page)", label, got)
	}
	wantBody := nginxNotFoundPage
	if method == http.MethodHead {
		wantBody = ""
	}
	if body := resp.Body.String(); body != wantBody {
		t.Errorf("%s: body = %q, want %q", label, body, wantBody)
	}
}
