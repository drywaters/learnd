package urlutil

import "testing"

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "strips tracking params and fragment",
			raw:  "https://Example.com/Path/?utm_source=newsletter&utm_medium=email#section",
			want: "https://example.com/Path",
		},
		{
			name: "preserves meaningful query params",
			raw:  "https://example.com/watch?v=abc&utm_source=foo",
			want: "https://example.com/watch?v=abc",
		},
		{
			name: "removes default http port and trims slash",
			raw:  "http://example.com:80/path/",
			want: "http://example.com/path",
		},
		{
			name: "removes default https port",
			raw:  "https://example.com:443/path",
			want: "https://example.com/path",
		},
		{
			name: "normalizes root path",
			raw:  "https://example.com",
			want: "https://example.com/",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := NormalizeURL(tt.raw)
			if err != nil {
				t.Fatalf("NormalizeURL returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("NormalizeURL = %q, want %q", got, tt.want)
			}
		})
	}
}
