package version

import "testing"

func TestCompareRejectsUnexpectedFormats(t *testing.T) {
	if _, ok := Compare("latest", "2.9.3"); ok {
		t.Fatal("expected non-semver latest tag to be rejected")
	}
	if _, ok := Compare("v2.9", "2.9.3"); ok {
		t.Fatal("expected short version to be rejected")
	}
}

// Fork releases are an upstream version plus a Mahsa revision (v3.9.0-mahsa-r1): the
// revision orders builds of one upstream version, and a plain upstream tag is revision 0.
func TestCompareMahsaRevisions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"v3.9.0-mahsa-r2", "3.9.0-mahsa-r1", 1},
		{"3.9.0-mahsa-r1", "v3.9.0-mahsa-r1", 0},
		{"3.9.0-mahsa-r1", "v3.9.0", 1},
		{"3.9.0-mahsa-r10", "3.9.0-mahsa-r9", 1},
		{"3.10.0", "3.9.0-mahsa-r5", 1},
		{"3.9.0-mahsa-r1", "v3.5.0", 1},
		{"v2.8.4-mahsa-r2", "3.9.0", -1},
	}
	for _, tc := range tests {
		t.Run(tc.a+" vs "+tc.b, func(t *testing.T) {
			got, ok := Compare(tc.a, tc.b)
			if !ok || got != tc.want {
				t.Fatalf("Compare(%q, %q) = %d, %v; want %d, true", tc.a, tc.b, got, ok, tc.want)
			}
		})
	}
}

func TestCompareRejectsOtherSuffixes(t *testing.T) {
	for _, v := range []string{"3.9.0-beta", "3.9.0-mahsa-r", "3.9.0-mahsa-rX", "3.9.0-mahsa-r1-extra", "3.9.0+mahsa"} {
		if _, ok := Compare(v, "3.9.0"); ok {
			t.Errorf("Compare(%q, ...) parsed an unknown version format", v)
		}
	}
}
