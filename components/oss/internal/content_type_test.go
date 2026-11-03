package internal

import "testing"

func TestContentTypeByObjectKey(t *testing.T) {
	cases := []struct {
		key  string
		want string
	}{
		{"invoice/u1/a.jpg", "image/jpeg"},
		{"invoice/u1/a.JPEG", "image/jpeg"},
		{"dir/a.png", "image/png"},
		{"a.webp", "image/webp"},
		{"a.gif", "image/gif"},
		{"a.svg", "image/svg+xml"},
		{"a.pdf", "application/pdf"},
		{"a.json", "application/json"},
		{"a.txt", "text/plain; charset=utf-8"},
		{"a.mp4", "video/mp4"},
		{"a.mp3", "audio/mpeg"},
		{"a.apk", "application/vnd.android.package-archive"},
		{"a.woff2", "font/woff2"},
		{"a.unknownext", "application/octet-stream"},
		{"noext", "application/octet-stream"},
		{"a.", "application/octet-stream"},
		{"", "application/octet-stream"},
	}
	for _, c := range cases {
		if got := contentTypeByObjectKey(c.key); got != c.want {
			t.Fatalf("contentTypeByObjectKey(%q) = %q, want %q", c.key, got, c.want)
		}
	}
}
