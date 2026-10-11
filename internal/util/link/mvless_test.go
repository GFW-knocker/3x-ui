package link

import (
	"encoding/json"
	"strings"
	"testing"
)

// An mvless:// link is a vless:// link for Mahsa-Core's MVLESS, so it must parse into the
// same outbound under the mvless protocol, with an identity distinct from the vless twin.
func TestParseLink_Mvless(t *testing.T) {
	query := "?type=ws&security=tls&host=H&path=%2Fws&sni=SNI&fp=chrome&flow=&encryption=none#node"
	vless, err := ParseLink("vless://11111111-2222-4333-8444-555555555555@h.example.com:443" + query)
	if err != nil {
		t.Fatalf("ParseLink(vless): %v", err)
	}
	mvless, err := ParseLink("mvless://11111111-2222-4333-8444-555555555555@h.example.com:443" + query)
	if err != nil {
		t.Fatalf("ParseLink(mvless): %v", err)
	}

	if got := mvless.Outbound["protocol"]; got != "mvless" {
		t.Fatalf("protocol = %v, want mvless", got)
	}
	vlessJSON, _ := json.Marshal(vless.Outbound)
	mvlessJSON, _ := json.Marshal(mvless.Outbound)
	if want := strings.Replace(string(vlessJSON), `"protocol":"vless"`, `"protocol":"mvless"`, 1); string(mvlessJSON) != want {
		t.Fatalf("mvless outbound\n got: %s\nwant: %s", mvlessJSON, want)
	}
	if mvless.Identity == vless.Identity {
		t.Fatalf("mvless and vless links share identity %q; subscription tags would collide", vless.Identity)
	}
}
