package service

import (
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database"
	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// The core builds REALITY only over RAW, XHTTP and gRPC and refuses the whole config
// otherwise, so one inbound left on REALITY after a transport switch stopped every inbound.
func TestAddInbound_RealityTransport(t *testing.T) {
	tests := []struct {
		network string
		wantErr bool
	}{
		{"", false},
		{"tcp", false},
		{"raw", false},
		{"xhttp", false},
		{"splithttp", false},
		{"grpc", false},
		{"quic", true},
		{"ws", true},
		{"kcp", true},
		{"httpupgrade", true},
	}
	for i, tt := range tests {
		t.Run("network="+tt.network, func(t *testing.T) {
			setupConflictDB(t)
			tag := "in-reality-" + tt.network
			stream := `{"network":"` + tt.network + `","security":"reality","realitySettings":{}}`
			_, _, err := (&InboundService{}).AddInbound(&model.Inbound{
				Tag: tag, Enable: true, Listen: "0.0.0.0", Port: 44310 + i,
				Protocol: model.VLESS, StreamSettings: stream, Settings: `{"clients":[]}`,
			})
			if tt.wantErr != (err != nil) {
				t.Fatalf("AddInbound(reality over %q) error = %v, want error %v", tt.network, err, tt.wantErr)
			}
			var count int64
			if err := database.GetDB().Model(&model.Inbound{}).Where("tag = ?", tag).Count(&count).Error; err != nil {
				t.Fatalf("count rows: %v", err)
			}
			if want := map[bool]int64{true: 0, false: 1}[tt.wantErr]; count != want {
				t.Fatalf("rows for %s = %d, want %d", tag, count, want)
			}
		})
	}
}

func TestUpdateInbound_RejectsRealityOverQuic(t *testing.T) {
	setupConflictDB(t)
	seedInboundConflict(t, "in-44320-tcp", "0.0.0.0", 44320, model.MVLESS,
		`{"network":"tcp","security":"reality","realitySettings":{}}`, `{"clients":[]}`)

	var existing model.Inbound
	if err := database.GetDB().Where("tag = ?", "in-44320-tcp").First(&existing).Error; err != nil {
		t.Fatalf("read seeded row: %v", err)
	}
	update := existing
	update.StreamSettings = `{"network":"quic","security":"reality","realitySettings":{},"quicSettings":{}}`
	if _, _, err := (&InboundService{}).UpdateInbound(&update); err == nil {
		t.Fatal("UpdateInbound: want an error for REALITY over QUIC, got nil")
	}

	var reloaded model.Inbound
	if err := database.GetDB().First(&reloaded, existing.Id).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if reloaded.StreamSettings != existing.StreamSettings {
		t.Fatalf("rejected StreamSettings was persisted\ngot:  %s\nwant: %s", reloaded.StreamSettings, existing.StreamSettings)
	}
}
