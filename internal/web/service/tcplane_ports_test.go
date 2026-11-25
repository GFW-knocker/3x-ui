package service

import (
	"fmt"
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/amneziawgnet"
	"github.com/GFW-knocker/3x-ui/v3/internal/database"
	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// laneStream is a quictcp/hytcp stream whose lane "ports" replace the inbound port:
// the core then binds TCP on every lane port and nothing on the inbound port.
func laneStream(network, ports string) string {
	return fmt.Sprintf(`{"network":%q,"security":"none","%sSettings":{"password":"lanepw","ports":%q}}`,
		network, network, ports)
}

func TestCheckPortConflict_TcpLanePorts(t *testing.T) {
	cases := []struct {
		name         string
		seedPort     int
		seedProtocol model.Protocol
		seedStream   string
		candidate    *model.Inbound
		wantPort     int // 0: no conflict
	}{
		{
			name: "lane range over a tcp inbound", seedPort: 3001, seedProtocol: model.VLESS,
			seedStream: `{"network":"tcp"}`,
			candidate: &model.Inbound{Listen: "0.0.0.0", Port: 443, Protocol: model.VLESS,
				StreamSettings: laneStream("quictcp", "3000-3002")},
			wantPort: 3001,
		},
		{
			name: "tcp inbound on a lane port", seedPort: 443, seedProtocol: model.VLESS,
			seedStream: laneStream("quictcp", "3000-3002"),
			candidate: &model.Inbound{Listen: "0.0.0.0", Port: 3002, Protocol: model.Trojan,
				StreamSettings: `{"network":"ws"}`},
			wantPort: 3002,
		},
		{
			name: "lane ports free the inbound port", seedPort: 443, seedProtocol: model.VLESS,
			seedStream: laneStream("hytcp", "3000-3002"),
			candidate: &model.Inbound{Listen: "0.0.0.0", Port: 443, Protocol: model.VLESS,
				StreamSettings: `{"network":"tcp"}`},
		},
		{
			name: "two lane ranges overlap", seedPort: 8443, seedProtocol: model.Hysteria,
			seedStream: laneStream("hytcp", "4000-4010"),
			candidate: &model.Inbound{Listen: "0.0.0.0", Port: 9443, Protocol: model.VLESS,
				StreamSettings: laneStream("quictcp", "4010,5000")},
			wantPort: 4010,
		},
		{
			name: "udp inbound shares a lane port", seedPort: 3001, seedProtocol: model.VLESS,
			seedStream: `{"network":"kcp"}`,
			candidate: &model.Inbound{Listen: "0.0.0.0", Port: 443, Protocol: model.VLESS,
				StreamSettings: laneStream("quictcp", "3000-3002")},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			setupConflictDB(t)
			seedInboundConflict(t, "seeded", "0.0.0.0", tc.seedPort, tc.seedProtocol, tc.seedStream, `{}`)
			got, err := (&InboundService{}).checkPortConflict(tc.candidate, 0)
			if err != nil {
				t.Fatalf("checkPortConflict: %v", err)
			}
			if tc.wantPort == 0 {
				if got != nil {
					t.Fatalf("want no conflict, got %s", got)
				}
				return
			}
			if got == nil || got.Port != tc.wantPort || got.Tag != "seeded" {
				t.Fatalf("want a conflict with 'seeded' on port %d, got %v", tc.wantPort, got)
			}
		})
	}
}

func TestCheckPortConflict_TcpLanePortsReachReservedPorts(t *testing.T) {
	setupConflictDB(t)
	svc := &InboundService{}
	api := &model.Inbound{Listen: "0.0.0.0", Port: 443, Protocol: model.VLESS,
		StreamSettings: laneStream("quictcp", fmt.Sprintf("%d-%d", defaultXrayAPIPort-2, defaultXrayAPIPort+2))}
	if got, err := svc.checkPortConflict(api, 0); err != nil || got == nil || got.Tag != "api" {
		t.Fatalf("a lane range over the Xray API port must conflict with it, got %v, %v", got, err)
	}

	seedInboundConflict(t, "awg-1", "0.0.0.0", 51820, model.AmneziaWG, ``, amneziawgRoutedSettings)
	var awg model.Inbound
	if err := database.GetDB().Where("tag = ?", "awg-1").First(&awg).Error; err != nil {
		t.Fatalf("read seeded row: %v", err)
	}
	relay := amneziawgnet.SOCKSPortForInbound(awg.Id)
	lanes := &model.Inbound{Listen: "0.0.0.0", Port: 443, Protocol: model.VLESS,
		StreamSettings: laneStream("quictcp", fmt.Sprintf("%d", relay))}
	if got, err := svc.checkPortConflict(lanes, 0); err != nil || got == nil || got.Tag != "awg-1" {
		t.Fatalf("a lane port on the AmneziaWG relay port %d must conflict, got %v, %v", relay, got, err)
	}

	seedInboundConflict(t, "lanes", "0.0.0.0", 443, model.VLESS, laneStream("quictcp", fmt.Sprintf("%d", relay)), `{}`)
	awg.Settings = amneziawgRoutedSettings
	if got, err := svc.checkPortConflict(&awg, awg.Id); err != nil || got == nil || got.Tag != "lanes" {
		t.Fatalf("the AmneziaWG relay port %d is another inbound's lane port; want a conflict, got %v, %v", relay, got, err)
	}
}

// The core opens one listener per lane port and refuses more than 1024, or an unparsable
// list, by failing the whole config; the panel must refuse it on save instead.
func TestAddInbound_ValidatesTcpLanePorts(t *testing.T) {
	setupConflictDB(t)
	cases := []struct {
		name, ports, wantErr string
	}{
		{"1025 ports", "20000-21024", "1024"},
		{"not a port list", "3000-abc", "lane ports"},
		{"1024 ports", "20000-21023", ""},
	}
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &model.Inbound{
				Tag: fmt.Sprintf("lanes-%d", i), Enable: true, Listen: "0.0.0.0", Port: 46300 + i, Protocol: model.VLESS,
				Settings:       fmt.Sprintf(`{"clients":[{"id":"b831381d-6324-4d53-ad4f-8cda48b3082%d","email":"lanes-%d"}],"decryption":"none"}`, i, i),
				StreamSettings: laneStream("quictcp", tc.ports),
			}
			_, _, err := (&InboundService{}).AddInbound(in)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("AddInbound refused %s: %v", tc.ports, err)
				}
				return
			}
			if err == nil || !strings.Contains(strings.ToLower(err.Error()), tc.wantErr) {
				t.Fatalf("AddInbound error = %v, want one naming %q", err, tc.wantErr)
			}
		})
	}
}

func TestBindConflicts_TcpLanePorts(t *testing.T) {
	cases := []struct {
		name     string
		inbounds string
		want     int
	}{
		{
			"lane port over another inbound's port",
			`{"listen":"0.0.0.0","port":443,"protocol":"vless","tag":"lanes","streamSettings":` + laneStream("quictcp", "3000-3002") + `},
			 {"listen":"0.0.0.0","port":3001,"protocol":"vmess","tag":"tcp"}`, 1,
		},
		{
			"lane ports free the inbound port",
			`{"listen":"0.0.0.0","port":443,"protocol":"vless","tag":"lanes","streamSettings":` + laneStream("hytcp", "3000-3002") + `},
			 {"listen":"0.0.0.0","port":443,"protocol":"vmess","tag":"tcp"}`, 0,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := bindConflicts(configFromInbounds(t, tc.inbounds), nil); len(got) != tc.want {
				t.Fatalf("bindConflicts = %v, want %d conflict(s)", got, tc.want)
			}
		})
	}
}
