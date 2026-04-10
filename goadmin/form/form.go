package form

// FieldType is the rendered form input variant.
type FieldType string

const (
	FieldHidden   FieldType = "hidden"
	FieldDisplay  FieldType = "display"
	FieldText     FieldType = "text"
	FieldEmail    FieldType = "email"
	FieldPassword FieldType = "password"
	FieldTextarea FieldType = "textarea"
	FieldSelect   FieldType = "select"
	FieldMulti    FieldType = "multiselect"
)

// Option represents one select option.
type Option struct {
	Value string
	Label string
}

// Field defines a form input.
type Field struct {
	Name        string
	Label       string
	Type        FieldType
	Required    bool
	Readonly    bool
	Help        string
	Placeholder string
	Options     []Option
}

// Builder defines one resource form.
type Builder struct {
	Title         string
	Description   string
	Fields        []*Field
	HideDelete    bool
	SubmitLabel   string
	DeleteLabel   string
	CancelBackURL string
}

// New creates a form builder.
func New() *Builder {
	return &Builder{
		SubmitLabel: "Save",
		DeleteLabel: "Delete",
	}
}

// Field adds a field.
func (b *Builder) Field(name, label string, typ FieldType) *Field {
	field := &Field{Name: name, Label: label, Type: typ}
	b.Fields = append(b.Fields, field)
	return field
}

// Hidden adds a hidden field.
func (b *Builder) Hidden(name string) *Field {
	return b.Field(name, "", FieldHidden)
}

// Display adds a read-only field.
func (b *Builder) Display(name, label string) *Field {
	field := b.Field(name, label, FieldDisplay)
	field.Readonly = true
	return field
}

// Text adds a text field.
func (b *Builder) Text(name, label string) *Field {
	return b.Field(name, label, FieldText)
}

// Email adds an email field.
func (b *Builder) Email(name, label string) *Field {
	return b.Field(name, label, FieldEmail)
}

// Password adds a password field.
func (b *Builder) Password(name, label string) *Field {
	return b.Field(name, label, FieldPassword)
}

// Textarea adds a textarea field.
func (b *Builder) Textarea(name, label string) *Field {
	return b.Field(name, label, FieldTextarea)
}

// Select adds a select field.
func (b *Builder) Select(name, label string, options ...Option) *Field {
	field := b.Field(name, label, FieldSelect)
	field.Options = append(field.Options, options...)
	return field
}

// MultiSelect adds a multiple select field.
func (b *Builder) MultiSelect(name, label string, options ...Option) *Field {
	field := b.Field(name, label, FieldMulti)
	field.Options = append(field.Options, options...)
	return field
}

// MarkRequired makes validation intent visible in the UI.
func (f *Field) MarkRequired() *Field {
	f.Required = true
	return f
}

// WithHelp sets contextual help text.
func (f *Field) WithHelp(help string) *Field {
	f.Help = help
	return f
}

// WithPlaceholder sets the placeholder text.
func (f *Field) WithPlaceholder(placeholder string) *Field {
	f.Placeholder = placeholder
	return f
}
