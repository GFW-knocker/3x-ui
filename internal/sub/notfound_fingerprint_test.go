package sub

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const nginxNotFoundPage = "<html>\r\n<head><title>404 Not Found</title></head>\r\n<body>\r\n<center><h1>404 Not Found</h1></center>\r\n<hr><center>nginx</center>\r\n</body>\r\n</html>\r\n"

// The sub port answers wrong paths with Go's "404 page not found", which names
// the runtime; it must read like a stock nginx instead.
func TestSubWrongPathProbesLookLikeNginx(t *testing.T) {
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
			if got := resp.Result().Header.Get("Server"); got != "nginx" {
				t.Errorf("%s: Server=%q, want nginx", label, got)
			}
			if got := resp.Result().Header.Get("Content-Type"); got != "text/html" {
				t.Errorf("%s: Content-Type=%q, want text/html", label, got)
			}
			if got := resp.Result().Header.Get("Content-Length"); got != "146" {
				t.Errorf("%s: Content-Length=%q, want 146", label, got)
			}
			if method == http.MethodGet && resp.Body.String() != nginxNotFoundPage {
				t.Errorf("%s: body = %q, want nginx's 404 page", label, resp.Body.String())
			}
		}
	}
}
