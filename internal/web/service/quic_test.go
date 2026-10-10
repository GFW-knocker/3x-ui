package service

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// Mahsa-Core keeps xray's QUIC transport: an inbound saved with it must reach the core with
// its quicSettings intact, and the core must build it, with TLS or on its internal cert.
func TestQuicInboundReachesXrayConfig(t *testing.T) {
	quicSettings := map[string]any{
		"security": "chacha20-poly1305", "key": "s3cret", "header": map[string]any{"type": "wechat-video"},
	}
	rawQuic, _ := json.Marshal(quicSettings)
	streams := map[string]string{
		"tls":  `{"network":"quic","security":"tls","tlsSettings":{"certificates":[{"certificateFile":"c.pem","keyFile":"k.pem"}]},"quicSettings":` + string(rawQuic) + `}`,
		"none": `{"network":"quic","security":"none","quicSettings":` + string(rawQuic) + `}`,
	}
	for name, stream := range streams {
		t.Run(name, func(t *testing.T) {
			setupConflictDB(t)
			_, _, err := (&InboundService{}).AddInbound(&model.Inbound{
				Tag: "in-quic", Enable: true, Listen: "127.0.0.1", Port: 45402, Protocol: model.VLESS,
				Settings:       `{"clients":[{"id":"5f2eb9d6-3a2f-4a55-9812-6ea1e2f7a333","email":"q1","enable":true}],"decryption":"none"}`,
				StreamSettings: stream,
			})
			if err != nil {
				t.Fatalf("AddInbound(quic): %v", err)
			}
			cfg, err := (&XrayService{}).GetXrayConfig()
			if err != nil {
				t.Fatalf("GetXrayConfig: %v", err)
			}
			var emitted map[string]any
			for i := range cfg.InboundConfigs {
				if cfg.InboundConfigs[i].Tag == "in-quic" {
					raw, _ := json.Marshal(cfg.InboundConfigs[i])
					_ = json.Unmarshal(raw, &emitted)
				}
			}
			if emitted == nil {
				t.Fatal("the quic inbound is missing from the generated config")
			}
			streamSettings, _ := emitted["streamSettings"].(map[string]any)
			if streamSettings["network"] != "quic" || !reflect.DeepEqual(streamSettings["quicSettings"], quicSettings) {
				t.Fatalf("streamSettings = %v, want network quic with quicSettings %v", streamSettings, quicSettings)
			}
			assertXrayAccepts(t, "the generated quic inbound", buildGoldenInbound(t, emitted))
		})
	}
}
