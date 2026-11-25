package link

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/GFW-knocker/Xray-core/infra/conf"
)

// buildCoreStream runs a parsed streamSettings through conf.StreamConfig, the loader the
// core calls at startup, so an imported link that the core refuses fails here.
func buildCoreStream(t *testing.T, stream map[string]any) {
	t.Helper()
	raw, err := json.Marshal(stream)
	if err != nil {
		t.Fatalf("marshal stream: %v", err)
	}
	var sc conf.StreamConfig
	if err := json.Unmarshal(raw, &sc); err != nil {
		t.Fatalf("unmarshal stream %s: %v", raw, err)
	}
	if _, err := sc.Build(); err != nil {
		t.Fatalf("the core refuses the imported stream %s: %v", raw, err)
	}
}

func TestParseLink_QuicTCPStream(t *testing.T) {
	want := map[string]any{
		"security": "aes-128-gcm",
		"key":      "k1",
		"header":   map[string]any{"type": "srtp"},
		"password": "lanepw",
		"ports":    "3000-3002",
	}
	vmessJSON := `{"v":"2","ps":"q","add":"q.example.com","port":443,"id":"11111111-2222-4333-8444-555555555555",` +
		`"net":"quictcp","type":"srtp","host":"aes-128-gcm","path":"k1","lanePassword":"lanepw","lanePorts":"3000-3002"}`
	for name, link := range map[string]string{
		"vless": "vless://11111111-2222-4333-8444-555555555555@q.example.com:443?type=quictcp&security=none" +
			"&quicSecurity=aes-128-gcm&key=k1&headerType=srtp&lanePassword=lanepw&lanePorts=3000-3002#q",
		"vmess": "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON)),
	} {
		t.Run(name, func(t *testing.T) {
			res, err := ParseLink(link)
			if err != nil {
				t.Fatalf("ParseLink: %v", err)
			}
			stream, _ := res.Outbound["streamSettings"].(map[string]any)
			if stream["network"] != "quictcp" {
				t.Fatalf("network = %v, want quictcp", stream["network"])
			}
			if got := stream["quictcpSettings"]; !reflect.DeepEqual(got, want) {
				t.Fatalf("quictcpSettings = %#v, want %#v", got, want)
			}
			buildCoreStream(t, stream)
		})
	}
}

func TestParseLink_Hysteria2HyTCP(t *testing.T) {
	res, err := ParseLink("hysteria2://secret@h.example.com:443?sni=h.example.com&type=hytcp&lanePassword=lanepw&lanePorts=4000#h")
	if err != nil {
		t.Fatalf("ParseLink: %v", err)
	}
	stream, _ := res.Outbound["streamSettings"].(map[string]any)
	if stream["network"] != "hytcp" {
		t.Fatalf("network = %v, want hytcp", stream["network"])
	}
	if _, ok := stream["hysteriaSettings"]; ok {
		t.Fatal("hytcp stream carries hysteriaSettings, which the core refuses beside hytcp")
	}
	want := map[string]any{"version": 2, "auth": "secret", "udpIdleTimeout": 60, "password": "lanepw", "ports": "4000"}
	if got := stream["hytcpSettings"]; !reflect.DeepEqual(got, want) {
		t.Fatalf("hytcpSettings = %#v, want %#v", got, want)
	}
	buildCoreStream(t, stream)
}

func TestParseLink_HyTCPStream(t *testing.T) {
	want := map[string]any{"version": 2, "auth": "hyauth", "udpIdleTimeout": 60, "password": "lanepw", "ports": "4000-4001"}
	vmessJSON := `{"v":"2","ps":"h","add":"h.example.com","port":443,"id":"11111111-2222-4333-8444-555555555555",` +
		`"net":"hytcp","tls":"tls","sni":"h.example.com","alpn":"h3","auth":"hyauth","lanePassword":"lanepw","lanePorts":"4000-4001"}`
	for name, link := range map[string]string{
		"vless": "vless://11111111-2222-4333-8444-555555555555@h.example.com:443?type=hytcp&security=tls" +
			"&sni=h.example.com&alpn=h3&auth=hyauth&lanePassword=lanepw&lanePorts=4000-4001#h",
		"vmess": "vmess://" + base64.StdEncoding.EncodeToString([]byte(vmessJSON)),
	} {
		t.Run(name, func(t *testing.T) {
			res, err := ParseLink(link)
			if err != nil {
				t.Fatalf("ParseLink: %v", err)
			}
			stream, _ := res.Outbound["streamSettings"].(map[string]any)
			if stream["network"] != "hytcp" || stream["security"] != "tls" {
				t.Fatalf("network/security = %v/%v, want hytcp/tls", stream["network"], stream["security"])
			}
			if got := stream["hytcpSettings"]; !reflect.DeepEqual(got, want) {
				t.Fatalf("hytcpSettings = %#v, want %#v", got, want)
			}
			buildCoreStream(t, stream)
		})
	}
}
