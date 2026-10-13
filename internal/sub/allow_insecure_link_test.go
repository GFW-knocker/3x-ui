package sub

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
	"github.com/GFW-knocker/3x-ui/v3/internal/util/link"
)

// Mahsa-Core kept allowInsecure after upstream Xray dropped it, so the inbound's
// TLS client switch must reach every exported shape or clients cannot use
// self-signed certificates. Each case flips the switch and checks the output.

func tlsStream(allowInsecure bool) string {
	return `{"network":"tcp","security":"tls","tcpSettings":{"header":{"type":"none"}},` +
		`"tlsSettings":{"serverName":"srv.example.com","alpn":["h2"],"settings":{"fingerprint":"chrome",` +
		`"allowInsecure":` + map[bool]string{true: "true", false: "false"}[allowInsecure] + `}}}`
}

func TestAllowInsecureReachesVlessAndTrojanLinks(t *testing.T) {
	for _, protocol := range []model.Protocol{model.VLESS, model.MVLESS, model.Trojan} {
		for _, allow := range []bool{true, false} {
			in := &model.Inbound{
				Id: 950001, Listen: "203.0.113.1", Port: 8443, Protocol: protocol,
				Settings:       `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","password":"pw","email":"user"}]}`,
				StreamSettings: tlsStream(allow),
			}
			var got string
			if protocol.IsVLESSFamily() {
				got = (&SubService{}).genVlessLink(in, "user")
			} else {
				got = (&SubService{}).genTrojanLink(in, "user")
			}
			if got == "" {
				t.Fatalf("%s: no link generated", protocol)
			}
			if !strings.HasPrefix(got, string(protocol)+"://") {
				t.Fatalf("%s: link uses the wrong scheme: %q", protocol, got)
			}
			if has := strings.Contains(got, "allowInsecure=1"); has != allow {
				t.Errorf("%s allowInsecure=%v: link %q, allowInsecure=1 present=%v", protocol, allow, got, has)
			}
			// Readability: the flag follows the security it relaxes, never leads the query.
			if allow && !strings.Contains(got, "security=tls&allowInsecure=1") {
				t.Errorf("%s: allowInsecure is not placed right after security=tls: %q", protocol, got)
			}
		}
	}
}

func TestAllowInsecureReachesVmessObject(t *testing.T) {
	in := &model.Inbound{
		Id: 950002, Listen: "203.0.113.1", Port: 8443, Protocol: model.VMESS,
		Settings:       `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user"}]}`,
		StreamSettings: tlsStream(true),
	}
	got := (&SubService{}).genVmessLink(in, "user")
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(got, "vmess://"))
	if err != nil {
		t.Fatalf("vmess link not base64: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("vmess object: %v", err)
	}
	if obj["allowInsecure"] != true {
		t.Fatalf("vmess object lacks allowInsecure: %v", obj)
	}
}

func TestAllowInsecureReachesHysteriaLink(t *testing.T) {
	in := &model.Inbound{
		Id: 950003, Listen: "203.0.113.1", Port: 443, Protocol: model.Hysteria,
		Settings:       `{"version":2,"clients":[{"auth":"secret","email":"user"}]}`,
		StreamSettings: `{"security":"tls","tlsSettings":{"serverName":"srv.example.com","settings":{"allowInsecure":true}}}`,
	}
	got := (&SubService{}).genHysteriaLink(in, "user")
	if !strings.Contains(got, "insecure=1") {
		t.Fatalf("hysteria link lacks insecure=1: %q", got)
	}
}

func TestAllowInsecureReachesJSONSubscriptionTLS(t *testing.T) {
	for _, allow := range []bool{true, false} {
		tlsData := map[string]any{}
		putClientTLSVerification(tlsData, map[string]any{"fingerprint": "chrome", "allowInsecure": allow})
		if _, ok := tlsData["allowInsecure"]; ok != allow {
			t.Errorf("allowInsecure=%v: xray-level tlsSettings = %v", allow, tlsData)
		}
	}
}

func TestAllowInsecureImportedFromLinks(t *testing.T) {
	for name, raw := range map[string]string{
		"vless":    "vless://11111111-2222-4333-8444-555555555555@srv.example.com:8443?type=tcp&security=tls&sni=srv.example.com&allowInsecure=1#r",
		"hysteria": "hysteria2://secret@srv.example.com:443?sni=srv.example.com&insecure=1#r",
	} {
		parsed, err := link.ParseLink(raw)
		if err != nil {
			t.Fatalf("%s: ParseLink: %v", name, err)
		}
		stream, _ := parsed.Outbound["streamSettings"].(map[string]any)
		tls, _ := stream["tlsSettings"].(map[string]any)
		if tls["allowInsecure"] != true {
			t.Errorf("%s: imported tlsSettings.allowInsecure = %#v, want true", name, tls["allowInsecure"])
		}
	}
}
