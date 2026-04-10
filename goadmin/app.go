package goadmin

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/zhenyangze/go-admin-build/goadmin/form"
	"github.com/zhenyangze/go-admin-build/goadmin/grid"
	"github.com/zhenyangze/go-admin-build/goadmin/show"
	"github.com/zhenyangze/go-admin-build/goadmin/tree"
)

//go:embed assets/templates/*.tmpl assets/styles/admin.css
var assetFS embed.FS

type dashboardFunc func(context.Context) (DashboardData, error)

// App is the reusable admin HTTP handler.
type App struct {
	cfg         Config
	auth        AuthService
	templates   *template.Template
	dashboard   dashboardFunc
	resourceMap map[string]Resource
	resources   []Resource
	css         []byte
}

// New creates a new admin application.
func New(cfg Config, authService AuthService) (*App, error) {
	cfg = cfg.WithDefaults()
	if cfg.SessionSecret == "" {
		return nil, errors.New("session secret is required")
	}

	css, err := assetFS.ReadFile("assets/styles/admin.css")
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("admin").Funcs(template.FuncMap{
		"safeHTML": func(v any) template.HTML {
			switch typed := v.(type) {
			case template.HTML:
				return typed
			default:
				return template.HTML(template.HTMLEscapeString(fmt.Sprint(v)))
			}
		},
	}).ParseFS(assetFS, "assets/templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:         cfg,
		auth:        authService,
		templates:   tmpl,
		resourceMap: map[string]Resource{},
		css:         css,
	}, nil
}

// Register adds a CRUD resource to the app.
func (a *App) Register(resource Resource) {
	resource.Path = strings.Trim(resource.Path, "/")
	a.resourceMap[resource.Path] = resource
	a.resources = append(a.resources, resource)
}

// SetDashboard configures a custom dashboard callback.
func (a *App) SetDashboard(fn func(context.Context) (DashboardData, error)) {
	a.dashboard = fn
}

// Handler returns the reusable HTTP handler.
func (a *App) Handler() http.Handler {
	return a
}

// ServeHTTP implements http.Handler.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	prefix := normalizePath(a.cfg.Prefix)
	if !strings.HasPrefix(normalizePath(r.URL.Path), prefix) && normalizePath(r.URL.Path) != prefix {
		http.NotFound(w, r)
		return
	}

	if normalizePath(r.URL.Path) == joinURL(prefix, "assets", "admin.css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(a.css)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, prefix)
	path = strings.Trim(path, "/")

	switch path {
	case "login":
		a.handleLogin(w, r)
		return
	case "logout":
		a.handleLogout(w, r)
		return
	}

	state, identity := a.requireIdentity(w, r)
	if identity == nil {
		return
	}

	if path == "" {
		a.handleDashboard(w, r, state, identity)
		return
	}

	parts := strings.Split(path, "/")
	resource, ok := a.resourceMap[parts[0]]
	if !ok {
		http.NotFound(w, r)
		return
	}
	allowed, err := a.auth.Authorize(r.Context(), identity, resource.Permission)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
		return
	}
	if !allowed {
		a.renderShell(w, r, state, identity, "dashboard", pageData{
			PageTitle:       "Forbidden",
			PageDescription: "You do not have permission to access this module.",
			Message:         "Forbidden",
			Content: dashboardView{
				Cards: []DashboardCard{{Title: "Access denied", Value: "403", Hint: "Permission check failed"}},
			},
		})
		return
	}

	method := currentMethod(r)
	switch len(parts) {
	case 1:
		if method == http.MethodGet {
			a.handleGrid(w, r, state, identity, resource)
			return
		}
		if method == http.MethodPost {
			a.handleCreate(w, r, state, identity, resource)
			return
		}
	case 2:
		if parts[1] == "new" && method == http.MethodGet {
			a.handleNewForm(w, r, state, identity, resource)
			return
		}
		if parts[1] == "tree" && method == http.MethodGet {
			a.handleTree(w, r, state, identity, resource)
			return
		}
		if method == http.MethodGet {
			a.handleShow(w, r, state, identity, resource, parts[1])
			return
		}
		if method == http.MethodPut {
			a.handleUpdate(w, r, state, identity, resource, parts[1])
			return
		}
	case 3:
		if parts[2] == "edit" && method == http.MethodGet {
			a.handleEditForm(w, r, state, identity, resource, parts[1])
			return
		}
		if parts[2] == "delete" && method == http.MethodPost {
			a.handleDelete(w, r, state, identity, resource, parts[1])
			return
		}
	}

	http.NotFound(w, r)
}

