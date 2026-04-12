package grid

import (
	"html/template"
	"strings"
)

// Formatter customizes grid cell rendering.
type Formatter func(record any, value any) template.HTML

// FilterKind describes a supported filter input.
type FilterKind string

const (
	FilterText   FilterKind = "text"
	FilterSelect FilterKind = "select"
)

// Option is used by select-like inputs.
type Option struct {
	Value string
	Label string
}

// ActionStyle describes the visual intent of an action.
type ActionStyle string

const (
	ActionDefault ActionStyle = "default"
	ActionPrimary ActionStyle = "primary"
	ActionGhost   ActionStyle = "ghost"
	ActionDanger  ActionStyle = "danger"
)

// PageAction describes a toolbar action above the grid.
type PageAction struct {
	Label   string
	URL     string
	Style   ActionStyle
	Method  string
	Confirm string
}

// RowActionURL builds a row action target from the current record.
type RowActionURL func(record any) string

// RowAction describes an action rendered for each row.
type RowAction struct {
	Label   string
	URL     RowActionURL
	Style   ActionStyle
	Method  string
	Confirm string
}

// Filter describes a grid filter field.
type Filter struct {
	Name    string
	Label   string
	Kind    FilterKind
	Options []Option
}

// Column describes one grid column.
type Column struct {
	Name      string
	Label     string
	Sortable  bool
	Formatter Formatter
}

// SortableColumn marks the column as sortable.
func (c *Column) SortableColumn() *Column {
	c.Sortable = true
	return c
}

// Display sets a custom formatter.
func (c *Column) Display(fn Formatter) *Column {
	c.Formatter = fn
	return c
}

// Builder defines the grid page.
type Builder struct {
	Title         string
	Description   string
	QuickSearches []string
	Columns       []*Column
	Filters       []*Filter
	CreateLabel   string
	DisableCreate bool
	DisableView   bool
	DisableEdit   bool
	DisableDelete bool
	PageActions   []*PageAction
	RowActions    []*RowAction
}

// New creates a grid builder.
func New() *Builder {
	return &Builder{CreateLabel: "Create"}
}

// Column adds a generic column.
func (b *Builder) Column(name, label string) *Column {
	col := &Column{Name: name, Label: label}
	b.Columns = append(b.Columns, col)
	return col
}

// ID adds an ID column.
func (b *Builder) ID(label string) *Column {
	return b.Column("ID", label)
}

// QuickSearch enables a single keyword search over fields.
func (b *Builder) QuickSearch(fields ...string) {
	b.QuickSearches = append(b.QuickSearches, fields...)
}

// Filter adds a filter input.
func (b *Builder) Filter(name, label string, kind FilterKind) *Filter {
	filter := &Filter{Name: name, Label: label, Kind: kind}
	b.Filters = append(b.Filters, filter)
	return filter
}

// PageAction adds a top-level grid action.
func (b *Builder) PageAction(label, url string) *PageAction {
	action := &PageAction{
		Label:  label,
		URL:    url,
		Style:  ActionGhost,
		Method: "GET",
	}
	b.PageActions = append(b.PageActions, action)
	return action
}

// RowAction adds a per-row action.
func (b *Builder) RowAction(label string, url RowActionURL) *RowAction {
	action := &RowAction{
		Label:  label,
		URL:    url,
		Style:  ActionGhost,
		Method: "GET",
	}
	b.RowActions = append(b.RowActions, action)
	return action
}

// WithOptions sets select options.
func (f *Filter) WithOptions(options ...Option) *Filter {
	f.Options = append(f.Options, options...)
	return f
}

// WithStyle overrides the action style.
func (a *PageAction) WithStyle(style ActionStyle) *PageAction {
	a.Style = style
	return a
}

// WithMethod overrides the HTTP method.
func (a *PageAction) WithMethod(method string) *PageAction {
	if method != "" {
		a.Method = strings.ToUpper(method)
	}
	return a
}

// WithConfirm adds a confirmation prompt.
func (a *PageAction) WithConfirm(message string) *PageAction {
	a.Confirm = message
	return a
}

// WithStyle overrides the action style.
func (a *RowAction) WithStyle(style ActionStyle) *RowAction {
	a.Style = style
	return a
}

// WithMethod overrides the HTTP method.
func (a *RowAction) WithMethod(method string) *RowAction {
	if method != "" {
		a.Method = strings.ToUpper(method)
	}
	return a
}

// WithConfirm adds a confirmation prompt.
func (a *RowAction) WithConfirm(message string) *RowAction {
	a.Confirm = message
	return a
}
