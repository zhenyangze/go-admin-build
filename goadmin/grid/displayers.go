package grid

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

// Displayer is the interface for grid cell displayers
type Displayer interface {
	// Display returns the HTML representation of the value
	Display(record any, value any) template.HTML
}

// DisplayerFunc is a function type that implements Displayer
type DisplayerFunc func(record any, value any) template.HTML

// Display implements the Displayer interface
func (f DisplayerFunc) Display(record any, value any) template.HTML {
	return f(record, value)
}

// ToFormatter converts a Displayer to a Formatter
func ToFormatter(d Displayer) Formatter {
	return func(record any, value any) template.HTML {
		return d.Display(record, value)
	}
}

// Display sets a displayer for the column
func (c *Column) Displayer(d Displayer) *Column {
	c.Formatter = ToFormatter(d)
	return c
}

// ==================== Badge Displayer ====================

// Badge styles
const (
	BadgeStyleDefault BadgeStyle = "default"
	BadgeStylePrimary BadgeStyle = "primary"
	BadgeStyleSuccess BadgeStyle = "success"
	BadgeStyleWarning BadgeStyle = "warning"
	BadgeStyleDanger  BadgeStyle = "danger"
	BadgeStyleInfo    BadgeStyle = "info"
)

// BadgeStyle describes the visual style of a badge
type BadgeStyle string

// BadgeDisplayer displays values as badges
type BadgeDisplayer struct {
	style     BadgeStyle
	colorFunc func(record any, value any) BadgeStyle
}

// Badge creates a new badge displayer
func Badge() *BadgeDisplayer {
	return &BadgeDisplayer{
		style: BadgeStyleDefault,
	}
}

// Style sets the badge style
func (b *BadgeDisplayer) Style(style BadgeStyle) *BadgeDisplayer {
	b.style = style
	return b
}

// Color sets a dynamic color function
func (b *BadgeDisplayer) Color(fn func(record any, value any) BadgeStyle) *BadgeDisplayer {
	b.colorFunc = fn
	return b
}

// Display implements the Displayer interface
func (b *BadgeDisplayer) Display(record any, value any) template.HTML {
	style := b.style
	if b.colorFunc != nil {
		style = b.colorFunc(record, value)
	}

	class := fmt.Sprintf("inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800",
		style, style)

	return template.HTML(fmt.Sprintf(`<span class="%s">%v</span>`, class, value))
}

// ==================== Label Displayer ====================

// Label styles
const (
	LabelStyleDefault LabelStyle = "default"
	LabelStylePrimary LabelStyle = "primary"
	LabelStyleSuccess LabelStyle = "success"
	LabelStyleWarning LabelStyle = "warning"
	LabelStyleDanger  LabelStyle = "danger"
	LabelStyleInfo    LabelStyle = "info"
)

// LabelStyle describes the visual style of a label
type LabelStyle string

// LabelDisplayer displays values as labels/tags
type LabelDisplayer struct {
	style     LabelStyle
	colorFunc func(record any, value any) LabelStyle
}

// Label creates a new label displayer
func Label() *LabelDisplayer {
	return &LabelDisplayer{
		style: LabelStyleDefault,
	}
}

// Style sets the label style
func (l *LabelDisplayer) Style(style LabelStyle) *LabelDisplayer {
	l.style = style
	return l
}

// Color sets a dynamic color function
func (l *LabelDisplayer) Color(fn func(record any, value any) LabelStyle) *LabelDisplayer {
	l.colorFunc = fn
	return l
}

// Display implements the Displayer interface
func (l *LabelDisplayer) Display(record any, value any) template.HTML {
	style := l.style
	if l.colorFunc != nil {
		style = l.colorFunc(record, value)
	}

	class := fmt.Sprintf("inline-flex items-center px-2 py-0.5 rounded text-sm font-medium bg-%s-100 text-%s-800",
		style, style)

	return template.HTML(fmt.Sprintf(`<span class="%s">%v</span>`, class, value))
}

// ==================== Image Displayer ====================

// ImageDisplayer displays image thumbnails
type ImageDisplayer struct {
	width     string
	height    string
	preview   bool
	imageFunc func(record any, value any) string
}

// Image creates a new image displayer
func Image() *ImageDisplayer {
	return &ImageDisplayer{
		width:   "50",
		height:  "50",
		preview: true,
	}
}

