package ids

import "testing"

func TestNew_Version7AndUnique(t *testing.T) {
	seen := make(map[string]bool, 1000)
	for range 1000 {
		id, err := New()
		if err != nil {
			t.Fatal(err)
		}
		if id.Version() != 7 {
			t.Fatalf("version = %d, want 7", id.Version())
		}
		if seen[id.String()] {
			t.Fatalf("duplicate id %s", id)
		}
		seen[id.String()] = true
	}
}

func TestParse_RoundTripAndInvalid(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(id.String())
	if err != nil || parsed != id {
		t.Fatalf("Parse(%s) = %s, %v", id, parsed, err)
	}
	if _, err := Parse("not-a-uuid"); err == nil {
		t.Fatal("Parse accepted invalid UUID")
	}
}
