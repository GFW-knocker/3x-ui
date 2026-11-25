package sub

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// quictcp/hytcp links carry only the lane options both ends must agree on (password, ports);
// the rest tune one side and keep the client's own defaults.
const quicTCPStream = `{"network":"quictcp","security":"none",
	"quictcpSettings":{"security":"aes-128-gcm","key":"k1","header":{"type":"srtp"},
	"password":"lanepw","ports":"3000-3002","maxLifeMs":30000,"lanes":12}}`

func TestGenVlessLink_QuicTCPParams(t *testing.T) {
	link := (&SubService{}).genVlessLink(shareLinkInbound(quicTCPStream), "user")
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse %q: %v", link, err)
	}
	query := u.Query()
	for key, want := range map[string]string{
		"type": "quictcp", "quicSecurity": "aes-128-gcm", "key": "k1", "headerType": "srtp",
		"lanePassword": "lanepw", "lanePorts": "3000-3002",
	} {
		if got := query.Get(key); got != want {
			t.Errorf("%s = %q, want %q (link %s)", key, got, want, link)
		}
	}
	for _, sideOnly := range []string{"maxLifeMs", "lanes"} {
		if query.Has(sideOnly) {
			t.Errorf("one-side lane option %s leaked into the link %s", sideOnly, link)
		}
	}
}

func TestGenVmessLink_QuicTCPObject(t *testing.T) {
	in := &model.Inbound{
		Listen: "203.0.113.1", Port: 443, Protocol: model.VMESS, Remark: "q",
		Settings:       `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user","security":"auto"}]}`,
		StreamSettings: quicTCPStream,
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
		"net": "quictcp", "type": "srtp", "host": "aes-128-gcm", "path": "k1",
		"lanePassword": "lanepw", "lanePorts": "3000-3002",
	} {
		if got := obj[key]; got != want {
			t.Errorf("%s = %v, want %q (object %s)", key, got, want, raw)
		}
	}
}

// A standard hysteria2 client cannot dial TCP lanes, so the link names the transport.
func TestGenHysteriaLink_HyTCPParams(t *testing.T) {
	in := &model.Inbound{
		Id: 920101, Listen: "203.0.113.1", Port: 443, Protocol: model.Hysteria,
		Settings: `{"version":2,"clients":[{"auth":"secret","email":"user"}]}`,
		StreamSettings: `{"network":"hytcp","security":"tls",
			"hytcpSettings":{"version":2,"udpIdleTimeout":60,"password":"lanepw","ports":4000}}`,
	}
	link := (&SubService{}).genHysteriaLink(in, "user")
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse %q: %v", link, err)
	}
	query := u.Query()
	for key, want := range map[string]string{"type": "hytcp", "lanePassword": "lanepw", "lanePorts": "4000"} {
		if got := query.Get(key); got != want {
			t.Errorf("%s = %q, want %q (link %s)", key, got, want, link)
		}
	}
}

// The JSON subscription's hysteria outbound must dial the same lanes the inbound listens on.
func TestGenHy_HyTCPKeepsTransportAndLanes(t *testing.T) {
	in := &model.Inbound{
		Listen: "203.0.113.1", Port: 443, Protocol: model.Hysteria,
		Settings: `{"version":2,"clients":[{"auth":"secret","email":"user"}]}`,
		StreamSettings: `{"network":"hytcp","security":"tls",
			"hytcpSettings":{"version":2,"udpIdleTimeout":90,"password":"lanepw","ports":"4000-4001","maxLanesPerIP":8}}`,
	}
	var stream map[string]any
	if err := json.Unmarshal([]byte(in.StreamSettings), &stream); err != nil {
		t.Fatal(err)
	}
	raw := (&SubJsonService{}).genHy(in, stream, model.Client{Auth: "secret"}, "")
	var outbound struct {
		StreamSettings map[string]any `json:"streamSettings"`
	}
	if err := json.Unmarshal(raw, &outbound); err != nil {
		t.Fatalf("outbound is not json: %s", raw)
	}
	if got := outbound.StreamSettings["network"]; got != "hytcp" {
		t.Fatalf("network = %v, want hytcp (%s)", got, raw)
	}
	if _, ok := outbound.StreamSettings["hysteriaSettings"]; ok {
		t.Fatalf("hytcp outbound carries hysteriaSettings, which the core refuses beside hytcp (%s)", raw)
	}
	hy, _ := outbound.StreamSettings["hytcpSettings"].(map[string]any)
	want := map[string]any{"version": float64(2), "auth": "secret", "udpIdleTimeout": float64(90), "password": "lanepw", "ports": "4000-4001"}
	for key, value := range want {
		if hy[key] != value {
			t.Errorf("hytcpSettings.%s = %v, want %v (%s)", key, hy[key], value, raw)
		}
	}
}

// mihomo's hysteria2 dials UDP and cannot reach an inbound that listens on TCP lanes.
func TestClashSkipsHyTCPInbound(t *testing.T) {
	svc := &SubClashService{}
	inbound := &model.Inbound{
		Protocol: model.Hysteria, Listen: "192.0.2.1", Port: 443, Settings: `{"version":2}`,
		StreamSettings: `{"network":"hytcp","security":"tls","hytcpSettings":{"version":2}}`,
	}
	stream := map[string]any{"network": "hytcp", "security": "tls"}
	client := model.Client{Email: "client", Auth: "secret", Enable: true}
	if proxy := svc.buildProxy(&SubService{}, inbound, client, stream, nil); proxy != nil {
		t.Fatalf("Clash got a hysteria2 proxy for a hytcp inbound: %v", proxy)
	}
}

// vless/vmess over hytcp carry the hysteria transport's own password; without it in
// the link the client dials a server that refuses every connection.
const hyTCPStream = `{"network":"hytcp","security":"tls","tlsSettings":{"serverName":"h.example.test","alpn":["h3"]},
	"hytcpSettings":{"version":2,"auth":"hyauth","udpIdleTimeout":60,"password":"lanepw","ports":"4000-4001","maxLifeMs":30000}}`

func TestGenVlessLink_HyTCPParams(t *testing.T) {
	link := (&SubService{}).genVlessLink(shareLinkInbound(hyTCPStream), "user")
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse %q: %v", link, err)
	}
	query := u.Query()
	for key, want := range map[string]string{
		"type": "hytcp", "security": "tls", "alpn": "h3", "auth": "hyauth",
		"lanePassword": "lanepw", "lanePorts": "4000-4001",
	} {
		if got := query.Get(key); got != want {
			t.Errorf("%s = %q, want %q (link %s)", key, got, want, link)
		}
	}
	if query.Has("maxLifeMs") {
		t.Errorf("one-side lane option maxLifeMs leaked into the link %s", link)
	}
}

func TestGenVmessLink_HyTCPObject(t *testing.T) {
	in := &model.Inbound{
		Listen: "203.0.113.1", Port: 443, Protocol: model.VMESS, Remark: "h",
		Settings:       `{"clients":[{"id":"11111111-2222-4333-8444-555555555555","email":"user","security":"auto"}]}`,
		StreamSettings: hyTCPStream,
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
		"net": "hytcp", "tls": "tls", "alpn": "h3", "auth": "hyauth",
		"lanePassword": "lanepw", "lanePorts": "4000-4001",
	} {
		if got := obj[key]; got != want {
			t.Errorf("%s = %v, want %q (object %s)", key, got, want, raw)
		}
	}
}