// Size sets the image size
func (i *ImageDisplayer) Size(width, height int) *ImageDisplayer {
	i.width = fmt.Sprintf("%d", width)
	i.height = fmt.Sprintf("%d", height)
	return i
}

// Preview enables/disables image preview on click
func (i *ImageDisplayer) Preview(enabled bool) *ImageDisplayer {
	i.preview = enabled
	return i
}

// Src sets a custom image source function
func (i *ImageDisplayer) Src(fn func(record any, value any) string) *ImageDisplayer {
	i.imageFunc = fn
	return i
}

// Display implements the Displayer interface
func (i *ImageDisplayer) Display(record any, value any) template.HTML {
	src := fmt.Sprintf("%v", value)
	if i.imageFunc != nil {
		src = i.imageFunc(record, value)
	}

	if src == "" || src == "<nil>" {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	imgHTML := fmt.Sprintf(
		`<img src="%s" class="rounded object-cover" style="width:%spx;height:%spx;" alt="">`,
		src, i.width, i.height)

	if i.preview {
		return template.HTML(fmt.Sprintf(
			`<a href="%s" target="_blank" class="inline-block hover:opacity-80 transition-opacity">%s</a>`,
			src, imgHTML))
	}

	return template.HTML(imgHTML)
}

// ==================== Link Displayer ====================

// LinkDisplayer displays values as clickable links
type LinkDisplayer struct {
	urlFunc   func(record any, value any) string
	target    string
	maxLength int
}

// Link creates a new link displayer
func Link() *LinkDisplayer {
	return &LinkDisplayer{
		target:    "_self",
		maxLength: 0,
	}
}

// URL sets the link URL function
func (l *LinkDisplayer) URL(fn func(record any, value any) string) *LinkDisplayer {
	l.urlFunc = fn
	return l
}

// Target sets the link target
func (l *LinkDisplayer) Target(target string) *LinkDisplayer {
	l.target = target
	return l
}

// MaxLength sets the maximum display length (0 = no limit)
func (l *LinkDisplayer) MaxLength(length int) *LinkDisplayer {
	l.maxLength = length
	return l
}

// Display implements the Displayer interface
func (l *LinkDisplayer) Display(record any, value any) template.HTML {
	displayValue := fmt.Sprintf("%v", value)
	if displayValue == "<nil>" {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	// Truncate if needed
	originalValue := displayValue
	if l.maxLength > 0 && len(displayValue) > l.maxLength {
		displayValue = displayValue[:l.maxLength] + "..."
	}

	href := "#"
	if l.urlFunc != nil {
		href = l.urlFunc(record, value)
	} else {
		// Default: use the value as URL
		href = originalValue
	}

	return template.HTML(fmt.Sprintf(
		`<a href="%s" target="%s" class="text-blue-600 hover:text-blue-800 hover:underline">%s</a>`,
		href, l.target, displayValue))
}

// ==================== ProgressBar Displayer ====================

// ProgressBarDisplayer displays values as progress bars
type ProgressBarDisplayer struct {
	min       float64
	max       float64
	colorFunc func(record any, value any) string
	showText  bool
}

// ProgressBar creates a new progress bar displayer
func ProgressBar() *ProgressBarDisplayer {
	return &ProgressBarDisplayer{
		min:      0,
		max:      100,
		showText: true,
	}
}

// Range sets the min/max range
func (p *ProgressBarDisplayer) Range(min, max float64) *ProgressBarDisplayer {
	p.min = min
	p.max = max
	return p
}

// Color sets a dynamic color function
func (p *ProgressBarDisplayer) Color(fn func(record any, value any) string) *ProgressBarDisplayer {
	p.colorFunc = fn
	return p
}

// ShowText enables/disables text display
func (p *ProgressBarDisplayer) ShowText(show bool) *ProgressBarDisplayer {
	p.showText = show
	return p
}

// Display implements the Displayer interface
func (p *ProgressBarDisplayer) Display(record any, value any) template.HTML {
	var numValue float64
	switch v := value.(type) {
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	default:
		numValue = 0
	}

	// Calculate percentage
	percentage := 0.0
	if p.max > p.min {
		percentage = ((numValue - p.min) / (p.max - p.min)) * 100
	}
	if percentage < 0 {
		percentage = 0
	}
	if percentage > 100 {
		percentage = 100
	}

	// Determine color
	color := "blue"
	if p.colorFunc != nil {
		color = p.colorFunc(record, value)
	} else {
		// Auto color based on percentage
		if percentage < 30 {
			color = "red"
		} else if percentage < 70 {
			color = "yellow"
		} else {
			color = "green"
		}
	}

	text := ""
	if p.showText {
		text = fmt.Sprintf(`<span class="text-xs text-gray-600 ml-2">%.1f%%</span>`, percentage)
	}

	return template.HTML(fmt.Sprintf(
		`<div class="flex items-center w-full max-w-xs">`+
			`<div class="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">`+
			`<div class="h-full bg-%s-500 rounded-full" style="width:%.1f%%"></div>`+
			`</div>%s`+
			`</div>`,
		color, percentage, text))
}

// ==================== SwitchDisplay Displayer ====================

// SwitchDisplayer displays boolean values as switches
type SwitchDisplayer struct {
	onText  string
	offText string
	onColor string
	offColor string
}

// SwitchDisplay creates a new switch displayer
func SwitchDisplay() *SwitchDisplayer {
	return &SwitchDisplayer{
		onText:   "ON",
		offText:  "OFF",
		onColor:  "green",
		offColor: "gray",
	}
}

// Text sets the on/off text
func (s *SwitchDisplayer) Text(on, off string) *SwitchDisplayer {
	s.onText = on
	s.offText = off
	return s
}

// Color sets the on/off colors
func (s *SwitchDisplayer) Color(on, off string) *SwitchDisplayer {
	s.onColor = on
	s.offColor = off
	return s
}

// Display implements the Displayer interface
func (s *SwitchDisplayer) Display(record any, value any) template.HTML {
	isOn := false
	switch v := value.(type) {
	case bool:
		isOn = v
	case int, int64:
		isOn = v != 0
	case string:
		isOn = v == "1" || v == "true" || v == "yes" || v == "on"
	}

	if isOn {
		return template.HTML(fmt.Sprintf(
			`<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">`+
				`<span class="w-2 h-2 bg-%s-500 rounded-full mr-1.5"></span>%s`+
				`</span>`,
			s.onColor, s.onColor, s.onColor, s.onText))
	}

	return template.HTML(fmt.Sprintf(
		`<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">`+
			`<span class="w-2 h-2 bg-%s-500 rounded-full mr-1.5"></span>%s`+
			`</span>`,
		s.offColor, s.offColor, s.offColor, s.offText))
}

// ==================== QRCode Displayer ====================

// QRCodeDisplayer displays QR codes
type QRCodeDisplayer struct {
	size  int
	textFunc func(record any, value any) string
}

// QRCode creates a new QR code displayer
func QRCode() *QRCodeDisplayer {
	return &QRCodeDisplayer{
		size: 64,
	}
}

// Size sets the QR code size
func (q *QRCodeDisplayer) Size(size int) *QRCodeDisplayer {
	q.size = size
	return q
}

// Text sets a custom text function for the QR code
func (q *QRCodeDisplayer) Text(fn func(record any, value any) string) *QRCodeDisplayer {
	q.textFunc = fn
	return q
}

// Display implements the Displayer interface
func (q *QRCodeDisplayer) Display(record any, value any) template.HTML {
	text := fmt.Sprintf("%v", value)
	if q.textFunc != nil {
		text = q.textFunc(record, value)
	}

	// Use a simple QR code generation service or inline SVG
	// For now, use a data URI with a placeholder that can be replaced with actual QR generation
	escapedText := url.QueryEscape(text)
	imgURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=%dx%d&data=%s",
		q.size, q.size, escapedText)

	return template.HTML(fmt.Sprintf(
		`<img src="%s" width="%d" height="%d" alt="QR Code" class="rounded" title="%s">`,
		imgURL, q.size, q.size, template.HTMLEscapeString(text)))
}

// ==================== Copyable Displayer ====================

// CopyableDisplayer displays values with a copy button
type CopyableDisplayer struct {
	maxLength int
	showIcon  bool
}

// Copyable creates a new copyable displayer
func Copyable() *CopyableDisplayer {
	return &CopyableDisplayer{
		maxLength: 50,
		showIcon:  true,
	}
}

// MaxLength sets the maximum display length
func (c *CopyableDisplayer) MaxLength(length int) *CopyableDisplayer {
	c.maxLength = length
	return c
}

// ShowIcon enables/disables the copy icon
func (c *CopyableDisplayer) ShowIcon(show bool) *CopyableDisplayer {
	c.showIcon = show
	return c
}

// Display implements the Displayer interface
func (c *CopyableDisplayer) Display(record any, value any) template.HTML {
	text := fmt.Sprintf("%v", value)
	if text == "<nil>" {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	displayText := text
	if c.maxLength > 0 && len(text) > c.maxLength {
		displayText = text[:c.maxLength] + "..."
	}

	iconHTML := ""
	if c.showIcon {
		iconHTML = `<svg class="w-4 h-4 ml-1 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">` +
			`<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path>` +
			`</svg>`
	}

	return template.HTML(fmt.Sprintf(
		`<span class="group inline-flex items-center cursor-pointer hover:text-blue-600" `+
			`onclick="navigator.clipboard.writeText('%s');this.classList.add('text-green-600');setTimeout(()=>this.classList.remove('text-green-600'),500)">`+
			`%s%s`+
			`</span>`,
		template.JSEscapeString(text), displayText, iconHTML))
}

// ==================== Limit Displayer ====================

// LimitDisplayer truncates text with tooltip
type LimitDisplayer struct {
	limit     int
	tooltip   bool
	replace   string
}

// Limit creates a new limit displayer
func Limit(length int) *LimitDisplayer {
	return &LimitDisplayer{
		limit:   length,
		tooltip: true,
		replace: "...",
	}
}

// Tooltip enables/disables tooltip
func (l *LimitDisplayer) Tooltip(enabled bool) *LimitDisplayer {
	l.tooltip = enabled
	return l
}

// Replace sets the replacement string
func (l *LimitDisplayer) Replace(replacement string) *LimitDisplayer {
	l.replace = replacement
	return l
}

// Display implements the Displayer interface
func (l *LimitDisplayer) Display(record any, value any) template.HTML {
	text := fmt.Sprintf("%v", value)
	if text == "<nil>" {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	if len(text) <= l.limit {
		return template.HTML(template.HTMLEscapeString(text))
	}

	truncated := text[:l.limit] + l.replace

	if l.tooltip {
		return template.HTML(fmt.Sprintf(
			`<span title="%s" class="cursor-help border-b border-dotted border-gray-400">%s</span>`,
			template.HTMLEscapeString(text), template.HTMLEscapeString(truncated)))
	}

	return template.HTML(template.HTMLEscapeString(truncated))
}

// ==================== Table Displayer (for nested data) ====================

// TableDisplayer displays nested data as a small table
type TableDisplayer struct {
	columns   []string
	keyFunc   func(record any) string
	dataFunc  func(record any) []map[string]any
	maxRows   int
}

// Table creates a new table displayer
func Table() *TableDisplayer {
	return &TableDisplayer{
		maxRows: 5,
	}
}

// Columns sets the column keys to display
func (t *TableDisplayer) Columns(keys ...string) *TableDisplayer {
	t.columns = keys
	return t
}

// Data sets the data function
func (t *TableDisplayer) Data(fn func(record any) []map[string]any) *TableDisplayer {
	t.dataFunc = fn
	return t
}

// MaxRows sets the maximum number of rows to display
func (t *TableDisplayer) MaxRows(n int) *TableDisplayer {
	t.maxRows = n
	return t
}

// Display implements the Displayer interface
func (t *TableDisplayer) Display(record any, value any) template.HTML {
	var data []map[string]any

	// Try to parse value as JSON array
	if str, ok := value.(string); ok && str != "" {
		if err := json.Unmarshal([]byte(str), &data); err != nil {
			// Try as single object
			var single map[string]any
			if err := json.Unmarshal([]byte(str), &single); err == nil {
				data = []map[string]any{single}
			}
		}
	} else if arr, ok := value.([]map[string]any); ok {
		data = arr
	} else if t.dataFunc != nil {
		data = t.dataFunc(record)
	}

	if len(data) == 0 {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	// Determine columns if not set
	cols := t.columns
	if len(cols) == 0 && len(data) > 0 {
		for k := range data[0] {
			cols = append(cols, k)
		}
	}

	// Build table HTML
	var sb strings.Builder
	sb.WriteString(`<table class="min-w-full text-xs border border-gray-200">`)

	// Header
	sb.WriteString(`<thead class="bg-gray-50"><tr>`)
	for _, col := range cols {
		sb.WriteString(fmt.Sprintf(`<th class="px-2 py-1 text-left font-medium text-gray-500 border-b">%s</th>`, col))
	}
	sb.WriteString(`</tr></thead>`)

	// Body
	sb.WriteString(`<tbody>`)
	for i, row := range data {
		if i >= t.maxRows {
			sb.WriteString(fmt.Sprintf(`<tr><td colspan="%d" class="px-2 py-1 text-gray-500 italic">... %d more</td></tr>`,
				len(cols), len(data)-t.maxRows))
			break
		}
		sb.WriteString(`<tr class="border-b border-gray-100">`)
		for _, col := range cols {
			val := ""
			if v, ok := row[col]; ok {
				val = fmt.Sprintf("%v", v)
			}
			sb.WriteString(fmt.Sprintf(`<td class="px-2 py-1 text-gray-700">%s</td>`, template.HTMLEscapeString(val)))
		}
		sb.WriteString(`</tr>`)
	}
	sb.WriteString(`</tbody></table>`)

	return template.HTML(sb.String())
}

// ==================== Editable Displayer ====================

// EditableDisplayer allows inline editing
type EditableDisplayer struct {
	fieldName string
	displayType string // text, select, textarea
	options   []Option
	saveURL   func(record any) string
}

// Editable creates a new editable displayer
func Editable(fieldName string) *EditableDisplayer {
	return &EditableDisplayer{
		fieldName:   fieldName,
		displayType: "text",
	}
}

// Type sets the input type
func (e *EditableDisplayer) Type(t string) *EditableDisplayer {
	e.displayType = t
	return e
}

// Options sets options for select type
func (e *EditableDisplayer) Options(options ...Option) *EditableDisplayer {
	e.options = options
	return e
}

// SaveURL sets the save URL function
func (e *EditableDisplayer) SaveURL(fn func(record any) string) *EditableDisplayer {
	e.saveURL = fn
	return e
}

// Display implements the Displayer interface
func (e *EditableDisplayer) Display(record any, value any) template.HTML {
	displayValue := fmt.Sprintf("%v", value)
	if displayValue == "<nil>" {
		displayValue = ""
	}

	var inputHTML string
	switch e.displayType {
	case "select":
		var opts strings.Builder
		opts.WriteString(`<select class="editable-input border rounded px-2 py-1 text-sm w-full">`)
		for _, opt := range e.options {
			selected := ""
			if opt.Value == displayValue {
				selected = " selected"
			}
			opts.WriteString(fmt.Sprintf(`<option value="%s"%s>%s</option>`, opt.Value, selected, opt.Label))
		}
		opts.WriteString(`</select>`)
		inputHTML = opts.String()
	case "textarea":
		inputHTML = fmt.Sprintf(`<textarea class="editable-input border rounded px-2 py-1 text-sm w-full" rows="3">%s</textarea>`,
			template.HTMLEscapeString(displayValue))
	default:
		inputHTML = fmt.Sprintf(`<input type="text" class="editable-input border rounded px-2 py-1 text-sm w-full" value="%s">`,
			template.HTMLEscapeString(displayValue))
	}

	return template.HTML(fmt.Sprintf(
		`<div class="editable-cell group cursor-pointer hover:bg-gray-50 p-1 rounded" data-field="%s">`+
			`<span class="editable-display">%s</span>`+
			`<span class="editable-edit hidden">%s</span>`+
			`</div>`,
		e.fieldName, template.HTMLEscapeString(displayValue), inputHTML))
}

// ==================== Button Displayer ====================

// ButtonDisplayer displays a button
type ButtonDisplayer struct {
	label     string
	style     ActionStyle
	urlFunc   func(record any, value any) string
	onClick   func(record any, value any) string
}

// Button creates a new button displayer
func Button(label string) *ButtonDisplayer {
	return &ButtonDisplayer{
		label: label,
		style: ActionDefault,
	}
}

// Style sets the button style
func (b *ButtonDisplayer) Style(style ActionStyle) *ButtonDisplayer {
	b.style = style
	return b
}

// URL sets the button URL
func (b *ButtonDisplayer) URL(fn func(record any, value any) string) *ButtonDisplayer {
	b.urlFunc = fn
	return b
}

// OnClick sets the onclick handler
func (b *ButtonDisplayer) OnClick(fn func(record any, value any) string) *ButtonDisplayer {
	b.onClick = fn
	return b
}

// Display implements the Displayer interface
func (b *ButtonDisplayer) Display(record any, value any) template.HTML {
	styleClasses := map[ActionStyle]string{
		ActionDefault: "bg-white border border-gray-300 text-gray-700 hover:bg-gray-50",
		ActionPrimary: "bg-blue-600 text-white hover:bg-blue-700",
		ActionGhost:   "bg-transparent text-gray-600 hover:bg-gray-100",
		ActionDanger:  "bg-red-600 text-white hover:bg-red-700",
	}

	class := styleClasses[b.style]
	if class == "" {
		class = styleClasses[ActionDefault]
	}

	href := "#"
	if b.urlFunc != nil {
		href = b.urlFunc(record, value)
	}

	onclick := ""
	if b.onClick != nil {
		onclick = fmt.Sprintf(` onclick="%s"`, b.onClick(record, value))
	}

	label := b.label
	if label == "" {
		label = fmt.Sprintf("%v", value)
	}

	return template.HTML(fmt.Sprintf(
		`<a href="%s" class="inline-flex items-center px-3 py-1.5 text-sm font-medium rounded %s%s">%s</a>`,
		href, class, onclick, label))
}

// ==================== DropdownActions Displayer ====================

// DropdownActionsDisplayer displays a dropdown menu of actions
type DropdownActionsDisplayer struct {
	label   string
	actions []DropdownAction
}

// DropdownAction represents a single action in the dropdown
type DropdownAction struct {
	Label   string
	URL     func(record any, value any) string
	Style   ActionStyle
	Confirm string
}

// DropdownActions creates a new dropdown actions displayer
func DropdownActions(label string) *DropdownActionsDisplayer {
	return &DropdownActionsDisplayer{
		label:   label,
		actions: []DropdownAction{},
	}
}

// Action adds an action to the dropdown
func (d *DropdownActionsDisplayer) Action(label string, url func(record any, value any) string) *DropdownActionsDisplayer {
	d.actions = append(d.actions, DropdownAction{
		Label: label,
		URL:   url,
		Style: ActionDefault,
	})
	return d
}

// ActionWithConfirm adds an action with confirmation
func (d *DropdownActionsDisplayer) ActionWithConfirm(label string, url func(record any, value any) string, confirm string) *DropdownActionsDisplayer {
	d.actions = append(d.actions, DropdownAction{
		Label:   label,
		URL:     url,
		Style:   ActionDanger,
		Confirm: confirm,
	})
	return d
}

// Display implements the Displayer interface
func (d *DropdownActionsDisplayer) Display(record any, value any) template.HTML {
	if len(d.actions) == 0 {
		return template.HTML(`<span class="text-gray-400">-</span>`)
	}

	var items strings.Builder
	for _, action := range d.actions {
		url := "#"
		if action.URL != nil {
			url = action.URL(record, value)
		}

		confirm := ""
		if action.Confirm != "" {
			confirm = fmt.Sprintf(` onclick="return confirm('%s')"`, template.JSEscapeString(action.Confirm))
		}

		class := "block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
		if action.Style == ActionDanger {
			class = "block px-4 py-2 text-sm text-red-600 hover:bg-red-50"
		}

		items.WriteString(fmt.Sprintf(
			`<a href="%s" class="%s"%s>%s</a>`,
			url, class, confirm, action.Label))
	}

	return template.HTML(fmt.Sprintf(
		`<div class="relative inline-block text-left" x-data="{ open: false }">`+
			`<button @click="open = !open" class="inline-flex items-center px-3 py-1.5 text-sm font-medium bg-white border border-gray-300 rounded hover:bg-gray-50">`+
			`%s <svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"></path></svg>`+
			`</button>`+
			`<div x-show="open" @click.away="open = false" class="absolute right-0 z-10 mt-2 w-48 bg-white rounded-md shadow-lg ring-1 ring-black ring-opacity-5" style="display: none;">`+
			`<div class="py-1">%s</div>`+
			`</div>`+
			`</div>`,
		d.label, items.String()))
}