func (a *App) requireIdentity(w http.ResponseWriter, r *http.Request) (*sessionState, *Identity) {
	state, err := a.readSession(r)
	if err != nil {
		http.Redirect(w, r, joinURL(a.cfg.Prefix, "login"), http.StatusFound)
		return nil, nil
	}
	identity, err := a.auth.FindIdentity(r.Context(), state.UserID)
	if err != nil {
		a.clearSession(w)
		http.Redirect(w, r, joinURL(a.cfg.Prefix, "login"), http.StatusFound)
		return nil, nil
	}
	return state, identity
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		_ = a.templates.ExecuteTemplate(w, "login", map[string]any{
			"AppName": a.cfg.AppName,
			"Title":   a.cfg.Title,
			"Prefix":  normalizePath(a.cfg.Prefix),
			"Error":   r.URL.Query().Get("error"),
			"Theme":   a.cfg.Theme,
		})
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	identity, err := a.auth.Authenticate(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Redirect(w, r, joinURL(a.cfg.Prefix, "login")+"?error="+url.QueryEscape(err.Error()), http.StatusFound)
		return
	}
	if err := a.writeSession(w, identity.ID); err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, normalizePath(a.cfg.Prefix), http.StatusFound)
}

func (a *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	if currentMethod(r) != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	state, _ := a.readSession(r)
	if state == nil || r.FormValue("_csrf") != state.CSRF {
		a.writeError(w, http.StatusBadRequest, errors.New("invalid csrf token"))
		return
	}
	a.clearSession(w)
	http.Redirect(w, r, joinURL(a.cfg.Prefix, "login"), http.StatusFound)
}

func (a *App) handleDashboard(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity) {
	data := DashboardData{
		Title:       "Dashboard",
		Description: "Go implementation inspired by Dcat Admin.",
		Cards: []DashboardCard{
			{Title: "Resources", Value: strconv.Itoa(len(a.resources)), Hint: "Registered admin modules"},
		},
	}
	if a.dashboard != nil {
		if custom, err := a.dashboard(r.Context()); err == nil {
			data = custom
		}
	}
	a.renderShell(w, r, state, identity, "dashboard", pageData{
		PageTitle:       data.Title,
		PageDescription: data.Description,
		Content:         dashboardView{Cards: data.Cards},
	})
}

func (a *App) handleGrid(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, resource Resource) {
	builder := grid.New()
	if resource.BuildGrid != nil {
		resource.BuildGrid(builder)
	}
	query := ListQuery{
		Page:      intFromQuery(r, "page", 1),
		PerPage:   intFromQuery(r, "per_page", 10),
		Search:    r.URL.Query().Get("q"),
		Sort:      r.URL.Query().Get("sort"),
		Direction: r.URL.Query().Get("direction"),
		Filters:   map[string]string{},
	}
	for _, filter := range builder.Filters {
		query.Filters[filter.Name] = r.URL.Query().Get("f_" + filter.Name)
	}

	result, err := resource.Repository.List(r.Context(), query)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
		return
	}

	baseURL := joinURL(a.cfg.Prefix, resource.Path)
	view := gridView{
		Title:         fallback(builder.Title, resource.Title),
		Description:   fallback(builder.Description, resource.Description),
		CreateURL:     joinURL(baseURL, "new"),
		EnableCreate:  !builder.DisableCreate && resource.BuildForm != nil,
		Columns:       a.buildGridColumns(r, baseURL, builder, query),
		Rows:          a.buildGridRows(baseURL, builder, result.Items),
		Filters:       a.buildGridFilters(builder, query),
		QuickSearch:   query.Search,
		CurrentPath:   baseURL,
		Pagination:    buildPagination(baseURL, query, result.Total, result.Page, result.PerPage),
		ResultSummary: fmt.Sprintf("Total %d records", result.Total),
		CSRF:          state.CSRF,
	}

	a.renderShell(w, r, state, identity, "grid", pageData{
		PageTitle:       view.Title,
		PageDescription: view.Description,
		Content:         view,
	})
}

