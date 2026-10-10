package link

import (
	"encoding/base64"
	"reflect"
	"testing"
)

// Imported QUIC links must rebuild the quicSettings Mahsa-Core reads: vless/trojan carry
// quicSecurity/key/headerType params, vmess carries them as host/path/type.
func TestParseLink_QuicStream(t *testing.T) {
	want := map[string]any{
		"security": "chacha20-poly1305",
		"key":      "s3cret",
		"header":   map[string]any{"type": "wechat-video"},
	}
	vmessJSON := `{"v":"2","ps":"q","add":"q.example.com","port":443,"id":"11111111-2222-4333-8444-555555555555",` +
		`"net":"quic","type":"wechat-video","host":"chacha20-poly1305","path":"s3cret","tls":"tls","sni":"q.example.com"}`
	for name, link := range map[string]string{
		"vless": "vless://11111111-2222-4333-8444-555555555555@q.example.com:443?type=quic&security=tls" +
			"&sni=q.example.com&quicSecurity=chacha20-poly1305&key=s3cret&headerType=wechat-video#q",
		"vmess": "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON)),
	} {
		t.Run(name, func(t *testing.T) {
			res, err := ParseLink(link)
			if err != nil {
				t.Fatalf("ParseLink: %v", err)
			}
			stream, _ := res.Outbound["streamSettings"].(map[string]any)
			if stream["network"] != "quic" {
				t.Fatalf("network = %v, want quic", stream["network"])
			}
			if got := stream["quicSettings"]; !reflect.DeepEqual(got, want) {
				t.Fatalf("quicSettings = %#v, want %#v", got, want)
			}
		})
	}
}

// A bare type=quic link names no cipher or header, which the core reads as none.
func TestParseLink_QuicDefaults(t *testing.T) {
	res, err := ParseLink("vless://11111111-2222-4333-8444-555555555555@q.example.com:443?type=quic&security=tls#q")
	if err != nil {
		t.Fatalf("ParseLink: %v", err)
	}
	stream, _ := res.Outbound["streamSettings"].(map[string]any)
	want := map[string]any{"security": "none", "key": "", "header": map[string]any{"type": "none"}}
	if got := stream["quicSettings"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("quicSettings = %#v, want %#v", got, want)
	}
}
