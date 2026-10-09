package service

import (
	"encoding/json"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

// An MVLESS inbound takes the same save and config path as VLESS: its stream survives
// normalization and every client reaches Mahsa-Core with its id and Vision flow.
func TestMvlessInboundReachesXrayConfig(t *testing.T) {
	setupConflictDB(t)
	const clientID = "5f2eb9d6-3a2f-4a55-9812-6ea1e2f7a333"
	_, _, err := (&InboundService{}).AddInbound(&model.Inbound{
		Tag: "in-mvless", Enable: true, Listen: "127.0.0.1", Port: 45401, Protocol: model.MVLESS,
		Settings: `{"clients":[{"id":"` + clientID + `","email":"mv1","flow":"xtls-rprx-vision","enable":true}],` +
			`"decryption":"none","fallbacks":[]}`,
		StreamSettings: `{"network":"tcp","security":"tls",` +
			`"tlsSettings":{"certificates":[{"certificateFile":"c.pem","keyFile":"k.pem"}]}}`,
	})
	if err != nil {
		t.Fatalf("AddInbound(mvless): %v", err)
	}

	cfg, err := (&XrayService{}).GetXrayConfig()
	if err != nil {
		t.Fatalf("GetXrayConfig: %v", err)
	}
	var emitted map[string]any
	for i := range cfg.InboundConfigs {
		if cfg.InboundConfigs[i].Tag == "in-mvless" {
			raw, _ := json.Marshal(cfg.InboundConfigs[i])
			_ = json.Unmarshal(raw, &emitted)
		}
	}
	if emitted == nil {
		t.Fatal("the mvless inbound is missing from the generated config")
	}
	if emitted["protocol"] != "mvless" {
		t.Fatalf("protocol = %v, want mvless", emitted["protocol"])
	}
	if stream, _ := emitted["streamSettings"].(map[string]any); stream["security"] != "tls" {
		t.Fatalf("streamSettings lost on save: %v", emitted["streamSettings"])
	}
	settings, _ := emitted["settings"].(map[string]any)
	clients, _ := settings["clients"].([]any)
	if len(clients) != 1 {
		t.Fatalf("clients = %v, want the one mvless client", settings["clients"])
	}
	client, _ := clients[0].(map[string]any)
	if client["id"] != clientID || client["flow"] != "xtls-rprx-vision" {
		t.Fatalf("client = %v, want id %s with the vision flow", client, clientID)
	}
	assertXrayAccepts(t, "the generated mvless inbound", buildGoldenInbound(t, emitted))
}
