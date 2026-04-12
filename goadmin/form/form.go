package form

// FieldType is the rendered form input variant.
type FieldType string

const (
	FieldHidden    FieldType = "hidden"
	FieldDisplay   FieldType = "display"
	FieldText      FieldType = "text"
	FieldEmail     FieldType = "email"
	FieldPassword  FieldType = "password"
	FieldTextarea  FieldType = "textarea"
	FieldSelect    FieldType = "select"
	FieldMulti     FieldType = "multiselect"
	FieldSwitch    FieldType = "switch"
	FieldTags      FieldType = "tags"
	FieldDate      FieldType = "date"
	FieldDatetime  FieldType = "datetime-local"
	FieldDateRange FieldType = "daterange"
	FieldUpload    FieldType = "file"
	FieldRepeater  FieldType = "repeater"
)

// Option represents one select option.
type Option struct {
	Value string
	Label string
}

// Field defines a form input.
type Field struct {
	Name              string
	SecondName        string
	Label             string
	Type              FieldType
	ValuePath         string
	SecondValuePath   string
	Multiple          bool
	MaxFileSize       int64
	AllowedExtensions []string
	Required          bool
	Readonly          bool
	Help              string
	Placeholder       string
	SecondPlaceholder string
	Options           []Option
	RepeaterFields    []*Field
	RepeaterMinRows   int
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

// Switch adds a boolean checkbox field.
func (b *Builder) Switch(name, label string) *Field {
	return b.Field(name, label, FieldSwitch)
}

// Tags adds a comma-separated tags field.
func (b *Builder) Tags(name, label string) *Field {
	return b.Field(name, label, FieldTags)
}

// Date adds a date field.
func (b *Builder) Date(name, label string) *Field {
	return b.Field(name, label, FieldDate)
}

// Datetime adds a datetime-local field.
func (b *Builder) Datetime(name, label string) *Field {
	return b.Field(name, label, FieldDatetime)
}

// DateRange adds a paired date range field.
func (b *Builder) DateRange(startName, endName, label string) *Field {
	field := b.Field(startName, label, FieldDateRange)
	field.SecondName = endName
	return field
}

// Upload adds a file upload field.
func (b *Builder) Upload(name, label string) *Field {
	return b.Field(name, label, FieldUpload)
}

// Repeater adds a has-many style nested editor backed by a JSON string field.
func (b *Builder) Repeater(name, label string) *RepeaterBuilder {
	field := b.Field(name, label, FieldRepeater)
	field.RepeaterMinRows = 3
	return &RepeaterBuilder{field: field}
}

// MultiUpload adds a multiple-file upload field.
func (b *Builder) MultiUpload(name, label string) *Field {
	field := b.Field(name, label, FieldUpload)
	field.Multiple = true
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

// ValueFrom overrides the record field used to populate the form value.
func (f *Field) ValueFrom(path string) *Field {
	f.ValuePath = path
	return f
}

// SecondValueFrom overrides the record field used to populate the second value.
func (f *Field) SecondValueFrom(path string) *Field {
	f.SecondValuePath = path
	return f
}

// WithSecondPlaceholder sets the placeholder for the second input in paired fields.
func (f *Field) WithSecondPlaceholder(placeholder string) *Field {
	f.SecondPlaceholder = placeholder
	return f
}

// AllowExtensions restricts upload file extensions.
func (f *Field) AllowExtensions(exts ...string) *Field {
	f.AllowedExtensions = append(f.AllowedExtensions, exts...)
	return f
}

// MaxSize limits upload file size in bytes.
func (f *Field) MaxSize(size int64) *Field {
	f.MaxFileSize = size
	return f
}

// RepeaterBuilder configures nested sub-fields.
type RepeaterBuilder struct {
	field *Field
}

func (b *RepeaterBuilder) add(name, label string, typ FieldType) *Field {
	child := &Field{Name: name, Label: label, Type: typ}
	b.field.RepeaterFields = append(b.field.RepeaterFields, child)
	return child
}

func (b *RepeaterBuilder) Hidden(name string) *Field {
	return b.add(name, "", FieldHidden)
}

func (b *RepeaterBuilder) Text(name, label string) *Field {
	return b.add(name, label, FieldText)
}

func (b *RepeaterBuilder) Textarea(name, label string) *Field {
	return b.add(name, label, FieldTextarea)
}

func (b *RepeaterBuilder) Date(name, label string) *Field {
	return b.add(name, label, FieldDate)
}

func (b *RepeaterBuilder) Datetime(name, label string) *Field {
	return b.add(name, label, FieldDatetime)
}

func (b *RepeaterBuilder) Select(name, label string, options ...Option) *Field {
	child := b.add(name, label, FieldSelect)
	child.Options = append(child.Options, options...)
	return child
}

func (b *RepeaterBuilder) MinRows(rows int) *RepeaterBuilder {
	if rows > 0 {
		b.field.RepeaterMinRows = rows
	}
	return b
}

func (b *RepeaterBuilder) ValueFrom(path string) *RepeaterBuilder {
	b.field.ValuePath = path
	return b
}

func (b *RepeaterBuilder) WithHelp(help string) *RepeaterBuilder {
	b.field.Help = help
	return b
}

func (b *RepeaterBuilder) MarkRequired() *RepeaterBuilder {
	b.field.Required = true
	return b
}