func (a *App) handleNewForm(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, resource Resource) {
	builder := form.New()
	if resource.BuildForm == nil {
		http.NotFound(w, r)
		return
	}
	resource.BuildForm(builder)
	view := a.buildFormView(resource, builder, nil, state.CSRF, "")
	a.renderShell(w, r, state, identity, "form", pageData{
		PageTitle:       fallback(builder.Title, "Create "+resource.Title),
		PageDescription: fallback(builder.Description, resource.Description),
		Content:         view,
	})
}

func (a *App) handleEditForm(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, resource Resource, id string) {
	builder := form.New()
	if resource.BuildForm == nil {
		http.NotFound(w, r)
		return
	}
	resource.BuildForm(builder)
	record, err := resource.Repository.Get(r.Context(), id)
	if err != nil {
		a.writeError(w, http.StatusNotFound, err)
		return
	}
	view := a.buildFormView(resource, builder, record, state.CSRF, id)
	a.renderShell(w, r, state, identity, "form", pageData{
		PageTitle:       fallback(builder.Title, "Edit "+resource.Title),
		PageDescription: fallback(builder.Description, resource.Description),
		Content:         view,
	})
}

func (a *App) handleShow(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, resource Resource, id string) {
	if resource.BuildShow == nil {
		http.NotFound(w, r)
		return
	}
	builder := show.New()
	resource.BuildShow(builder)
	record, err := resource.Repository.Get(r.Context(), id)
	if err != nil {
		a.writeError(w, http.StatusNotFound, err)
		return
	}
	view := a.buildShowView(resource, builder, record, id)
	a.renderShell(w, r, state, identity, "show", pageData{
		PageTitle:       fallback(builder.Title, resource.Title+" detail"),
		PageDescription: fallback(builder.Description, resource.Description),
		Content:         view,
	})
}

func (a *App) handleTree(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, resource Resource) {
	provider, ok := resource.Repository.(TreeProvider)
	if !ok || resource.BuildTree == nil {
		http.NotFound(w, r)
		return
	}
	builder := tree.New()
	resource.BuildTree(builder)
	nodes, err := provider.Tree(r.Context())
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
		return
	}
	a.renderShell(w, r, state, identity, "tree", pageData{
		PageTitle:       fallback(builder.Title, resource.Title+" tree"),
		PageDescription: fallback(builder.Description, resource.Description),
		Content: treeView{
			Title:       fallback(builder.Title, resource.Title+" tree"),
			Description: fallback(builder.Description, resource.Description),
			EmptyText:   builder.EmptyText,
			Nodes:       nodes,
		},
	})
}

func (a *App) handleCreate(w http.ResponseWriter, r *http.Request, state *sessionState, _ *Identity, resource Resource) {
	if resource.BuildForm == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	if r.FormValue("_csrf") != state.CSRF {
		a.writeError(w, http.StatusBadRequest, errors.New("invalid csrf token"))
		return
	}
	values := a.formValues(resource)
	if err := resource.Repository.Create(r.Context(), values(r)); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	http.Redirect(w, r, joinURL(a.cfg.Prefix, resource.Path)+"?flash=created", http.StatusFound)
}

func (a *App) handleUpdate(w http.ResponseWriter, r *http.Request, state *sessionState, _ *Identity, resource Resource, id string) {
	if resource.BuildForm == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	if r.FormValue("_csrf") != state.CSRF {
		a.writeError(w, http.StatusBadRequest, errors.New("invalid csrf token"))
		return
	}
	values := a.formValues(resource)
	if err := resource.Repository.Update(r.Context(), id, values(r)); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	http.Redirect(w, r, joinURL(a.cfg.Prefix, resource.Path, id)+"?flash=updated", http.StatusFound)
}

