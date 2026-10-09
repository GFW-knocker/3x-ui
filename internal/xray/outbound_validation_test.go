package xray

import (
	"strings"
	"testing"
)

// TestValidateOutboundConfig_AllowsUnencryptedPublicVless pins Mahsa-Core's policy: it keeps
// unencrypted vless to public addresses that xray-core v26.7.11 started to refuse.
func TestValidateOutboundConfig_AllowsUnencryptedPublicVless(t *testing.T) {
	publicPlaintext := `{
		"protocol": "vless",
		"settings": {"address": "1.2.3.4", "port": 443, "id": "b831381d-6324-4d53-ad4f-8cda48b30811", "encryption": "none"},
		"streamSettings": {"network": "tcp", "security": "none"}
	}`
	if err := ValidateOutboundConfig([]byte(publicPlaintext)); err != nil {
		t.Fatalf("Mahsa-Core accepts a public unencrypted vless outbound, got: %v", err)
	}

	privatePlaintext := `{
		"protocol": "vless",
		"settings": {"address": "10.0.0.1", "port": 443, "id": "b831381d-6324-4d53-ad4f-8cda48b30811", "encryption": "none"},
		"streamSettings": {"network": "tcp", "security": "none"}
	}`
	if err := ValidateOutboundConfig([]byte(privatePlaintext)); err != nil {
		t.Fatalf("a private-address plaintext vless outbound must stay valid, got: %v", err)
	}

	publicTLS := `{
		"protocol": "vless",
		"settings": {"address": "1.2.3.4", "port": 443, "id": "b831381d-6324-4d53-ad4f-8cda48b30811", "encryption": "none"},
		"streamSettings": {"network": "tcp", "security": "tls", "tlsSettings": {"serverName": "example.com"}}
	}`
	if err := ValidateOutboundConfig([]byte(publicTLS)); err != nil {
		t.Fatalf("a TLS-secured public vless outbound must stay valid, got: %v", err)
	}
}

// The core feeds remoteDNS to netip.MustParseAddr when it creates the outbound, so a
// non-IP entry ("local" until 26.9.30) panics it at startup; conf.Build() lets it through.
func TestValidateOutboundConfig_RejectsWireguardRemoteDNSThatIsNotAnIP(t *testing.T) {
	outbound := func(remoteDNS string) []byte {
		return []byte(`{
			"protocol": "wireguard",
			"settings": {
				"secretKey": "yAnz5TF+lXXJte14tji3zlMNq+hd2rYUIgJBgB3fBmk=",
				"address": ["10.0.0.2/32"],
				"peers": [{"publicKey": "xTIBA5rboUvnH4htodjb6e697QjLERt1NAB4mZqp8Dg=", "endpoint": "162.159.192.1:2408"}],
				"remoteDNS": ` + remoteDNS + `
			}
		}`)
	}
	for _, rejected := range []string{`["local"]`, `["1.1.1.1", "dns.google"]`} {
		err := ValidateOutboundConfig(outbound(rejected))
		if err == nil || !strings.Contains(err.Error(), "remoteDNS") {
			t.Errorf("remoteDNS %s: want a remoteDNS refusal, got %v", rejected, err)
		}
	}
	if err := ValidateOutboundConfig(outbound(`["1.1.1.1", "2606:4700:4700::1111"]`)); err != nil {
		t.Fatalf("IP remoteDNS entries must stay valid, got: %v", err)
	}
}
