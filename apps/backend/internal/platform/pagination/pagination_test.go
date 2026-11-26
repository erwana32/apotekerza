package pagination

import "testing"

func TestParse_LimitAndCursor(t *testing.T) {
	tests := []struct {
		name    string
		limit   string
		cursor  string
		want    int
		wantErr bool
	}{
		{"default", "", "", DefaultLimit, false},
		{"valid", "50", "abc", 50, false},
		{"upper boundary", "100", "", MaxLimit, false},
		{"zero", "0", "", 0, true},
		{"above max", "101", "", 0, true},
		{"negative", "-1", "", 0, true},
		{"not a number", "many", "", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params, err := Parse(tt.limit, tt.cursor)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if params.Limit != tt.want || params.Cursor != tt.cursor {
				t.Fatalf("Parse() = %+v, want limit %d cursor %q", params, tt.want, tt.cursor)
			}
		})
	}
}

func TestCursor_RoundTrip(t *testing.T) {
	encoded := EncodeCursor("product-id-123")
	if decoded, err := DecodeCursor(encoded); err != nil || decoded != "product-id-123" {
		t.Fatalf("DecodeCursor(%q) = %q, %v", encoded, decoded, err)
	}
	if _, err := DecodeCursor("not base64!"); err == nil {
		t.Fatal("DecodeCursor accepted invalid cursor")
	}
}