func (a *App) handleDelete(w http.ResponseWriter, r *http.Request, state *sessionState, _ *Identity, resource Resource, id string) {
	if err := r.ParseForm(); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	if r.FormValue("_csrf") != state.CSRF {
		a.writeError(w, http.StatusBadRequest, errors.New("invalid csrf token"))
		return
	}
	if err := resource.Repository.Delete(r.Context(), id); err != nil {
		a.writeError(w, http.StatusBadRequest, err)
		return
	}
	http.Redirect(w, r, joinURL(a.cfg.Prefix, resource.Path)+"?flash=deleted", http.StatusFound)
}

func (a *App) formValues(resource Resource) func(*http.Request) map[string]string {
	builder := form.New()
	resource.BuildForm(builder)
	return func(r *http.Request) map[string]string {
		values := map[string]string{}
		for _, field := range builder.Fields {
			if field.Type == form.FieldDisplay {
				continue
			}
			raw := r.FormValue(field.Name)
			if field.Type == form.FieldPassword && raw == "" {
				continue
			}
			values[field.Name] = raw
		}
		return values
	}
}

func (a *App) renderShell(w http.ResponseWriter, r *http.Request, state *sessionState, identity *Identity, contentTemplate string, data pageData) {
	navigation, err := a.auth.Navigation(r.Context(), identity)
	if err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
		return
	}
	page := layoutData{
		AppName:         a.cfg.AppName,
		Title:           a.cfg.Title,
		PageTitle:       data.PageTitle,
		PageDescription: data.PageDescription,
		CurrentPath:     normalizePath(r.URL.Path),
		Prefix:          normalizePath(a.cfg.Prefix),
		Flash:           r.URL.Query().Get("flash"),
		Theme:           a.cfg.Theme,
		User:            identity,
		Menu:            a.menuView(navigation, normalizePath(r.URL.Path)),
		CSRF:            state.CSRF,
		ContentHTML:     a.renderPartial(contentTemplate, data.Content),
		Message:         data.Message,
	}
	if err := a.templates.ExecuteTemplate(w, "layout", page); err != nil {
		a.writeError(w, http.StatusInternalServerError, err)
	}
}

func (a *App) renderPartial(name string, data any) template.HTML {
	var buf bytes.Buffer
	if err := a.templates.ExecuteTemplate(&buf, name, data); err != nil {
		return template.HTML(template.HTMLEscapeString(err.Error()))
	}
	return template.HTML(buf.String())
}

func (a *App) writeError(w http.ResponseWriter, code int, err error) {
	http.Error(w, err.Error(), code)
}

func currentMethod(r *http.Request) string {
	if r.Method != http.MethodPost {
		return r.Method
	}
	_ = r.ParseForm()
	if override := r.FormValue("_method"); override != "" {
		return strings.ToUpper(override)
	}
	return r.Method
}

func intFromQuery(r *http.Request, key string, fallbackValue int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallbackValue
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallbackValue
	}
	return value
}

func fallback(value, other string) string {
	if value != "" {
		return value
	}
	return other
}

type pageData struct {
	PageTitle       string
	PageDescription string
	Message         string
	Content         any
}

type layoutData struct {
	AppName         string
	Title           string
	PageTitle       string
	PageDescription string
	CurrentPath     string
	Prefix          string
	Flash           string
	Theme           any
	User            *Identity
	Menu            []menuItemView
	CSRF            string
	ContentHTML     template.HTML
	Message         string
}

type menuItemView struct {
	Title    string
	Icon     string
	URL      string
	Active   bool
	Expanded bool
	Children []menuItemView
}

type dashboardView struct {
	Cards []DashboardCard
}

type gridView struct {
	Title         string
	Description   string
	QuickSearch   string
	Filters       []gridFilterView
	Columns       []gridColumnView
	Rows          []gridRowView
	Pagination    []paginationLink
	CreateURL     string
	EnableCreate  bool
	CurrentPath   string
	ResultSummary string
	CSRF          string
}

