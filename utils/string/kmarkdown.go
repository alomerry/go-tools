package string

import "strings"

var kmarkdownEscapeReplacer = strings.NewReplacer(
	`\`, `\\`,
	`*`, `\*`,
	`_`, `\_`,
	`~`, `\~`,
	"`", "\\`",
	`<`, `\<`,
	`>`, `\>`,
	`[`, `\[`,
	`]`, `\]`,
	`(`, `\(`,
	`)`, `\)`,
	`@`, `\@`,
	`#`, `\#`,
	`!`, `\!`,
)

func EscapeKMarkdown(s string) string {
	if s == "" {
		return s
	}
	return kmarkdownEscapeReplacer.Replace(s)
}
