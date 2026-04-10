package grid

import (
	"html/template"
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

// WithOptions sets select options.
func (f *Filter) WithOptions(options ...Option) *Filter {
	f.Options = append(f.Options, options...)
	return f
}
