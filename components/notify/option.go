package notify

// Option 定义了消息配置选项接口
type Option interface {
	apply(*Message)
}

// OptionFunc 是 Option 接口的函数实现
type OptionFunc func(*Message)

func (f OptionFunc) apply(msg *Message) {
	f(msg)
}

// WithSubject 设置消息主题
func WithSubject(subject string) Option {
	return OptionFunc(func(msg *Message) {
		msg.Subject = subject
	})
}

// WithContent 设置消息内容
func WithContent(content string) Option {
	return OptionFunc(func(msg *Message) {
		msg.Content = content
	})
}

// WithAttachments 设置消息附件
func WithAttachments(attachments []string) Option {
	return OptionFunc(func(msg *Message) {
		msg.Attachments = attachments
	})
}

// WithExtra 设置额外信息
func WithExtra(key string, value interface{}) Option {
	return OptionFunc(func(msg *Message) {
		if msg.Extra == nil {
			msg.Extra = make(map[string]interface{})
		}
		msg.Extra[key] = value
	})
}

// WithExtras 批量设置额外信息
func WithExtras(extras map[string]interface{}) Option {
	return OptionFunc(func(msg *Message) {
		if msg.Extra == nil {
			msg.Extra = make(map[string]interface{})
		}
		for k, v := range extras {
			msg.Extra[k] = v
		}
	})
}

// WithButtons sets the interactive buttons attached to the message. Drivers
// that support interactive cards render them as clickable elements.
func WithButtons(buttons []Button) Option {
	return OptionFunc(func(msg *Message) {
		msg.Buttons = buttons
	})
}

// WithRawSections sets pre-rendered kmarkdown sections carried verbatim on the
// message. A driver that supports raw sections renders them instead of the
// escaped Subject/Content block, giving the caller full control over the
// kmarkdown layout (e.g. a paragraph table with a bold header and individually
// escaped cells). The section content is the caller's responsibility to
// escape; the driver must NOT re-escape it. See RawSection for details.
func WithRawSections(sections []RawSection) Option {
	return OptionFunc(func(msg *Message) {
		msg.RawSections = sections
	})
}
