package log

import "testing"

func TestNew_ValidAndInvalidLevel(t *testing.T) {
	if logger, err := New("info"); err != nil || logger == nil {
		t.Fatalf("New(info) = %v, %v", logger, err)
	}
	if logger, err := New("not-a-level"); err == nil || logger != nil {
		t.Fatalf("New(not-a-level) = %v, %v; want error", logger, err)
	}
}
