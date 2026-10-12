package sub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/GFW-knocker/3x-ui/v3/internal/web/service"
)

// A directory listing of the Vite bundle on the sub port identifies the panel
// to scanners that never learned the subscription path (#assets-listing).
func TestSubAssetsNeverListDirectories(t *testing.T) {
	initSubDB(t)
	oldDistFS := distFS
	distFS = fstest.MapFS{
		"dist/subpage.html":             {Data: []byte(`<!doctype html><html><head></head><body></body></html>`)},
		"dist/assets/index-ab12cd34.js": {Data: []byte("console.log('x')")},
		"dist/assets/img/logo.png":      {Data: []byte("png")},
	}
	t.Cleanup(func() { distFS = oldDistFS })

	subPath, err := (&service.SettingService{}).GetSubPath()
	if err != nil {
		t.Fatal(err)
	}
	router, err := (&Server{}).initRouter()
	if err != nil {
		t.Fatal(err)
	}
	under := func(rest string) string { return strings.TrimRight(subPath, "/") + rest }
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, path, nil))
		return resp
	}

	for _, path := range []string{"/assets/index-ab12cd34.js", under("/assets/index-ab12cd34.js"), under("/abc/assets/img/logo.png")} {
		if resp := get(path); resp.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 (files must still be served)", path, resp.Code)
		}
	}
	for _, path := range []string{"/assets/", "/assets", "/assets/img/", under("/assets/"), under("/abc/assets/img/")} {
		resp := get(path)
		if resp.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, resp.Code)
		}
		if body := resp.Body.String(); len(body) > 0 && (contains(body, "index-ab12cd34.js") || contains(body, "logo.png")) {
			t.Errorf("GET %s leaked a directory listing: %q", path, body)
		}
	}
}

func contains(s, sub string) bool { return strings.Contains(s, sub) }
