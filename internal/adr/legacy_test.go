package adr

import "testing"

func TestParseLegacyMADR(t *testing.T) {
	src := "# 12. Use Kafka for events\n\nDate: 2025-11-02\n\nStatus: accepted\n\n## Context\nVolume.\n"
	r, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if r.Number != 12 || r.Title != "Use Kafka for events" ||
		r.Status != Accepted || r.Date != "2025-11-02" {
		t.Fatalf("bad legacy record: %+v", r)
	}
}

func TestParseLegacyADRPrefix(t *testing.T) {
	src := "# ADR-0003: Use Redis for cache\n\n## Context\nLatency.\n"
	r, err := Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if r.Number != 3 || r.Title != "Use Redis for cache" {
		t.Fatalf("bad: %+v", r)
	}
	if r.Status != Proposed {
		t.Fatalf("default status should be proposed, got %q", r.Status)
	}
}

func TestParseLegacyMissingNumber(t *testing.T) {
	src := "# Some document without a number\n\ntext\n"
	_, err := Parse([]byte(src))
	if err == nil || !strings_contains(err.Error(), "missing numbered title") {
		t.Fatalf("want missing-number error, got %v", err)
	}
}

func strings_contains(h, n string) bool {
	return len(h) >= len(n) && (h == n || len(n) == 0 || indexOf(h, n) >= 0)
}

func indexOf(h, n string) int {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return i
		}
	}
	return -1
}
