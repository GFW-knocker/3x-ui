package maskcompat

import (
	"encoding/json"
	"testing"
)

func TestUpgradeLegacyXdns(t *testing.T) {
	tests := []struct {
		name string
		mask string
		want string
	}{
		{
			name: "bare server domain becomes TXT with the legacy EDNS0 size",
			mask: `{"type":"xdns","settings":{"domains":["t.example.com"]}}`,
			want: `{"domains":[{"edns0":1232,"name":"t.example.com","types":[16]}]}`,
		},
		{
			name: "method suffixes map to their record types",
			mask: `{"type":"xdns","settings":{"domains":["a.example.com:a","q.example.com:AAAA","t.example.com:txt"]}}`,
			want: `{"domains":[{"edns0":1232,"name":"a.example.com","types":[1]},{"edns0":1232,"name":"q.example.com","types":[28]},{"edns0":1232,"name":"t.example.com","types":[16]}]}`,
		},
		{
			name: "client resolver splits into its domain and a udp resolver",
			mask: `{"type":"XDNS","settings":{"resolvers":["t.example.com:a+udp://8.8.8.8:53"]}}`,
			want: `{"domains":[{"edns0":1232,"name":"t.example.com","types":[1]}],"resolvers":[{"settings":{"addr":"8.8.8.8:53"},"type":"udp"}]}`,
		},
		{
			name: "a resolver for an already listed domain adds no duplicate",
			mask: `{"type":"xdns","settings":{"domains":["t.example.com"],"resolvers":["T.example.com+udp://1.1.1.1:53"]}}`,
			want: `{"domains":[{"edns0":1232,"name":"t.example.com","types":[16]}],"resolvers":[{"settings":{"addr":"1.1.1.1:53"},"type":"udp"}]}`,
		},
		{
			name: "entries the old core refused are dropped",
			mask: `{"type":"xdns","settings":{"domains":["t.example.com","m.example.com:mx"],"resolvers":["1.1.1.1:53"]}}`,
			want: `{"domains":[{"edns0":1232,"name":"t.example.com","types":[16]}]}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var mask map[string]any
			if err := json.Unmarshal([]byte(tc.mask), &mask); err != nil {
				t.Fatalf("unmarshal mask: %v", err)
			}
			finalmask := map[string]any{"udp": []any{mask}}
			if !UpgradeLegacyXdns(finalmask) {
				t.Fatal("UpgradeLegacyXdns reported no change for a legacy mask")
			}
			got, err := json.Marshal(mask["settings"])
			if err != nil {
				t.Fatalf("marshal upgraded settings: %v", err)
			}
			if string(got) != tc.want {
				t.Fatalf("upgraded settings\n got: %s\nwant: %s", got, tc.want)
			}
		})
	}
}

func TestUpgradeLegacyXdnsLeavesCurrentShapeAlone(t *testing.T) {
	const current = `{"udp":[{"type":"xdns","settings":{"domains":[{"name":"t.example.com","types":[16,28],"edns0":1232}],"resolvers":[{"type":"tcp","settings":{"addr":"8.8.8.8:53"}}],"extraPoll":2}},{"type":"salamander","settings":{"password":"x"}}]}`
	var finalmask map[string]any
	if err := json.Unmarshal([]byte(current), &finalmask); err != nil {
		t.Fatalf("unmarshal finalmask: %v", err)
	}
	if UpgradeLegacyXdns(finalmask) {
		t.Fatal("UpgradeLegacyXdns rewrote a mask that is already in the object shape")
	}
	var want map[string]any
	_ = json.Unmarshal([]byte(current), &want)
	got, _ := json.Marshal(finalmask)
	wantJSON, _ := json.Marshal(want)
	if string(got) != string(wantJSON) {
		t.Fatalf("finalmask changed\n got: %s\nwant: %s", got, wantJSON)
	}
}
