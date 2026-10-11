package sub

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// Mahsa-Core keeps xray's QUIC transport; links carry it as v2rayN and the pre-v3 MX-UI did:
// quicSecurity/key/headerType query params, and type/host/path in the vmess JSON object.
const quicStream = `{"network":"quic","security":"tls","tlsSettings":{"serverName":"q.example.com"},
	"quicSettings":{"security":"chacha20-poly1305","key":"s3cret","header":{"type":"wechat-video"}}}`

func TestGenVlessLink_QuicParams(t *testing.T) {
	link := (&SubService{}).genVlessLink(shareLinkInbound(quicStream), "user")
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse %q: %v", link, err)
	}
	query := u.Query()
	for key, want := range map[string]string{
		"type": "quic", "quicSecurity": "chacha20-poly1305", "key": "s3cret", "headerType": "wechat-video",
	} {
		if got := query.Get(key); got != want {
			t.Errorf("%s = %q, want %q (link %s)", key, got, want, link)
		}
	}
}

func TestGenVmessLink_QuicObject(t *testing.T) {
	in := &model.Inbound{
		Listen: "203.0.113.1", Port: 443, Protocol: model.VMESS, Remark: "q",
		Settings:       `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user","security":"auto"}]}`,
		StreamSettings: quicStream,
	}
	link := (&SubService{}).genVmessLink(in, "user")
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(link, "vmess://"))
	if err != nil {
		t.Fatalf("vmess link is not base64: %q", link)
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("vmess link is not json: %s", raw)
	}
	for key, want := range map[string]string{
		"net": "quic", "type": "wechat-video", "host": "chacha20-poly1305", "path": "s3cret",
	} {
		if got := obj[key]; got != want {
			t.Errorf("%s = %v, want %q (object %s)", key, got, want, raw)
		}
	}
}
