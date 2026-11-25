package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// The inbound forms write quictcpSettings/hytcpSettings as the quic or hysteria fields plus
// the server-side lane options; the core must build that shape or every inbound stops.
func TestTcpLaneInboundsBuildInXray(t *testing.T) {
	const tls = `"security":"tls","tlsSettings":{"serverName":"lane.example.test","alpn":["h3"],
		"certificates":[{"certificateFile":"/etc/ssl/certs/lane.crt","keyFile":"/etc/ssl/private/lane.key"}]}`
	cases := []struct {
		name     string
		protocol model.Protocol
		settings string
		stream   string
	}{
		{
			name:     "vless quictcp",
			protocol: model.VLESS,
			settings: `{"clients":[{"id":"b831381d-6324-4d53-ad4f-8cda48b30811","email":"lane"}],"decryption":"none"}`,
			stream: `{"network":"quictcp","security":"none","quictcpSettings":{"security":"none","key":"",
				"header":{"type":"none"},"password":"lanepw","ports":"3000-3002","maxLifeMs":30000,
				"sndbuf":0,"maxLanesPerIP":64},"finalmask":{"udp":[{"type":"salamander","settings":{"password":"saltpw"}}]}}`,
		},
		{
			name:     "hysteria hytcp",
			protocol: model.Hysteria,
			settings: `{"version":2,"clients":[{"auth":"secret","email":"lane"}]}`,
			stream: `{"network":"hytcp",` + tls + `,"hytcpSettings":{"version":2,"udpIdleTimeout":60,
				"password":"lanepw","ports":4000,"maxLifeMs":30000}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ib := &model.Inbound{
				Protocol: tc.protocol, Port: 8443, Listen: "127.0.0.1", Tag: "lane-in",
				Settings: tc.settings, StreamSettings: tc.stream,
			}
			raw, err := json.Marshal(ib.GenXrayInboundConfig())
			if err != nil {
				t.Fatalf("marshal generated inbound: %v", err)
			}
			var generated map[string]any
			if err := json.Unmarshal(raw, &generated); err != nil {
				t.Fatalf("decode generated inbound: %v", err)
			}
			assertXrayAccepts(t, "the "+tc.name+" inbound", buildGoldenInbound(t, generated))
		})
	}
}

// Without the hysteria protocol's user list, the hysteria listener needs its own auth and
// TLS, or the core fails to listen and every inbound goes down with it.
func TestAddInbound_HysteriaTransportOnOtherProtocols(t *testing.T) {
	setupConflictDB(t)
	const tls = `"security":"tls","tlsSettings":{"alpn":["h3"],
		"certificates":[{"certificateFile":"/etc/ssl/certs/lane.crt","keyFile":"/etc/ssl/private/lane.key"}]}`
	vless := `{"clients":[{"id":"b831381d-6324-4d53-ad4f-8cda48b30811","email":"hy-%d"}],"decryption":"none"}`
	cases := []struct {
		name     string
		protocol model.Protocol
		settings string
		stream   string
		wantErr  string
	}{
		{"vless hytcp without auth", model.VLESS, vless,
			`{"network":"hytcp",` + tls + `,"hytcpSettings":{"version":2,"auth":""}}`, "transport password"},
		{"vmess hysteria without auth", model.VMESS, `{"clients":[{"id":"b831381d-6324-4d53-ad4f-8cda48b30812","email":"hy-%d"}]}`,
			`{"network":"hysteria",` + tls + `,"hysteriaSettings":{"version":2}}`, "transport password"},
		{"vless hytcp without tls", model.VLESS, vless,
			`{"network":"hytcp","security":"none","hytcpSettings":{"version":2,"auth":"hyauth"}}`, "TLS"},
		{"vless hytcp with auth and tls", model.VLESS, vless,
			`{"network":"hytcp",` + tls + `,"hytcpSettings":{"version":2,"auth":"hyauth"}}`, ""},
		{"hysteria hytcp keeps per-client auth", model.Hysteria, `{"version":2,"clients":[{"auth":"secret","email":"hy-%d"}]}`,
			`{"network":"hytcp",` + tls + `,"hytcpSettings":{"version":2}}`, ""},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &model.Inbound{
				Tag: fmt.Sprintf("hy-in-%d", i), Enable: true, Listen: "0.0.0.0", Port: 46100 + i,
				Protocol: tc.protocol, Settings: fmt.Sprintf(tc.settings, i), StreamSettings: tc.stream,
			}
			_, _, err := (&InboundService{}).AddInbound(in)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("AddInbound refused a working inbound: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("AddInbound error = %v, want one naming %q", err, tc.wantErr)
			}
		})
	}
}

func TestUpdateInbound_RejectsHyTcpWithoutAuth(t *testing.T) {
	setupConflictDB(t)
	stream := `{"network":"hytcp","security":"tls","tlsSettings":{"alpn":["h3"],
		"certificates":[{"certificateFile":"/etc/ssl/certs/lane.crt","keyFile":"/etc/ssl/private/lane.key"}]},
		"hytcpSettings":{"version":2,"auth":"hyauth"}}`
	in := &model.Inbound{
		Tag: "hy-upd", Enable: true, Listen: "0.0.0.0", Port: 46200, Protocol: model.VLESS,
		Settings:       `{"clients":[{"id":"b831381d-6324-4d53-ad4f-8cda48b30813","email":"hy-upd"}],"decryption":"none"}`,
		StreamSettings: stream,
	}
	svc := &InboundService{}
	created, _, err := svc.AddInbound(in)
	if err != nil {
		t.Fatalf("AddInbound: %v", err)
	}
	created.StreamSettings = strings.Replace(stream, `"auth":"hyauth"`, `"auth":""`, 1)
	if _, _, err := svc.UpdateInbound(created); err == nil || !strings.Contains(err.Error(), "transport password") {
		t.Fatalf("UpdateInbound error = %v, want one naming the transport password", err)
	}
}
