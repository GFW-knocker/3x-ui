package service

import (
	"slices"
	"testing"
)

// TestSelectXrayVersions runs the real GFW-knocker/Xray-core release list (2026-10-09)
// through the filter: two tag schemes coexist, so a single numeric floor is wrong.
func TestSelectXrayVersions(t *testing.T) {
	tags := []string{
		"v1.26.10-mahsa-r2", "v1.26.10-mahsa-r1", "v1.26.9-mahsa-r6", "v1.26.9-mahsa-r5",
		"v1.26.9-mahsa-r4", "v1.26.9-mahsa-r3", "v1.26.9-mahsa-r2", "v1.26.9-mahsa-r1",
		"v26.8.30-mahsa-r1", "v1.26.5-mahsa-r1", "v1.26.3-mahsa-r1", "v1.26.2-mahsa-r1",
		"v1.25.8-mahsa-r1", "v1.25.7-mahsa-r1", "v1.25.4-mahsa-r2", "v1.25.4-mahsa-r1",
		"v1.25.2-mahsa-r1", "v1.25.1-mahsa-r1", "v24.12.18-mahsa-r1", "v1.8.23-mahsa-r3",
		"v1.8.23-mahsa-r2", "v1.8.23-mahsa-r1", "v1.8.21-mahsa-r2", "v1.8.21-mahsa-r1",
		"v1.8.13-mahsa-r2", "v1.8.13-mahsa-r1", "v1.8.10-r2-wireguard-fakehost",
		"v1.8.10-wgrd-fhost", "v1.8.7-wireguard-fakehost",
		// Module-only tags and plain XTLS releases are not Mahsa-Core builds.
		"v1.260327.0", "v26.9.30",
	}
	releases := make([]Release, len(tags))
	for i, tag := range tags {
		releases[i] = Release{TagName: tag}
	}

	got := selectXrayVersions(releases)
	want := []string{
		"v1.26.10-mahsa-r2", "v1.26.10-mahsa-r1", "v1.26.9-mahsa-r6", "v1.26.9-mahsa-r5",
		"v1.26.9-mahsa-r4", "v1.26.9-mahsa-r3", "v1.26.9-mahsa-r2", "v1.26.9-mahsa-r1",
		"v26.8.30-mahsa-r1",
	}
	if !slices.Equal(got, want) {
		t.Fatalf("selectXrayVersions\n got: %q\nwant: %q", got, want)
	}
}
