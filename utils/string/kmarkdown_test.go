package string

import "testing"

func TestEscapeKMarkdown(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"plain", "1.2.3.4", "1.2.3.4"},
		{"bold injection", "hello *world*", `hello \*world\*`},
		{"italic injection", "a_b_c", `a\_b\_c`},
		{"strike injection", "a~b", `a\~b`},
		{"link injection", "[click](http://x)", `\[click\]\(http://x\)`},
		{"mention injection", "@everyone", `\@everyone`},
		{"hashtag injection", "#tag", `\#tag`},
		{"backslash first", `a\b`, `a\\b`},
		{"backtick", "a`b", "a\\`b"},
		{"angle", "<code>", `\<code\>`},
		{"bang", "a!b", `a\!b`},
		{"chinese", "中国 北京", "中国 北京"},
		{"mixed", "IP: 1.2.3.4 (内网)", `IP: 1.2.3.4 \(内网\)`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeKMarkdown(tt.in); got != tt.want {
				t.Errorf("EscapeKMarkdown(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
