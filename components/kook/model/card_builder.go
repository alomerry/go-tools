package model

// CardMessageBuilder builds a card message
type CardMessageBuilder struct {
	cards []Card
}

func NewCardMessageBuilder() *CardMessageBuilder {
	return &CardMessageBuilder{}
}

func (b *CardMessageBuilder) AddCard(card Card) *CardMessageBuilder {
	b.cards = append(b.cards, card)
	return b
}

func (b *CardMessageBuilder) Build() string {
	return CardMessage(b.cards).String()
}

// CardBuilder builds a single card
type CardBuilder struct {
	theme   CardTheme
	size    CardSize
	modules []any
}

func NewCardBuilder() *CardBuilder {
	return &CardBuilder{}
}

func (b *CardBuilder) Theme(t CardTheme) *CardBuilder {
	b.theme = t
	return b
}

func (b *CardBuilder) Size(s CardSize) *CardBuilder {
	b.size = s
	return b
}

func (b *CardBuilder) AddHeader(text string) *CardBuilder {
	b.modules = append(b.modules, ModuleHeader{
		Type: ModuleTypeHeader,
		Text: ElementText{
			Type:    ElementTypePlainText,
			Content: text,
		},
	})
	return b
}

func (b *CardBuilder) AddSectionKmarkdown(text string) *CardBuilder {
	b.modules = append(b.modules, ModuleSection{
		Type: ModuleTypeSection,
		Text: ElementText{
			Type:    ElementTypeKMarkdown,
			Content: text,
		},
	})
	return b
}

func (b *CardBuilder) AddSectionText(text string) *CardBuilder {
	b.modules = append(b.modules, ModuleSection{
		Type: ModuleTypeSection,
		Text: ElementText{
			Type:    ElementTypePlainText,
			Content: text,
		},
	})
	return b
}

func (b *CardBuilder) AddDivider() *CardBuilder {
	b.modules = append(b.modules, ModuleDivider{
		Type: ModuleTypeDivider,
	})
	return b
}

func (b *CardBuilder) AddContextKmarkdown(text string) *CardBuilder {
	b.modules = append(b.modules, ModuleContext{
		Type: ModuleTypeContext,
		Elements: []any{
			ElementText{
				Type:    ElementTypeKMarkdown,
				Content: text,
			},
		},
	})
	return b
}

// AddActionGroup appends an action-group module holding the given buttons.
// Each button uses Click "return-val" so the configured Value is echoed back
// to the webhook on click. Used by the notify kook driver to render the
// interactive buttons carried on a notify.Message.
func (b *CardBuilder) AddActionGroup(buttons []ElementButton) *CardBuilder {
	if len(buttons) == 0 {
		return b
	}
	elements := make([]any, 0, len(buttons))
	for i := range buttons {
		elements = append(elements, buttons[i])
	}
	b.modules = append(b.modules, ModuleActionGroup{
		Type:     ModuleTypeActionGroup,
		Elements: elements,
	})
	return b
}

func (b *CardBuilder) Build() Card {
	return Card{
		Type:    "card",
		Theme:   b.theme,
		Size:    b.size,
		Modules: b.modules,
	}
}
