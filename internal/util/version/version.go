// Package version compares 3x-ui release versions ("v3.8.0" or "3.8.0"), including this
// fork's "v3.9.0-mahsa-r1": an upstream version plus a revision, where plain is revision 0.
package version

import (
	"strconv"
	"strings"
)

// Compare returns -1, 0 or 1 as a is older than, equal to or newer than b, and
// false when either side is not MAJOR.MINOR.PATCH with an optional -mahsa-rN.
func Compare(a, b string) (int, bool) {
	aParts, okA := parse(a)
	bParts, okB := parse(b)
	if !okA || !okB {
		return 0, false
	}
	for i := range len(aParts) {
		if aParts[i] > bParts[i] {
			return 1, true
		}
		if aParts[i] < bParts[i] {
			return -1, true
		}
	}
	return 0, true
}

// Normalize strips surrounding space and a leading "v" from a version tag.
func Normalize(v string) string {
	return strings.TrimPrefix(strings.TrimSpace(v), "v")
}

const mahsaRevisionMarker = "-mahsa-r"

func parse(v string) ([4]int, bool) {
	var result [4]int
	base, revision, forked := strings.Cut(Normalize(v), mahsaRevisionMarker)
	if forked {
		n, ok := atoiStrict(revision)
		if !ok {
			return result, false
		}
		result[3] = n
	}
	parts := strings.Split(base, ".")
	if len(parts) != 3 {
		return result, false
	}
	for i, part := range parts {
		n, ok := atoiStrict(part)
		if !ok {
			return result, false
		}
		result[i] = n
	}
	return result, true
}

// atoiStrict accepts only plain decimal digits; strconv.Atoi would also take "+1".
func atoiStrict(s string) (int, bool) {
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}
