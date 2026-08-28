package model

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kmarkdownField is a tiny helper building a kmarkdown ElementText, mirroring
// what a caller passes into AddParagraph for a table cell.
func kmarkdownField(content string) ElementText {
	return ElementText{Type: ElementTypeKMarkdown, Content: content}
}

// TestAddParagraph_EmbeddedAsSectionText asserts AddParagraph produces a
// section module whose text is a paragraph structure element (type=section,
// text={type:paragraph, cols, fields}) — the shape Kook expects for a
// multi-column table.
func TestAddParagraph_EmbeddedAsSectionText(t *testing.T) {
	fields := []ElementText{
		kmarkdownField("**IP**"), kmarkdownField("**归属地/运营商**"), kmarkdownField("**次数**"),
		kmarkdownField("1.2.3.4"), kmarkdownField("中国 阿里云"), kmarkdownField("3"),
	}
	card := NewCardBuilder().Theme(CardThemeDanger).AddParagraph(3, fields).Build()

	require.Len(t, card.Modules, 1)
	sec, ok := card.Modules[0].(ModuleSection)
	require.True(t, ok, "paragraph must be wrapped in a section module")
	assert.Equal(t, ModuleTypeSection, sec.Type)

	para, ok := sec.Text.(ElementParagraph)
	require.True(t, ok, "section text must be an ElementParagraph")
	assert.Equal(t, ElementTypeParagraph, para.Type)
	assert.Equal(t, 3, para.Cols)
	require.Len(t, para.Fields, 6)
	assert.Equal(t, "**IP**", para.Fields[0].Content)
	assert.Equal(t, "3", para.Fields[5].Content)
}

// TestAddParagraph_ColsClamp asserts cols is clamped to Kook's [1, 3] range.
func TestAddParagraph_ColsClamp(t *testing.T) {
	cases := []struct{ in, want int }{
		{in: 0, want: 1},
		{in: -2, want: 1},
		{in: 1, want: 1},
		{in: 2, want: 2},
		{in: 3, want: 3},
		{in: 4, want: 3},
		{in: 9, want: 3},
	}
	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			card := NewCardBuilder().AddParagraph(tt.in, []ElementText{kmarkdownField("x")}).Build()
			sec := card.Modules[0].(ModuleSection)
			para := sec.Text.(ElementParagraph)
			assert.Equal(t, tt.want, para.Cols)
		})
	}
}

// TestAddParagraph_JSONShape asserts the marshalled JSON carries the paragraph
// under text with type=paragraph and the fields array, so a real Kook payload
// round-trips correctly.
func TestAddParagraph_JSONShape(t *testing.T) {
	fields := []ElementText{
		kmarkdownField("**IP**"), kmarkdownField("**次数**"),
		kmarkdownField("1.2.3.4"), kmarkdownField("2"),
	}
	card := NewCardBuilder().Theme(CardThemeDanger).AddParagraph(2, fields).Build()
	out := CardMessage([]Card{card}).String()

	assert.Contains(t, out, `"type":"paragraph"`)
	assert.Contains(t, out, `"cols":2`)
	// fields carry the kmarkdown type and the verbatim content (no escaping by
	// the builder — the caller's responsibility).
	assert.Contains(t, out, `"**IP**"`)
	assert.Contains(t, out, `"1.2.3.4"`)

	// sanity: the section wrapper type is present
	var cm CardMessage
	require.NoError(t, json.Unmarshal([]byte(out), &cm))
	require.Len(t, cm, 1)
	require.Len(t, cm[0].Modules, 1)
	raw, err := json.Marshal(cm[0].Modules[0])
	require.NoError(t, err)
	body := string(raw)
	assert.True(t, strings.Contains(body, `"type":"section"`))
	assert.True(t, strings.Contains(body, `"type":"paragraph"`))
}

// TestAddParagraph_RendersAlongsideHeader asserts a paragraph section coexists
// with other modules (header/divider) in build order, matching how the kook
// driver lays out a title header + paragraph table.
func TestAddParagraph_RendersAlongsideHeader(t *testing.T) {
	card := NewCardBuilder().
		Theme(CardThemeDanger).
		AddHeader("闲时自动封禁汇总 - 共 2 个 IP").
		AddParagraph(3, []ElementText{
			kmarkdownField("**IP**"), kmarkdownField("**归属地/运营商**"), kmarkdownField("**次数**"),
			kmarkdownField("1.2.3.4"), kmarkdownField("中国 阿里云"), kmarkdownField("3"),
		}).
		Build()

	require.Len(t, card.Modules, 2)
	hdr, ok := card.Modules[0].(ModuleHeader)
	require.True(t, ok)
	assert.Equal(t, "闲时自动封禁汇总 - 共 2 个 IP", hdr.Text.Content)
	sec, ok := card.Modules[1].(ModuleSection)
	require.True(t, ok)
	_, ok = sec.Text.(ElementParagraph)
	assert.True(t, ok)
}