type gridFilterView struct {
	Name     string
	Label    string
	Kind     string
	Value    string
	Options  []gridOptionView
	InputKey string
}

type gridOptionView struct {
	Label    string
	Value    string
	Selected bool
}

type gridColumnView struct {
	Label    string
	Sortable bool
	SortURL  string
}

type gridRowView struct {
	Cells     []template.HTML
	ShowURL   string
	EditURL   string
	DeleteURL string
}

type paginationLink struct {
	Label   string
	URL     string
	Active  bool
	Current bool
}

type formView struct {
	Title        string
	Description  string
	Action       string
	Method       string
	DeleteAction string
	ShowDelete   bool
	BackURL      string
	SubmitLabel  string
	DeleteLabel  string
	CSRF         string
	Fields       []formFieldView
}

type formFieldView struct {
	Name        string
	Label       string
	Type        string
	Value       string
	Help        string
	Required    bool
	Readonly    bool
	Placeholder string
	Options     []gridOptionView
}

type showView struct {
	Title       string
	Description string
	Items       []showItemView
	EditURL     string
	BackURL     string
}

type showItemView struct {
	Type  string
	Label string
	Title string
	Value template.HTML
}

type treeView struct {
	Title       string
	Description string
	EmptyText   string
	Nodes       []TreeNode
}

func (a *App) buildGridColumns(r *http.Request, baseURL string, builder *grid.Builder, query ListQuery) []gridColumnView {
	var columns []gridColumnView
	for _, column := range builder.Columns {
		link := ""
		if column.Sortable {
			values := cloneQuery(r.URL.Query())
			values.Set("sort", column.Name)
			if query.Sort == column.Name && strings.EqualFold(query.Direction, "asc") {
				values.Set("direction", "desc")
			} else {
				values.Set("direction", "asc")
			}
			link = buildURL(baseURL, values)
		}
		columns = append(columns, gridColumnView{
			Label:    column.Label,
			Sortable: column.Sortable,
			SortURL:  link,
		})
	}
	return columns
}

func (a *App) buildGridRows(baseURL string, builder *grid.Builder, items []any) []gridRowView {
	rows := make([]gridRowView, 0, len(items))
	for _, item := range items {
		row := gridRowView{}
		for _, column := range builder.Columns {
			value := valueFromPath(item, column.Name)
			if column.Formatter != nil {
				row.Cells = append(row.Cells, column.Formatter(item, value))
				continue
			}
			row.Cells = append(row.Cells, toHTML(formatValue(value)))
		}
		id := formatValue(valueFromPath(item, "ID"))
		row.ShowURL = joinURL(baseURL, id)
		row.EditURL = joinURL(baseURL, id, "edit")
		row.DeleteURL = joinURL(baseURL, id, "delete")
		rows = append(rows, row)
	}
	return rows
}

func (a *App) buildGridFilters(builder *grid.Builder, query ListQuery) []gridFilterView {
	filters := make([]gridFilterView, 0, len(builder.Filters))
	for _, filter := range builder.Filters {
		view := gridFilterView{
			Name:     filter.Name,
			Label:    filter.Label,
			Kind:     string(filter.Kind),
			Value:    query.Filters[filter.Name],
			InputKey: "f_" + filter.Name,
		}
		for _, option := range filter.Options {
			view.Options = append(view.Options, gridOptionView{
				Label:    option.Label,
				Value:    option.Value,
				Selected: query.Filters[filter.Name] == option.Value,
			})
		}
		filters = append(filters, view)
	}
	return filters
}

