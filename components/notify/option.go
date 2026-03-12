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
