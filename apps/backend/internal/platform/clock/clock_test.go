package clock

import (
	"testing"
	"time"
)

func TestSystem_NowIsUTC(t *testing.T) {
	before := time.Now().UTC()
	got := System().Now()
	after := time.Now().UTC()
	if got.Before(before) || got.After(after) {
		t.Fatalf("Now() = %s, want between %s and %s", got, before, after)
	}
	if got.Location() != time.UTC {
		t.Fatalf("location = %s, want UTC", got.Location())
	}
}

func TestFixed_ReturnsSameMoment(t *testing.T) {
	moment := time.Date(2026, 11, 26, 10, 30, 0, 0, time.UTC)
	inWIB := moment.In(time.FixedZone("WIB", 7*60*60))
	fixed := Fixed(inWIB)
	if got := fixed.Now(); !got.Equal(moment) {
		t.Fatalf("Now() = %s, want %s", got, moment)
	}
}