func buildPagination(baseURL string, query ListQuery, total int64, page, perPage int) []paginationLink {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))
	if totalPages <= 1 {
		return nil
	}
	links := make([]paginationLink, 0, totalPages)
	for i := 1; i <= totalPages; i++ {
		values := url.Values{}
		values.Set("page", strconv.Itoa(i))
		values.Set("per_page", strconv.Itoa(perPage))
		if query.Search != "" {
			values.Set("q", query.Search)
		}
		if query.Sort != "" {
			values.Set("sort", query.Sort)
			values.Set("direction", query.Direction)
		}
		for key, value := range query.Filters {
			if value != "" {
				values.Set("f_"+key, value)
			}
		}
		links = append(links, paginationLink{
			Label:   strconv.Itoa(i),
			URL:     buildURL(baseURL, values),
			Active:  i == page,
			Current: i == page,
		})
	}
	return links
}

func (a *App) buildFormView(resource Resource, builder *form.Builder, record any, csrf, id string) formView {
	baseURL := joinURL(a.cfg.Prefix, resource.Path)
	view := formView{
		Title:       fallback(builder.Title, resource.Title),
		Description: fallback(builder.Description, resource.Description),
		Action:      baseURL,
		Method:      http.MethodPost,
		ShowDelete:  !builder.HideDelete && id != "",
		DeleteLabel: builder.DeleteLabel,
		SubmitLabel: builder.SubmitLabel,
		CSRF:        csrf,
		BackURL:     baseURL,
	}
	if id != "" {
		view.Action = joinURL(baseURL, id)
		view.Method = http.MethodPut
		view.DeleteAction = joinURL(baseURL, id, "delete")
	}
	for _, field := range builder.Fields {
		entry := formFieldView{
			Name:        field.Name,
			Label:       field.Label,
			Type:        string(field.Type),
			Help:        field.Help,
			Required:    field.Required,
			Readonly:    field.Readonly,
			Placeholder: field.Placeholder,
		}
		if record != nil {
			entry.Value = formatValue(valueFromPath(record, field.Name))
		}
		for _, option := range field.Options {
			entry.Options = append(entry.Options, gridOptionView{
				Label:    option.Label,
				Value:    option.Value,
				Selected: entry.Value == option.Value,
			})
		}
		view.Fields = append(view.Fields, entry)
	}
	return view
}

func (a *App) buildShowView(resource Resource, builder *show.Builder, record any, id string) showView {
	view := showView{
		Title:       fallback(builder.Title, resource.Title),
		Description: fallback(builder.Description, resource.Description),
		EditURL:     joinURL(a.cfg.Prefix, resource.Path, id, "edit"),
		BackURL:     joinURL(a.cfg.Prefix, resource.Path),
	}
	for _, item := range builder.Items {
		switch item.Type {
		case show.ItemDivider:
			view.Items = append(view.Items, showItemView{Type: "divider", Title: item.Title})
		default:
			value := valueFromPath(record, item.Name)
			rendered := toHTML(formatValue(value))
			if item.Formatter != nil {
				rendered = item.Formatter(record, value)
			}
			view.Items = append(view.Items, showItemView{
				Type:  "field",
				Label: item.Label,
				Value: rendered,
			})
		}
	}
	return view
}

func (a *App) menuView(items []NavigationItem, currentPath string) []menuItemView {
	out := make([]menuItemView, 0, len(items))
	for _, item := range items {
		url := ""
		if item.URI != "" {
			if item.URI == "/" {
				url = normalizePath(a.cfg.Prefix)
			} else if strings.HasPrefix(item.URI, "/") {
				url = normalizePath(item.URI)
			} else {
				url = joinURL(a.cfg.Prefix, item.URI)
			}
		}
		children := a.menuView(item.Children, currentPath)
		active := url != "" && (currentPath == url || strings.HasPrefix(currentPath, strings.TrimRight(url, "/")+"/"))
		expanded := active
		for _, child := range children {
			if child.Active || child.Expanded {
				expanded = true
				active = active || child.Active
			}
		}
		out = append(out, menuItemView{
			Title:    item.Title,
			Icon:     item.Icon,
			URL:      url,
			Active:   active,
			Expanded: expanded,
			Children: children,
		})
	}
	return out
}

func cloneQuery(values url.Values) url.Values {
	cloned := url.Values{}
	for key, list := range values {
		for _, value := range list {
			cloned.Add(key, value)
		}
	}
	return cloned
}
