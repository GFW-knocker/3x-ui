package sub

import (
	"strings"
	"testing"

	"github.com/GFW-knocker/3x-ui/v3/internal/database"
	"github.com/GFW-knocker/3x-ui/v3/internal/database/model"
)

// Mahsa-Core's MVLESS shares VLESS's settings and differs only on the wire, so every
// renderer must emit exactly the VLESS output under the mvless name.

func asMvless(ib *model.Inbound) *model.Inbound {
	mv := *ib
	mv.Protocol = model.MVLESS
	return &mv
}

func TestGetLink_MvlessIsVlessUnderItsOwnScheme(t *testing.T) {
	externalProxies := `{
		"network":"tcp","security":"tls",
		"tcpSettings":{"header":{"type":"none"}},
		"tlsSettings":{"serverName":"base.sni","alpn":["h2"],"settings":{"fingerprint":"chrome"}},
		"externalProxy":[
			{"forceTls":"tls","dest":"cdn1.example.com","port":8443,"remark":"R1","sni":"sni1.example.com"},
			{"forceTls":"none","dest":"cdn2.example.com","port":80,"remark":"R2"}
		]
	}`
	s := &SubService{}
	for name, ib := range map[string]*model.Inbound{
		"vision over reality xhttp": flowTestInbound(xhttpRealityStream, testMlkemEncryption),
		"external proxies":          charVlessInbound(externalProxies),
	} {
		t.Run(name, func(t *testing.T) {
			vless := s.GetLink(ib, "user")
			if !strings.HasPrefix(vless, "vless://") {
				t.Fatalf("fixture produced no vless link: %q", vless)
			}
			want := strings.ReplaceAll(vless, "vless://", "mvless://")
			if got := s.GetLink(asMvless(ib), "user"); got != want {
				t.Fatalf("mvless link\n got: %q\nwant: %q", got, want)
			}
		})
	}
}

// renderSubscription serves one seeded inbound through the raw, JSON and Clash renderers.
func renderSubscription(t *testing.T, protocol model.Protocol) (links, jsonOut, clash string) {
	t.Helper()
	seedSubDB(t)
	ib := seedSubInbound(t, "s1", "mv", 4460, 1, wsTLSStream)
	if err := database.GetDB().Model(ib).Update("protocol", protocol).Error; err != nil {
		t.Fatalf("set protocol: %v", err)
	}
	raw, _, _, _, err := NewSubService("").GetSubs("s1", "req.example.com")
	if err != nil {
		t.Fatalf("GetSubs(%s): %v", protocol, err)
	}
	jsonOut, _, err = NewSubJsonService("", "", "", "", NewSubService("")).GetJson("s1", "req.example.com", false)
	if err != nil {
		t.Fatalf("GetJson(%s): %v", protocol, err)
	}
	clash, _, err = NewSubClashService(false, "", NewSubService("")).GetClash("s1", "req.example.com")
	if err != nil {
		t.Fatalf("GetClash(%s): %v", protocol, err)
	}
	return strings.Join(raw, "\n"), jsonOut, clash
}

func TestSubscription_MvlessInbound(t *testing.T) {
	vlessLinks, vlessJSON, vlessClash := renderSubscription(t, model.VLESS)
	mvLinks, mvJSON, mvClash := renderSubscription(t, model.MVLESS)

	if !strings.HasPrefix(vlessLinks, "vless://") {
		t.Fatalf("vless fixture produced no link: %q", vlessLinks)
	}
	if want := strings.ReplaceAll(vlessLinks, "vless://", "mvless://"); mvLinks != want {
		t.Fatalf("raw links\n got: %q\nwant: %q", mvLinks, want)
	}

	if !strings.Contains(vlessJSON, `"protocol": "vless"`) {
		t.Fatalf("vless fixture produced no vless outbound:\n%s", vlessJSON)
	}
	if want := strings.ReplaceAll(vlessJSON, `"protocol": "vless"`, `"protocol": "mvless"`); mvJSON != want {
		t.Fatalf("json outbounds\n got: %s\nwant: %s", mvJSON, want)
	}

	// mihomo has no MVLESS, so the node is left out instead of being mislabelled vless.
	if !strings.Contains(vlessClash, "type: vless") {
		t.Fatalf("vless fixture produced no clash proxy:\n%s", vlessClash)
	}
	if strings.Contains(mvClash, "type: vless") || strings.Contains(mvClash, "4460") {
		t.Fatalf("clash must not carry the mvless node:\n%s", mvClash)
	}
}
