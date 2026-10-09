package xray

import (
	"testing"

	"github.com/GFW-knocker/Xray-core/proxy/mvless"
	"github.com/GFW-knocker/Xray-core/proxy/vless"
	"google.golang.org/protobuf/proto"
)

// Mahsa-Core's MVLESS inbound only accepts its own account type for live AddUser, so the
// account must be typed mvless while carrying every field the VLESS builder fills.
func TestBuildUserAccountMvlessMirrorsVless(t *testing.T) {
	user := map[string]any{
		"email":    "mv@example.test",
		"id":       "5f2eb9d6-3a2f-4a55-9812-6ea1e2f7a333",
		"flow":     "xtls-rprx-vision",
		"reverse":  map[string]any{"tag": "portal"},
		"testseed": []any{float64(900), float64(500), float64(900), float64(256)},
		"testpre":  float64(2),
	}
	vlessTM, err := buildUserAccount("vless", user)
	if err != nil {
		t.Fatalf("buildUserAccount(vless): %v", err)
	}
	mvlessTM, err := buildUserAccount("mvless", user)
	if err != nil {
		t.Fatalf("buildUserAccount(mvless): %v", err)
	}
	if mvlessTM == nil {
		t.Fatal("buildUserAccount returned no account for mvless")
	}
	if want := "xray.proxy.mvless.Account"; mvlessTM.Type != want {
		t.Fatalf("account type = %q, want %q", mvlessTM.Type, want)
	}

	v := new(vless.Account)
	if err := proto.Unmarshal(vlessTM.Value, v); err != nil {
		t.Fatalf("unmarshal vless account: %v", err)
	}
	mv := new(mvless.Account)
	if err := proto.Unmarshal(mvlessTM.Value, mv); err != nil {
		t.Fatalf("unmarshal mvless account: %v", err)
	}
	if mv.Id != v.Id || mv.Flow != v.Flow || mv.Testpre != v.Testpre ||
		mv.GetReverse().GetTag() != v.GetReverse().GetTag() || len(mv.Testseed) != len(v.Testseed) {
		t.Fatalf("mvless account %+v does not mirror vless account %+v", mv, v)
	}
	for i := range v.Testseed {
		if mv.Testseed[i] != v.Testseed[i] {
			t.Fatalf("testseed = %v, want %v", mv.Testseed, v.Testseed)
		}
	}
	if v.Id == "" || v.Flow == "" || v.GetReverse().GetTag() == "" || v.Testpre == 0 || len(v.Testseed) != 4 {
		t.Fatalf("vless fixture lost a field, so the comparison proves nothing: %+v", v)
	}
}
