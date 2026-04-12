package goadmin

import "context"

// Values carries submitted form values, including repeated keys.
type Values map[string][]string

// First returns the first submitted value for a key.
func (v Values) First(key string) string {
	values := v[key]
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// All returns every submitted value for a key.
func (v Values) All(key string) []string {
	return append([]string(nil), v[key]...)
}

// ListQuery drives grid pages.
type ListQuery struct {
	Page      int
	PerPage   int
	Search    string
	Sort      string
	Direction string
	Filters   map[string]string
}

// ListResult is a paginated repository response.
type ListResult struct {
	Items   []any
	Total   int64
	Page    int
	PerPage int
}

// TreeNode represents a hierarchical record.
type TreeNode struct {
	ID          string
	ParentID    string
	Title       string
	Description string
	Children    []TreeNode
}

// Repository is the core CRUD contract used by resources.
type Repository interface {
	List(ctx context.Context, query ListQuery) (ListResult, error)
	Get(ctx context.Context, id string) (any, error)
	Create(ctx context.Context, values Values) error
	Update(ctx context.Context, id string, values Values) error
	Delete(ctx context.Context, id string) error
}

// TreeProvider adds tree-page support to a repository.
type TreeProvider interface {
	Tree(ctx context.Context) ([]TreeNode, error)
}

// Identity is the authenticated admin user context.
type Identity struct {
	ID          uint
	Username    string
	DisplayName string
	Roles       []string
	Permissions map[string]bool
}

// NavigationItem is a sidebar entry.
type NavigationItem struct {
	ID       uint
	ParentID uint
	Title    string
	Icon     string
	URI      string
	Children []NavigationItem
}

// AuthService abstracts login, session reload, menu loading, and authorization.
type AuthService interface {
	Authenticate(ctx context.Context, username, password string) (*Identity, error)
	FindIdentity(ctx context.Context, id uint) (*Identity, error)
	Navigation(ctx context.Context, identity *Identity) ([]NavigationItem, error)
	Authorize(ctx context.Context, identity *Identity, permission string) (bool, error)
}

// DashboardCard is a lightweight metric widget.
type DashboardCard struct {
	Title string
	Value string
	Hint  string
}

// DashboardAction is a quick entry/button inside dashboard widgets.
type DashboardAction struct {
	Label string
	URL   string
	Style string
}

// DashboardPanelItem is one row/card inside a dashboard panel.
type DashboardPanelItem struct {
	Title       string
	Description string
	Value       string
	URL         string
	Tags        []string
}

// DashboardPanel is a reusable dashboard widget block.
type DashboardPanel struct {
	Title       string
	Description string
	EmptyText   string
	Actions     []DashboardAction
	Items       []DashboardPanelItem
}

// DashboardData powers the home page.
type DashboardData struct {
	Title       string
	Description string
	Cards       []DashboardCard
	Panels      []DashboardPanel
}
