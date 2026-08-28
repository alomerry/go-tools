package model

import "encoding/json"

type CardTheme string

const (
	CardThemePrimary   CardTheme = "primary"
	CardThemeSuccess   CardTheme = "success"
	CardThemeDanger    CardTheme = "danger"
	CardThemeWarning   CardTheme = "warning"
	CardThemeInfo      CardTheme = "info"
	CardThemeSecondary CardTheme = "secondary"
	CardThemeNone      CardTheme = "none"
)

type CardSize string

const (
	CardSizeLg CardSize = "lg"
	CardSizeSm CardSize = "sm"
)

type ModuleType string

const (
	ModuleTypeHeader      ModuleType = "header"
	ModuleTypeSection     ModuleType = "section"
	ModuleTypeImageGroup  ModuleType = "image-group"
	ModuleTypeContainer   ModuleType = "container"
	ModuleTypeActionGroup ModuleType = "action-group"
	ModuleTypeContext     ModuleType = "context"
	ModuleTypeDivider     ModuleType = "divider"
	ModuleTypeFile        ModuleType = "file"
	ModuleTypeAudio       ModuleType = "audio"
	ModuleTypeVideo       ModuleType = "video"
	ModuleTypeCountdown   ModuleType = "countdown"
	ModuleTypeInvite      ModuleType = "invite"
)

type ElementType string

const (
	ElementTypePlainText ElementType = "plain-text"
	ElementTypeKMarkdown ElementType = "kmarkdown"
	ElementTypeImage     ElementType = "image"
	ElementTypeButton    ElementType = "button"
	ElementTypeParagraph ElementType = "paragraph"
)

// CardMessage is an array of cards
type CardMessage []Card

func (m CardMessage) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

type Card struct {
	Type    string    `json:"type"`
	Theme   CardTheme `json:"theme,omitempty"`
	Size    CardSize  `json:"size,omitempty"`
	Modules []any     `json:"modules"`
}

type ModuleHeader struct {
	Type ModuleType  `json:"type"`
	Text ElementText `json:"text"`
}

type ModuleSection struct {
	Type      ModuleType `json:"type"`
	Text      any        `json:"text,omitempty"` // ElementText or ElementParagraph
	Mode      string     `json:"mode,omitempty"` // right, left, hidden
	Accessory any        `json:"accessory,omitempty"`
}

type ModuleDivider struct {
	Type ModuleType `json:"type"`
}

type ModuleActionGroup struct {
	Type     ModuleType `json:"type"`
	Elements []any      `json:"elements"`
}

type ModuleContext struct {
	Type     ModuleType `json:"type"`
	Elements []any      `json:"elements"`
}

type ElementText struct {
	Type    ElementType `json:"type"`
	Content string      `json:"content"`
}

type ElementImage struct {
	Type   ElementType `json:"type"`
	Src    string      `json:"src"`
	Alt    string      `json:"alt,omitempty"`
	Size   string      `json:"size,omitempty"` // sm, lg
	Circle bool        `json:"circle,omitempty"`
}

type ElementButton struct {
	Type  ElementType `json:"type"`
	Theme CardTheme   `json:"theme,omitempty"`
	Value string      `json:"value"`
	Click string      `json:"click,omitempty"` // link, return-val
	Text  ElementText `json:"text"`
}

// ElementParagraph is the Kook paragraph structure element. It lays out its
// Fields into a multi-column grid with at most Cols columns (Kook enforces a
// hard cap of 3). It is meant to be embedded as a section module's Text value
// (type="section", text={type:"paragraph", cols, fields}) to render a
// table-like layout, e.g. an aggregated IP ban summary.
//
// Unlike a plain ElementText, the Fields of a paragraph are kmarkdown text
// elements rendered verbatim (the caller is responsible for escaping each
// field's content), so a paragraph can carry bold headers and multi-field rows
// without the driver re-escaping the whole content.
type ElementParagraph struct {
	Type   ElementType   `json:"type"` // "paragraph"
	Cols   int           `json:"cols"`
	Fields []ElementText `json:"fields"`
}
