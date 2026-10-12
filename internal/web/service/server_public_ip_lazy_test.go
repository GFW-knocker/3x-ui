package service

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/GFW-knocker/3x-ui/v3/internal/database/dbtest"
)

// stubPublicIPServices points both lookup lists at one local server that counts
// hits and answers only once release is closed.
func stubPublicIPServices(t *testing.T) (hits *atomic.Int32, release chan struct{}) {
	t.Helper()
	hits = new(atomic.Int32)
	release = make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		select {
		case <-release:
		case <-r.Context().Done():
			return
		}
		_, _ = w.Write([]byte("203.0.113.7"))
	}))
	t.Cleanup(srv.Close)

	savedV4, savedV6 := publicIPv4Services, publicIPv6Services
	publicIPv4Services = []string{srv.URL}
	publicIPv6Services = []string{srv.URL}
	t.Cleanup(func() { publicIPv4Services, publicIPv6Services = savedV4, savedV6 })
	return hits, release
}

func initServerTestDB(t *testing.T) {
	t.Helper()
	dbDir := t.TempDir()
	t.Setenv("XUI_DB_FOLDER", dbDir)
	dbtest.InitDB(t, filepath.Join(dbDir, "x-ui.db"))
}

// The @2s ticker samples status whether or not anyone is logged in; a lookup
// started from there tells ipify/ipinfo about every panel at boot.
func TestStatusSampleNeverStartsPublicIPLookup(t *testing.T) {
	initServerTestDB(t)
	hits, release := stubPublicIPServices(t)
	close(release)

	svc := &ServerService{}
	for i := 0; i < 3; i++ {
		if status := svc.RefreshStatus(); status == nil || status.PublicIP.IPv4 != "" {
			t.Fatalf("sample %d: status = %+v, want a sample with no public IP", i, status)
		}
	}
	if status := svc.CurrentStatus(); status == nil || status.PublicIP.IPv4 != "" {
		t.Fatalf("CurrentStatus = %+v, want a sample with no public IP", status)
	}
	time.Sleep(300 * time.Millisecond)
	if n := hits.Load(); n != 0 {
		t.Fatalf("status sampling contacted the public IP services %d times, want 0", n)
	}
}

// A box with no IPv6 route spends 3s per lookup service; the dashboard request
// that triggers the lookup must not wait for it.
func TestDashboardLookupResolvesOffTheRequestPath(t *testing.T) {
	initServerTestDB(t)
	hits, release := stubPublicIPServices(t)

	svc := &ServerService{}
	svc.EnsurePublicIPsInBackground()
	status := svc.CurrentStatus()
	if status == nil || status.PublicIP.IPv4 != "" {
		t.Fatalf("CurrentStatus = %+v, want an immediate sample with the IP still unresolved", status)
	}

	close(release)
	deadline := time.Now().Add(10 * time.Second)
	for {
		if ipv4, _ := svc.publicIPs(); ipv4 == "203.0.113.7" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the background lookup never cached the public IP")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if got := svc.GetStatus(nil).PublicIP.IPv4; got != "203.0.113.7" {
		t.Fatalf("GetStatus after resolution: PublicIP.IPv4 = %q, want the cached address", got)
	}
	if hits.Load() == 0 {
		t.Fatal("the lookup services were never contacted")
	}
}

// With the lookup switched off nothing may reach ipify/ipinfo, and an address
// cached while it was on must disappear rather than linger as stale data.
func TestPublicIPLookupDisabledContactsNothing(t *testing.T) {
	initServerTestDB(t)
	hits, release := stubPublicIPServices(t)
	close(release)
	if err := (&SettingService{}).SetPublicIpLookupEnable(false); err != nil {
		t.Fatal(err)
	}

	svc := &ServerService{cachedIPv4: "198.51.100.9", cachedIPv6: "2001:db8::9"}
	if got := svc.GetStatusWithPublicIP(nil).PublicIP; got.IPv4 != "" || got.IPv6 != "" {
		t.Fatalf("GetStatusWithPublicIP kept a stale address %+v, want it cleared", got)
	}
	if host := svc.backupHost(""); host != "x-ui" {
		t.Fatalf("backupHost = %q, want the x-ui fallback instead of a looked-up address", host)
	}
	svc.EnsurePublicIPsInBackground()
	time.Sleep(300 * time.Millisecond)
	if n := hits.Load(); n != 0 {
		t.Fatalf("disabled lookup still contacted the services %d times", n)
	}
}
