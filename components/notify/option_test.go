package notify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestWithRawSections asserts the option populates Message.RawSections so a
// driver can render pre-escaped kmarkdown sections verbatim.
func TestWithRawSections(t *testing.T) {
	sections := []RawSection{
		{Cols: 3, Fields: []string{"**IP**", "**次数**"}},
		{Cols: 0},
	}
	msg := &Message{}
	WithRawSections(sections).apply(msg)

	assert.Equal(t, sections, msg.RawSections)
	assert.Len(t, msg.RawSections, 2)
	assert.Equal(t, 3, msg.RawSections[0].Cols)
}

// TestWithRawSections_Empty asserts an empty slice still sets the field (so a
// caller can explicitly opt into the raw path, though the kook driver treats
// an empty slice as "fall back to escaped content").
func TestWithRawSections_Empty(t *testing.T) {
	msg := &Message{}
	WithRawSections(nil).apply(msg)
	assert.Nil(t, msg.RawSections)
}

// TestMessage_OptionCompose asserts RawSections compose with the other options
// the kook driver relies on (subject, extras for group/level), mirroring how
// the ip-ban aggregator builds its message.
func TestMessage_OptionCompose(t *testing.T) {
	msg := &Message{}
	for _, opt := range []Option{
		WithSubject("闲时自动封禁汇总 - 共 2 个 IP"),
		WithRawSections([]RawSection{{Cols: 3, Fields: []string{"**IP**"}}}),
		WithExtra("group", "alarm"),
		WithExtra("level", "Error"),
	} {
		opt.apply(msg)
	}

	assert.Equal(t, "闲时自动封禁汇总 - 共 2 个 IP", msg.Subject)
	assert.Len(t, msg.RawSections, 1)
	assert.Equal(t, "alarm", msg.Extra["group"])
	assert.Equal(t, "Error", msg.Extra["level"])
}
