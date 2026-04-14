package demo

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhenyangze/go-admin-build/examples/demo/models"
	demoseed "github.com/zhenyangze/go-admin-build/examples/demo/seed"
	"github.com/zhenyangze/goadmin"
	"github.com/zhenyangze/goadmin/audit"
	"github.com/zhenyangze/goadmin/auth"
	"github.com/zhenyangze/goadmin/form"
	"github.com/zhenyangze/goadmin/grid"
	"github.com/zhenyangze/goadmin/show"
	"github.com/zhenyangze/goadmin/store/gormstore"
	"github.com/zhenyangze/goadmin/tree"
	"github.com/zhenyangze/goadmin/widgets/alert"
	"github.com/zhenyangze/goadmin/widgets/chart"
	"github.com/zhenyangze/goadmin/widgets/dropdown"
	"github.com/zhenyangze/goadmin/widgets/tab"
	widgetform "github.com/zhenyangze/goadmin/widgets/form"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Build creates a fully seeded demo app backed by SQLite.
func Build(dbPath string) (*goadmin.App, *gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}
	app, err := buildWithDB(db, filepath.Join(filepath.Dir(dbPath), "uploads"))
	if err != nil {
		return nil, nil, err
	}
	return app, db, nil
}

// BuildWithDB wires the reusable admin package to the demo models.
func BuildWithDB(db *gorm.DB) (*goadmin.App, error) {
	return buildWithDB(db, "tmp/demo/uploads")
}

func buildWithDB(db *gorm.DB, uploadDir string) (*goadmin.App, error) {
	if err := migrate(db); err != nil {
		return nil, err
	}
	if err := seed(db); err != nil {
		return nil, err
	}

	app, err := goadmin.New(goadmin.Config{
		AppName:       "Go Admin Build",
		Title:         "Go Admin Build",
		Prefix:        "/admin",
		SessionSecret: "go-admin-build-demo-secret",
		UploadDir:     uploadDir,
		UploadPath:    "uploads",
	}, auth.NewService(db))
	if err != nil {
		return nil, err
	}

	app.SetDashboard(func(ctx context.Context) (goadmin.DashboardData, error) {
		var users int64
		var roles int64
		var permissions int64
		var menus int64
		var articles int64
		var categories int64
		if err := db.WithContext(ctx).Model(&auth.User{}).Count(&users).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&auth.Role{}).Count(&roles).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&auth.Permission{}).Count(&permissions).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&auth.Menu{}).Count(&menus).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&models.Article{}).Count(&articles).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&models.Category{}).Count(&categories).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		var projects int64
		if err := db.WithContext(ctx).Model(&models.Project{}).Count(&projects).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		var audits int64
		if err := db.WithContext(ctx).Model(&models.AuditLog{}).Count(&audits).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		var tickets int64
		if err := db.WithContext(ctx).Model(&models.Ticket{}).Count(&tickets).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		var reports int64
		if err := db.WithContext(ctx).Model(&models.ReportSnapshot{}).Count(&reports).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		var recentArticles []models.Article
		if err := db.WithContext(ctx).
			Preload("Category").
			Order("created_at desc, id desc").
			Limit(3).
			Find(&recentArticles).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		recentItems := make([]goadmin.DashboardPanelItem, 0, len(recentArticles))
		for _, article := range recentArticles {
			recentItems = append(recentItems, goadmin.DashboardPanelItem{
				Title:       article.Title,
				Description: fmt.Sprintf("%s · %s", article.Status, article.Category.Name),
				Value:       formatDashboardTime(article.PublishedAt),
				URL:         fmt.Sprintf("/admin/articles/%d", article.ID),
				Tags:        []string{article.Status},
			})
		}
		var reportsList []models.ReportSnapshot
		if err := db.WithContext(ctx).Order("snapshot_at desc, id desc").Limit(3).Find(&reportsList).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		operationalOverview := make([]goadmin.DashboardPanelItem, 0, len(reportsList))
		for _, report := range reportsList {
			operationalOverview = append(operationalOverview, goadmin.DashboardPanelItem{
				Title:       report.Name,
				Description: fmt.Sprintf("%s · %s", report.Metric, report.Dimension),
				Value:       fmt.Sprintf("%d", report.Value),
				URL:         fmt.Sprintf("/admin/reports/%d", report.ID),
				Tags:        []string{report.Trend},
			})
		}
		moduleGuide := []goadmin.DashboardPanelItem{
			{Title: "Users / Roles / Permissions / Menus", Description: "Built-in auth, RBAC, menu visibility, and admin table management.", Value: "auth+rbac", URL: "/admin/users", Tags: []string{"auth", "rbac", "admin"}},
			{Title: "Articles", Description: "Complex CRUD with uploads, multi-upload, tags, date range, FAQ relation, and links relation.", Value: "content", URL: "/admin/articles", Tags: []string{"complex-crud", "upload", "nested"}},
			{Title: "Projects", Description: "Business CRUD with relation-backed milestones and timeline fields.", Value: "business", URL: "/admin/projects", Tags: []string{"business", "relations"}},
			{Title: "Tickets", Description: "Pure generic GORM CRUD demo using belongs-to preload, filters, and detail pages.", Value: "generic", URL: "/admin/tickets", Tags: []string{"generic", "gormstore"}},
			{Title: "Audit Logs", Description: "Read-only operational/audit module with search, filters, and detail view.", Value: "readonly", URL: "/admin/audits", Tags: []string{"readonly", "audit"}},
			{Title: "Reports", Description: "Read-only metrics/reporting module using the generic repository in browse-only mode.", Value: "reporting", URL: "/admin/reports", Tags: []string{"readonly", "reporting"}},
			{Title: "Category / Menu Trees", Description: "Hierarchy rendering and tree navigation examples.", Value: "tree", URL: "/admin/categories/tree", Tags: []string{"tree", "navigation"}},
		}
		walkthrough := []goadmin.DashboardPanelItem{
			{Title: "1. Verify auth & menu", Description: "Open Users, Roles, Permissions, and Menus to validate built-in admin tables.", URL: "/admin/users"},
			{Title: "2. Exercise complex CRUD", Description: "Open Articles to see uploads, repeaters, tags, nested relations, and custom grid actions.", URL: "/admin/articles"},
			{Title: "3. Exercise relation business CRUD", Description: "Open Projects and create/edit milestones.", URL: "/admin/projects"},
			{Title: "4. Exercise generic CRUD", Description: "Open Tickets to see how much can be built with the plain generic repository.", URL: "/admin/tickets"},
			{Title: "5. Exercise read-only pages", Description: "Open Audit Logs and Reports to validate browse-only list/detail pages.", URL: "/admin/audits"},
			{Title: "6. Exercise tree pages", Description: "Open Categories Tree and Menus Tree to verify hierarchical rendering.", URL: "/admin/categories/tree"},
		}
		capabilities := []goadmin.DashboardPanelItem{
			{Title: "What this demo proves", Description: "Gin mount, SQLite persistence, GORM adapters, RBAC, grid/form/show/tree, uploads, repeaters, relation-backed nested editing, hooks, and smoke automation.", Value: "framework-complete"},
			{Title: "Runbook", Description: "Use scripts/run_demo.sh for a fresh start and scripts/demo_smoke.sh for one-command verification.", Value: "ops"},
			{Title: "Credentials", Description: "Default demo login remains admin / admin for fast validation.", Value: "demo-login"},
		}
		helpers := []goadmin.DashboardPanelItem{
			{Title: "Fresh reset", Description: "Run `DEMO_RESET=1 ./scripts/run_demo.sh` to rebuild the DB and uploads from seed data.", Value: "reset"},
			{Title: "Smoke validation", Description: "Run `./scripts/demo_smoke.sh` after changes to verify the full demo still works.", Value: "smoke"},
			{Title: "Uploads location", Description: "Uploads are stored next to the chosen DB path inside an `uploads/` directory.", Value: "uploads"},
		}
		matrix := []goadmin.DashboardPanelItem{
			{Title: "Auth / RBAC", Description: "Users, Roles, Permissions, Menus", Value: "users + roles + permissions + menus"},
			{Title: "Generic CRUD", Description: "Tickets module built mostly from gormstore.New + builders", Value: "tickets"},
			{Title: "Complex CRUD", Description: "Articles with uploads, tags, date ranges, repeaters, and relation-backed children", Value: "articles"},
			{Title: "Business CRUD", Description: "Projects with milestones", Value: "projects"},
			{Title: "Read-only pages", Description: "Audit Logs + Reports", Value: "readonly"},
			{Title: "Tree pages", Description: "Categories + Menus", Value: "tree"},
		}
		return goadmin.DashboardData{
			Title:       "Dashboard",
			Description: "Dcat-inspired modules running on Go + GORM + TailwindCSS.",
			Cards: []goadmin.DashboardCard{
				{Title: "Admin users", Value: fmt.Sprintf("%d", users), Hint: "Built-in auth model"},
				{Title: "Roles", Value: fmt.Sprintf("%d", roles), Hint: "RBAC seed data"},
				{Title: "Permissions", Value: fmt.Sprintf("%d", permissions), Hint: "Permission registry"},
				{Title: "Menus", Value: fmt.Sprintf("%d", menus), Hint: "Sidebar navigation tree"},
				{Title: "Articles", Value: fmt.Sprintf("%d", articles), Hint: "Grid/Form/Show demo"},
				{Title: "Categories", Value: fmt.Sprintf("%d", categories), Hint: "Tree demo"},
				{Title: "Projects", Value: fmt.Sprintf("%d", projects), Hint: "Relation-backed business demo"},
				{Title: "Audit Logs", Value: fmt.Sprintf("%d", audits), Hint: "Read-only operational demo"},
				{Title: "Tickets", Value: fmt.Sprintf("%d", tickets), Hint: "Generic GORM CRUD demo"},
				{Title: "Reports", Value: fmt.Sprintf("%d", reports), Hint: "Read-only reporting demo"},
			},
			Panels: []goadmin.DashboardPanel{
				{
					Title:       "Quick Access",
					Description: "Jump into the most-used admin modules.",
					Actions: []goadmin.DashboardAction{
						{Label: "Users", URL: "/admin/users", Style: "ghost"},
						{Label: "Articles", URL: "/admin/articles", Style: "primary"},
						{Label: "Projects", URL: "/admin/projects", Style: "ghost"},
						{Label: "Audit Logs", URL: "/admin/audits", Style: "ghost"},
						{Label: "Tickets", URL: "/admin/tickets", Style: "ghost"},
						{Label: "Reports", URL: "/admin/reports", Style: "ghost"},
						{Label: "Demo Center", URL: "/admin/demo-center", Style: "ghost"},
						{Label: "Category Tree", URL: "/admin/categories/tree", Style: "ghost"},
					},
				},
				{
					Title:       "Recent Articles",
					Description: "Latest content managed through the reusable Grid/Form/Show flow.",
					EmptyText:   "No articles yet.",
					Items:       recentItems,
					Actions: []goadmin.DashboardAction{
						{Label: "Create Article", URL: "/admin/articles/new", Style: "primary"},
					},
				},
				{
					Title:       "Module Guide",
					Description: "Each module demonstrates a different slice of the framework.",
					Items:       moduleGuide,
				},
				{
					Title:       "Suggested Walkthrough",
					Description: "Recommended order for live demos or self-checks.",
					Items:       walkthrough,
					Actions: []goadmin.DashboardAction{
						{Label: "Start with Articles", URL: "/admin/articles", Style: "ghost"},
					},
				},
				{
					Title:       "Framework Coverage",
					Description: "High-level summary of what the current demo exercises.",
					Items:       capabilities,
				},
				{
					Title:       "Operational Overview",
					Description: "Latest seeded report snapshots from the read-only reporting module.",
					Items:       operationalOverview,
					Actions: []goadmin.DashboardAction{
						{Label: "Open Reports", URL: "/admin/reports", Style: "ghost"},
					},
				},
				{
					Title:       "Demo Helpers",
					Description: "Operational commands and reset behavior for local demo runs.",
					Items:       helpers,
					Actions: []goadmin.DashboardAction{
						{Label: "Demo Center", URL: "/admin/demo-center", Style: "primary"},
					},
				},
				{
					Title:       "Capability Matrix",
					Description: "Which module best demonstrates each framework slice.",
					Items:       matrix,
				},
			},
		}, nil
	})

	app.RegisterDashboardPage(goadmin.DashboardPage{
		Path:        "demo-center",
		Title:       "Demo Center",
		Description: "Operational helpers, reset/export shortcuts, and a condensed module verification guide.",
		Build: func(ctx context.Context, r *http.Request, _ *goadmin.Identity) (goadmin.DashboardData, error) {
			return goadmin.DashboardData{
				Title:       "Demo Center",
				Description: "Use this page to reset, export, and quickly verify the demo instance.",
				Panels: []goadmin.DashboardPanel{
					{
						Title:       "Reset & Run",
						Description: "Fastest way to get back to a clean seed state.",
						Items: []goadmin.DashboardPanelItem{
							{Title: "Fresh run command", Description: "Rebuild the selected DB and uploads before startup.", Value: "DEMO_RESET=1 ./scripts/run_demo.sh"},
							{Title: "Fresh DB path", Description: "Use a dedicated DB file for clean demos.", Value: "DEMO_DB_PATH=tmp/demo/fresh-admin.db"},
						},
					},
					{
						Title:       "Export",
						Description: "Download a JSON snapshot of the current demo dataset.",
						Actions: []goadmin.DashboardAction{
							{Label: "Export JSON", URL: "/admin/demo-export.json", Style: "primary"},
						},
						Items: []goadmin.DashboardPanelItem{
							{Title: "Format", Description: "JSON export includes counts and primary business tables.", Value: "json"},
						},
					},
					{
						Title:       "Verify Modules",
						Description: "Recommended quick links for live demonstrations.",
						Items: []goadmin.DashboardPanelItem{
							{Title: "Articles", Description: "Uploads, nested relations, and custom renderers.", URL: "/admin/articles", Tags: []string{"complex-crud"}},
							{Title: "Projects", Description: "Business CRUD with milestones.", URL: "/admin/projects", Tags: []string{"business"}},
							{Title: "Tickets", Description: "Generic repository CRUD.", URL: "/admin/tickets", Tags: []string{"generic"}},
							{Title: "Reports", Description: "Read-only metrics.", URL: "/admin/reports", Tags: []string{"reporting"}},
						},
					},
				},
			}, nil
		},
	})

	app.RegisterRoute(goadmin.Route{
		Path:    "demo-export.json",
		Methods: []string{"GET"},
		Handler: func(w http.ResponseWriter, r *http.Request, _ *goadmin.Identity) error {
			var payload struct {
				Articles []models.Article        `json:"articles"`
				Projects []models.Project        `json:"projects"`
				Tickets  []models.Ticket         `json:"tickets"`
				Reports  []models.ReportSnapshot `json:"reports"`
				Audits   []models.AuditLog       `json:"audits"`
				Counts   map[string]int64        `json:"counts"`
			}
			db.WithContext(r.Context()).Preload("Category").Preload("FAQs").Preload("Links").Find(&payload.Articles)
			db.WithContext(r.Context()).Preload("Milestones").Find(&payload.Projects)
			db.WithContext(r.Context()).Preload("Project").Find(&payload.Tickets)
			db.WithContext(r.Context()).Find(&payload.Reports)
			db.WithContext(r.Context()).Find(&payload.Audits)
			payload.Counts = map[string]int64{
				"articles": int64(len(payload.Articles)),
				"projects": int64(len(payload.Projects)),
				"tickets":  int64(len(payload.Tickets)),
				"reports":  int64(len(payload.Reports)),
				"audits":   int64(len(payload.Audits)),
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			return json.NewEncoder(w).Encode(payload)
		},
	})

	registerUsers(app, db)
	registerRoles(app, db)
	registerPermissions(app, db)
	registerMenus(app, db)
	registerArticles(app, db)
	registerCategories(app, db)
	registerProjects(app, db)
	registerAuditLogs(app, db)
	registerLoginLogs(app, db)
	registerTickets(app, db)
	registerReports(app, db)

	// Register a Tool Form example - Settings form
	settingsForm := widgetform.New().
		Title("系统设置").
		Description("配置系统基本参数").
		Text("site_name", "网站名称").
		Text("site_description", "网站描述").
		Email("admin_email", "管理员邮箱").
		Switch("maintenance_mode", "维护模式").
		Select("timezone", "时区",
			form.Option{Value: "Asia/Shanghai", Label: "北京时间"},
			form.Option{Value: "Asia/Tokyo", Label: "东京时间"},
			form.Option{Value: "America/New_York", Label: "纽约时间"},
		).
		Number("items_per_page", "每页条目数").
		DefaultValue("items_per_page", "20").
		DefaultValue("timezone", "Asia/Shanghai").
		Handle(func(ctx context.Context, values url.Values) error {
			// Process form data
			fmt.Printf("Settings updated: %+v\n", values)
			return nil
		}).
		Success("设置已保存", "alert('设置保存成功！');")

	app.RegisterToolForm("settings", settingsForm)

	// Register Widgets Demo Page
	registerWidgetsDemo(app)

	return app, nil
}

func migrate(db *gorm.DB) error {
	if err := auth.AutoMigrate(db); err != nil {
		return err
	}
	// 迁移登录日志和审计日志表
	if err := db.AutoMigrate(&auth.LoginLog{}, &audit.AuditLog{}); err != nil {
		return err
	}
	return db.AutoMigrate(&models.Category{}, &models.Article{}, &models.ArticleFAQ{}, &models.ArticleLink{}, &models.Project{}, &models.ProjectMilestone{}, &models.AuditLog{}, &models.Ticket{}, &models.ReportSnapshot{})
}

func seed(db *gorm.DB) error {
	return demoseed.Run(db)
}

// registerWidgetsDemo registers a dashboard page showcasing all widget components
func registerWidgetsDemo(app *goadmin.App) {
	app.RegisterDashboardPage(goadmin.DashboardPage{
		Path:        "widgets-demo",
		Title:       "组件演示",
		Description: "展示 GoAdmin 提供的各种页面组件（Widgets）",
		Build: func(ctx context.Context, r *http.Request, identity *goadmin.Identity) (goadmin.DashboardData, error) {
			// Create sample charts
			_ = chart.Line().
				Title("访问量趋势").
				Subtitle("近6个月数据").
				Labels("1月", "2月", "3月", "4月", "5月", "6月").
				Dataset("2024", []float64{120, 190, 300, 500, 200, 320}).
				Dataset("2023", []float64{100, 150, 250, 400, 180, 280}).
				SmoothLine().
				Height("250px").
				Render()

			_ = chart.Bar().
				Title("产品销量").
				Labels("产品A", "产品B", "产品C", "产品D", "产品E").
				Dataset("销量", []float64{150, 230, 180, 320, 290}).
				Height("200px").
				Render()

			_ = chart.Pie().
				Title("用户分布").
				Labels("北京", "上海", "广州", "深圳", "其他").
				Dataset("用户", []float64{30, 25, 20, 15, 10}).
				Colors("#FF6384", "#36A2EB", "#FFCE56", "#4BC0C0", "#9966FF").
				Height("200px").
				Render()

			// Create tabs
			tabsBuilder := tab.New()
			tabsBuilder.Add("图表展示", template.HTML(`<div style="padding: 1rem;"><p>图表组件支持折线图、柱状图、饼图、雷达图等多种类型，基于 Chart.js 实现。</p></div>`))
			tabsBuilder.Add("表单组件", template.HTML(`<div style="padding: 1rem;"><p>表单组件支持文本、选择、日期、文件上传等 30+ 种字段类型。</p></div>`))
			tabsBuilder.Add("表格组件", template.HTML(`<div style="padding: 1rem;"><p>表格组件支持排序、筛选、分页、批量操作等功能。</p></div>`))
			_ = tabsBuilder.Render()

			// Create dropdown
			ddBuilder := dropdown.New("操作").Class("btn-primary")
			ddBuilder.Button("查看详情", "#")
			ddBuilder.Button("编辑", "#")
			ddBuilder.Divider()
			ddBuilder.Link("导出Excel", "#")
			ddBuilder.Link("导出PDF", "#")
			ddBuilder.Divider()
			deleteItem := ddBuilder.Link("删除", "#")
			deleteItem.WithConfirm("确定要删除吗？").Danger()
			_ = ddBuilder.Render()

			// Create alerts
			_ = alert.Success("操作已成功完成！").Title("成功").Dismissible(true).Render()
			_ = alert.Info("这是一条提示信息，用于向用户展示一般性说明。").Dismissible(true).Render()
			_ = alert.Warning("请注意：此操作将影响多个关联数据。").Dismissible(true).Render()

			return goadmin.DashboardData{
				Title:       "组件演示中心",
				Description: "体验 GoAdmin 提供的丰富组件库",
				Panels: []goadmin.DashboardPanel{
					{
						Title:       "图表组件 Chart",
						Description: "支持多种图表类型，数据可视化",
						Items: []goadmin.DashboardPanelItem{
							{Title: "折线图", Description: "Line Chart - 展示趋势变化", Value: "支持多系列、平滑曲线"},
							{Title: "柱状图", Description: "Bar Chart - 展示分类对比", Value: "支持水平/垂直方向"},
							{Title: "饼图", Description: "Pie Chart - 展示占比分布", Value: "支持环形图、玫瑰图"},
						},
					},
					{
						Title:       "选项卡组件 Tab",
						Description: "组织内容到多个选项卡面板",
						Items: []goadmin.DashboardPanelItem{
							{Title: "横向选项卡", Description: "Horizontal Tabs - 顶部导航", Value: "适合内容分类展示"},
							{Title: "纵向选项卡", Description: "Vertical Tabs - 侧边导航", Value: "适合多步骤流程"},
						},
					},
					{
						Title:       "下拉菜单 Dropdown",
						Description: "操作按钮下拉菜单",
						Items: []goadmin.DashboardPanelItem{
							{Title: "基础下拉", Description: "支持图标、分割线", Value: "确认对话框"},
							{Title: "右键菜单", Description: "支持自定义触发方式", Value: "多级菜单"},
						},
					},
					{
						Title:       "警告提示 Alert",
						Description: "内联警告和 Toast 通知",
						Items: []goadmin.DashboardPanelItem{
							{Title: "内联警告", Description: "Inline Alert - 页面内显示", Value: "支持四种类型"},
							{Title: "Toast 通知", Description: "Toast Notification - 浮动提示", Value: "自动消失"},
						},
					},
					{
						Title:       "异步加载 Async",
						Description: "动态加载内容",
						Items: []goadmin.DashboardPanelItem{
							{Title: "异步卡片", Description: "Card with async content", Value: "支持加载状态"},
							{Title: "自动刷新", Description: "Auto-refresh", Value: "定时更新数据"},
						},
					},
				},
			}, nil
		},
	})

	// Register widget demo API endpoint
	app.RegisterRoute(goadmin.Route{
		Path:    "widgets-demo/data",
		Methods: []string{"GET"},
		Handler: func(w http.ResponseWriter, r *http.Request, identity *goadmin.Identity) error {
			data := map[string]interface{}{
				"timestamp": time.Now().Format("2006-01-02 15:04:05"),
				"charts": map[string]interface{}{
					"visits": []int{120, 190, 300, 500, 200, 320},
					"sales":  []int{80, 120, 180, 250, 150, 220},
				},
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			return json.NewEncoder(w).Encode(data)
		},
	})
}

func registerUsers(app *goadmin.App, db *gorm.DB) {
	repo := auth.NewUserRepository(db)

	app.Register(goadmin.Resource{
		Name:           "users",
		Path:           "users",
		Title:          "Users",
		Description:    "Built-in admin accounts",
		Permission:     "users.manage",
		EmptyText:      "No admin users yet.",
		CapabilityTags: []string{"auth", "rbac"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Username", "Username").SortableColumn()
			b.Column("Name", "Name")
			b.Column("Roles", "Roles")
			b.Column("CreatedAt", "Created")
			b.QuickSearch("Username", "Name")
			b.Filter("Username", "Username", grid.FilterText)
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Username", "Username").MarkRequired()
			b.Text("Name", "Display Name").MarkRequired()
			b.Password("Password", "Password").WithHelp("Required on create. Leave blank on edit to keep the current password.")
			b.MultiSelect("RoleIDs", "Roles", roleOptions(db)...).MarkRequired().ValueFrom("Roles")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Username", "Username")
			b.Field("Name", "Display Name")
			b.Field("Roles", "Roles")
			b.Divider("Meta")
			b.Field("CreatedAt", "Created At")
			b.Field("UpdatedAt", "Updated At")
		},
	})
}

func registerRoles(app *goadmin.App, db *gorm.DB) {
	repo := auth.NewRoleRepository(db)

	app.Register(goadmin.Resource{
		Name:           "roles",
		Path:           "roles",
		Title:          "Roles",
		Description:    "RBAC roles",
		Permission:     "roles.manage",
		EmptyText:      "No roles yet.",
		CapabilityTags: []string{"auth", "rbac"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name")
			b.Column("Slug", "Slug").SortableColumn()
			b.Column("Permissions", "Permissions")
			b.Column("Menus", "Menus")
			b.QuickSearch("Name", "Slug")
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Name", "Name").MarkRequired()
			b.Text("Slug", "Slug").MarkRequired()
			b.MultiSelect("PermissionIDs", "Permissions", permissionIDOptions(db)...).ValueFrom("Permissions")
			b.MultiSelect("MenuIDs", "Menus", menuOptions(db)...).ValueFrom("Menus")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Slug", "Slug")
			b.Field("Permissions", "Permissions")
			b.Field("Menus", "Menus")
			b.Field("CreatedAt", "Created At")
		},
	})
}

func registerPermissions(app *goadmin.App, db *gorm.DB) {
	repo := auth.NewPermissionRepository(db)

	app.Register(goadmin.Resource{
		Name:           "permissions",
		Path:           "permissions",
		Title:          "Permissions",
		Description:    "Built-in permission registry.",
		Permission:     "permissions.manage",
		EmptyText:      "No permissions yet.",
		CapabilityTags: []string{"auth", "rbac"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name")
			b.Column("Slug", "Slug").SortableColumn()
			b.Column("Path", "Path")
			b.QuickSearch("Name", "Slug", "Path")
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Name", "Name").MarkRequired()
			b.Text("Slug", "Slug").MarkRequired()
			b.Text("Path", "Path").WithPlaceholder("/admin/articles*")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Slug", "Slug")
			b.Field("Path", "Path")
			b.Field("CreatedAt", "Created At")
		},
	})
}

func registerMenus(app *goadmin.App, db *gorm.DB) {
	repo := auth.NewMenuRepository(db)

	app.Register(goadmin.Resource{
		Name:           "menus",
		Path:           "menus",
		Title:          "Menus",
		Description:    "Sidebar menu management with flat and tree views.",
		Permission:     "menus.manage",
		EmptyText:      "No menus yet.",
		CapabilityTags: []string{"auth", "navigation", "tree"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Title", "Title")
			b.Column("ParentID", "Parent ID")
			b.Column("Order", "Order").SortableColumn()
			b.Column("URI", "URI")
			b.Column("PermissionSlug", "Permission")
			b.QuickSearch("Title", "URI", "PermissionSlug")
			b.PageAction("Tree View", "/admin/menus/tree").WithStyle(grid.ActionGhost)
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Select("ParentID", "Parent", append([]form.Option{{Value: "0", Label: "Root"}}, menuOptions(db)...)...)
			b.Text("Order", "Order").WithPlaceholder("10")
			b.Text("Title", "Title").MarkRequired()
			b.Text("Icon", "Icon").WithPlaceholder("shield-check")
			b.Text("URI", "URI").WithPlaceholder("users or /")
			b.Select("PermissionSlug", "Permission", append([]form.Option{{Value: "", Label: "Public / inherited"}}, permissionSlugOptions(db)...)...)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Title", "Title")
			b.Field("ParentID", "Parent ID")
			b.Field("Order", "Order")
			b.Field("URI", "URI")
			b.Field("PermissionSlug", "Permission")
		},
		BuildTree: func(b *tree.Builder) {
			b.Title = "Menu Tree"
			b.Description = "Built-in sidebar navigation hierarchy."
			b.Branch("Title", "URI")
		},
	})
}

func registerArticles(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.NewWithRelations[models.Article](db).
		AddRelation(gormstore.NewRepeaterRelation[models.Article, models.ArticleFAQ](gormstore.RepeaterRelationConfig[models.ArticleFAQ]{
			FormKey:          "FAQ",
			ForeignKeyField:  "ArticleID",
			ForeignKeyColumn: "article_id",
			SortField:        "Sort",
			UpdateStrategy:   gormstore.RelationReplace,
			DeleteMissing:    true,
		})).
		AddRelation(gormstore.NewRepeaterRelation[models.Article, models.ArticleLink](gormstore.RepeaterRelationConfig[models.ArticleLink]{
			FormKey:          "Links",
			ForeignKeyField:  "ArticleID",
			ForeignKeyColumn: "article_id",
			SortField:        "Sort",
			IDField:          "ID",
			UpdateStrategy:   gormstore.RelationMerge,
			DeleteMissing:    false,
		}))
	repo.SearchFields = []string{"Title", "Summary"}
	repo.FilterFields = []string{"Status", "CategoryID"}
	repo.Preloads = []string{"Category", "FAQs", "Links"}
	repo.DefaultOrder = "id desc"

	// 添加审计日志 Hook
	auditHook := audit.NewHook(db, "articles")
	auditHook.SetIdentityFunc(func(ctx context.Context) *goadmin.Identity {
		if v := ctx.Value("identity"); v != nil {
			if id, ok := v.(*goadmin.Identity); ok {
				return id
			}
		}
		return nil
	})
	repo.AddHook(auditHook)

	app.Register(goadmin.Resource{
		Name:           "articles",
		Path:           "articles",
		Title:          "Articles",
		Description:    "Grid / Form / Show demo with GORM relation loading.",
		Permission:     "articles.manage",
		EmptyText:      "No articles yet.",
		CapabilityTags: []string{"complex-crud", "upload", "nested"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Title", "Title").SortableColumn()
			b.Column("Status", "Status").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-warn"
				if text == "published" {
					class = "badge badge-ok"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("Featured", "Featured").Display(func(_ any, value any) template.HTML {
				if featured, ok := value.(bool); ok && featured {
					return template.HTML(`<span class="badge badge-ok">Yes</span>`)
				}
				return template.HTML(`<span class="badge badge-warn">No</span>`)
			})
			b.Column("Category.Name", "Category")
			b.Column("Image", "Image").Display(func(_ any, value any) template.HTML {
				return renderAssetPreview(fmt.Sprint(value), "h-10 w-10 rounded object-cover")
			})
			b.Column("Gallery", "Gallery").Display(func(_ any, value any) template.HTML {
				items := nonEmptyCSV(fmt.Sprint(value))
				if len(items) == 0 {
					return template.HTML("—")
				}
				return template.HTML(template.HTMLEscapeString(fmt.Sprintf("%d files", len(items))))
			})
			b.Column("FAQs", "FAQ").Display(func(_ any, value any) template.HTML {
				return template.HTML(template.HTMLEscapeString(fmt.Sprintf("%d items", faqCount(value))))
			})
			b.Column("Links", "Links").Display(func(_ any, value any) template.HTML {
				return template.HTML(template.HTMLEscapeString(fmt.Sprintf("%d links", linkCount(value))))
			})
			b.Column("Tags", "Tags").Display(func(_ any, value any) template.HTML {
				return renderTagBadges(fmt.Sprint(value))
			})
			b.Column("VisibleFrom", "Visible Range").Display(func(record any, _ any) template.HTML {
				article, ok := record.(models.Article)
				if !ok {
					return template.HTML("—")
				}
				if article.VisibleFrom.IsZero() && article.VisibleTo.IsZero() {
					return template.HTML("—")
				}
				return template.HTML(template.HTMLEscapeString(fmt.Sprintf("%s → %s", formatDashboardDate(article.VisibleFrom), formatDashboardDate(article.VisibleTo))))
			})
			b.Column("PublishedAt", "Published At")
			b.Column("CreatedAt", "Created")
			b.QuickSearch("Title", "Summary")
			b.Filter("Status", "Status", grid.FilterSelect).WithOptions(
				grid.Option{Value: "draft", Label: "Draft"},
				grid.Option{Value: "published", Label: "Published"},
			)
			// Page Actions with dropdown-style grouping
			b.PageAction("Published", "/admin/articles?f_Status=published").WithStyle(grid.ActionGhost)
			b.PageAction("Related Tickets", "/admin/tickets").WithStyle(grid.ActionGhost)
			// Row Actions
			b.RowAction("Category", func(record any) string {
				article, ok := record.(models.Article)
				if !ok || article.CategoryID == 0 {
					return ""
				}
				return fmt.Sprintf("/admin/categories/%d", article.CategoryID)
			})
			// Use tools with outline style
			b.UseToolsWithOutline()
			// Custom tool - refresh button is shown by default, but we can add custom tools
			// b.HideRefresh() // Uncomment to hide default refresh button
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Title", "Title").MarkRequired()
			b.Textarea("Summary", "Summary").MarkRequired()
			b.Select("Status", "Status",
				form.Option{Value: "draft", Label: "Draft"},
				form.Option{Value: "published", Label: "Published"},
			)
			b.Switch("Featured", "Featured").WithHelp("Highlight this article on overview pages.")
			b.Upload("Image", "Image").AllowExtensions(".png", ".jpg", ".jpeg", ".webp").MaxSize(2 << 20).WithHelp("Upload a local image/file for this article.")
			b.MultiUpload("Gallery", "Gallery").AllowExtensions(".png", ".jpg", ".jpeg", ".webp").MaxSize(2 << 20).WithHelp("Upload multiple gallery images/files.")
			faq := b.Repeater("FAQ", "FAQ")
			faq.Text("Question", "Question").WithPlaceholder("Question")
			faq.Textarea("Answer", "Answer").WithPlaceholder("Answer")
			faq.MinRows(2).ValueFrom("FAQs")
			links := b.Repeater("Links", "Links")
			links.Hidden("ID")
			links.Text("Label", "Label").WithPlaceholder("Label")
			links.Text("URL", "URL").WithPlaceholder("https://example.com")
			links.MinRows(2).ValueFrom("Links")
			b.Tags("Tags", "Tags").WithPlaceholder("go,admin,tailwind").WithHelp("Comma-separated tags.")
			b.Datetime("PublishedAt", "Published At").WithHelp("Optional; useful for scheduled/published content.")
			b.DateRange("VisibleFrom", "VisibleTo", "Visible Range").
				WithHelp("Optional date window for visibility.").
				WithSecondPlaceholder("End date")
			b.Select("CategoryID", "Category", categoryOptions(db)...)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Title", "Title")
			b.Field("Summary", "Summary")
			b.Divider("Publishing")
			b.Field("Status", "Status")
			b.Field("Featured", "Featured")
			b.Field("Image", "Image").Display(func(_ any, value any) template.HTML {
				return renderAssetPreview(fmt.Sprint(value), "max-w-xs rounded border")
			})
			b.Field("Gallery", "Gallery").Display(func(_ any, value any) template.HTML {
				return renderGalleryList(fmt.Sprint(value))
			})
			b.Field("FAQs", "FAQ").Display(func(_ any, value any) template.HTML {
				return renderFAQList(value)
			})
			b.Field("Links", "Links").Display(func(_ any, value any) template.HTML {
				return renderLinkList(value)
			})
			b.Field("Tags", "Tags").Display(func(_ any, value any) template.HTML {
				return renderTagBadges(fmt.Sprint(value))
			})
			b.Field("PublishedAt", "Published At")
			b.Field("VisibleFrom", "Visible From")
			b.Field("VisibleTo", "Visible To")
			b.Field("Category.Name", "Category")
			b.Field("CreatedAt", "Created At")
			b.Field("UpdatedAt", "Updated At")
		},
	})
}

func registerCategories(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[models.Category](db)
	repo.SearchFields = []string{"Name", "Description"}
	repo.DefaultOrder = "sort asc, id asc"
	repo.TreeConfig = &gormstore.TreeConfig{
		IDField:          "ID",
		ParentField:      "ParentID",
		TitleField:       "Name",
		DescriptionField: "Description",
		OrderField:       "sort",
	}

	app.Register(goadmin.Resource{
		Name:           "categories",
		Path:           "categories",
		Title:          "Categories",
		Description:    "Tree and flat management demo.",
		Permission:     "categories.manage",
		EmptyText:      "No categories yet.",
		CapabilityTags: []string{"tree", "generic-crud"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name")
			b.Column("ParentID", "Parent ID")
			b.Column("Sort", "Sort").SortableColumn()
			b.PageAction("Tree View", "/admin/categories/tree").WithStyle(grid.ActionGhost)
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Name", "Name").MarkRequired()
			b.Textarea("Description", "Description")
			b.Select("ParentID", "Parent", append([]form.Option{{Value: "0", Label: "Root"}}, categoryOptions(db)...)...)
			b.Text("Sort", "Sort").WithPlaceholder("10")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Description", "Description")
			b.Field("ParentID", "Parent ID")
			b.Field("Sort", "Sort")
		},
		BuildTree: func(b *tree.Builder) {
			b.Title = "Category Tree"
			b.Description = "Hierarchical categories rendered from the generic tree repository."
			b.Branch("Name", "Description")
		},
	})
}

func registerProjects(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.NewWithRepeater[models.Project, models.ProjectMilestone](db, gormstore.RepeaterRelationConfig[models.ProjectMilestone]{
		FormKey:          "Milestones",
		ForeignKeyField:  "ProjectID",
		ForeignKeyColumn: "project_id",
		SortField:        "Sort",
		UpdateStrategy:   gormstore.RelationReplace,
		DeleteMissing:    true,
	})
	repo.SearchFields = []string{"Name", "Owner", "Description"}
	repo.FilterFields = []string{"Status"}
	repo.Preloads = []string{"Milestones"}
	repo.DefaultOrder = "id desc"

	app.Register(goadmin.Resource{
		Name:           "projects",
		Path:           "projects",
		Title:          "Projects",
		Description:    "Business-style demo module with relation-backed milestones.",
		Permission:     "projects.manage",
		EmptyText:      "No projects yet.",
		CapabilityTags: []string{"business-crud", "relations"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name").SortableColumn()
			b.Column("Owner", "Owner")
			b.Column("Status", "Status").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-warn"
				if text == "active" {
					class = "badge badge-ok"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("Budget", "Budget")
			b.Column("Milestones", "Milestones").Display(func(_ any, value any) template.HTML {
				switch typed := value.(type) {
				case []models.ProjectMilestone:
					return template.HTML(template.HTMLEscapeString(fmt.Sprintf("%d items", len(typed))))
				default:
					return template.HTML("0 items")
				}
			})
			b.QuickSearch("Name", "Owner", "Description")
			b.Filter("Status", "Status", grid.FilterSelect).WithOptions(
				grid.Option{Value: "planning", Label: "Planning"},
				grid.Option{Value: "active", Label: "Active"},
				grid.Option{Value: "done", Label: "Done"},
			)
			b.PageAction("Open Tickets", "/admin/tickets").WithStyle(grid.ActionGhost)
			b.RowAction("Tickets", func(record any) string {
				project, ok := record.(models.Project)
				if !ok {
					return ""
				}
				return fmt.Sprintf("/admin/tickets?f_ProjectID=%d", project.ID)
			})
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Name", "Name").MarkRequired()
			b.Text("Owner", "Owner").MarkRequired()
			b.Select("Status", "Status",
				form.Option{Value: "planning", Label: "Planning"},
				form.Option{Value: "active", Label: "Active"},
				form.Option{Value: "done", Label: "Done"},
			)
			b.Text("Budget", "Budget").WithPlaceholder("100000")
			b.Date("StartsAt", "Start Date")
			b.Date("EndsAt", "End Date")
			b.Textarea("Description", "Description")
			milestones := b.Repeater("Milestones", "Milestones")
			milestones.Text("Title", "Title").WithPlaceholder("Milestone title")
			milestones.Date("Deadline", "Deadline")
			milestones.MinRows(2).ValueFrom("Milestones")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Owner", "Owner")
			b.Field("Status", "Status")
			b.Field("Budget", "Budget")
			b.Field("StartsAt", "Start Date")
			b.Field("EndsAt", "End Date")
			b.Field("Description", "Description")
			b.Field("Milestones", "Milestones").Display(func(_ any, value any) template.HTML {
				switch typed := value.(type) {
				case []models.ProjectMilestone:
					if len(typed) == 0 {
						return template.HTML("—")
					}
					var buf strings.Builder
					buf.WriteString(`<div class="stack">`)
					for _, milestone := range typed {
						buf.WriteString(fmt.Sprintf(`<div class="detail-row"><div><div class="detail-label">%s</div><div class="metric-hint">%s</div></div></div>`,
							template.HTMLEscapeString(milestone.Title),
							template.HTMLEscapeString(formatDashboardDate(milestone.Deadline)),
						))
					}
					buf.WriteString(`</div>`)
					return template.HTML(buf.String())
				default:
					return template.HTML("—")
				}
			})
		},
	})
}

func registerAuditLogs(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[models.AuditLog](db)
	repo.SearchFields = []string{"Actor", "Action", "Resource", "Detail"}
	repo.FilterFields = []string{"Level", "Resource"}
	repo.DefaultOrder = "created_at desc, id desc"

	app.Register(goadmin.Resource{
		Name:           "audits",
		Path:           "audits",
		Title:          "Audit Logs",
		Description:    "Read-only operational activity feed for the demo environment.",
		Permission:     "audits.view",
		EmptyText:      "No audit log entries yet.",
		CapabilityTags: []string{"readonly", "audit"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.DisableCreate = true
			b.DisableEdit = true
			b.DisableDelete = true
			b.Column("ID", "ID").SortableColumn()
			b.Column("Actor", "Actor")
			b.Column("Action", "Action")
			b.Column("Resource", "Resource")
			b.Column("Level", "Level").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-ok"
				if text == "warning" {
					class = "badge badge-warn"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("Detail", "Detail").Display(func(_ any, value any) template.HTML {
				text := strings.TrimSpace(fmt.Sprint(value))
				if len(text) > 80 {
					text = text[:80] + "…"
				}
				return template.HTML(template.HTMLEscapeString(text))
			})
			b.Column("IP", "IP")
			b.Column("CreatedAt", "At").SortableColumn()
			b.QuickSearch("Actor", "Action", "Resource", "Detail")
			b.Filter("Level", "Level", grid.FilterSelect).WithOptions(
				grid.Option{Value: "info", Label: "Info"},
				grid.Option{Value: "warning", Label: "Warning"},
			)
			b.Filter("Resource", "Resource", grid.FilterSelect).WithOptions(
				grid.Option{Value: "article", Label: "Article"},
				grid.Option{Value: "project", Label: "Project"},
				grid.Option{Value: "demo", Label: "Demo"},
			)
			b.PageAction("Open Reports", "/admin/reports").WithStyle(grid.ActionGhost)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Actor", "Actor")
			b.Field("Action", "Action")
			b.Field("Resource", "Resource")
			b.Field("ResourceID", "Resource ID")
			b.Field("Level", "Level")
			b.Field("IP", "IP")
			b.Field("Detail", "Detail")
			b.Field("CreatedAt", "Created At")
		},
	})
}


func registerLoginLogs(app *goadmin.App, db *gorm.DB) {
	repo := auth.NewLoginLogRepository(db)

	app.Register(goadmin.Resource{
		Name:           "login-logs",
		Path:           "login-logs",
		Title:          "Login Logs",
		Description:    "User authentication activity log",
		Permission:     "login-logs.view",
		EmptyText:      "No login log entries yet.",
		CapabilityTags: []string{"audit", "security"},
		Repository:     repo,
		BuildGrid: func(b *grid.Builder) {
			b.DisableCreate = true
			b.DisableEdit = true
			b.DisableDelete = false

			b.Column("ID", "ID").SortableColumn()
			b.Column("Username", "Username")
			b.Column("Action", "Action").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-info"
				switch text {
				case "login":
					class = "badge badge-success"
				case "logout":
					class = "badge badge-info"
				case "failed":
					class = "badge badge-danger"
				case "locked":
					class = "badge badge-warning"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("IP", "IP")
			b.Column("UserAgent", "Browser").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				if len(text) > 50 {
					text = text[:50] + "..."
				}
				return template.HTML(template.HTMLEscapeString(text))
			})
			b.Column("Reason", "Reason").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				if text == "" {
					return template.HTML("-")
				}
				return template.HTML(template.HTMLEscapeString(text))
			})
			b.Column("CreatedAt", "Time").SortableColumn()

			b.QuickSearch("Username", "IP", "UserAgent")
			b.Filter("Action", "Action", grid.FilterSelect).WithOptions(
				grid.Option{Value: "login", Label: "Login"},
				grid.Option{Value: "logout", Label: "Logout"},
				grid.Option{Value: "failed", Label: "Failed"},
				grid.Option{Value: "locked", Label: "Locked"},
			)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("UserID", "User ID")
			b.Field("Username", "Username")
			b.Field("Action", "Action")
			b.Field("IP", "IP")
			b.Field("UserAgent", "User Agent")
			b.Field("Reason", "Reason")
			b.Field("CreatedAt", "Created At")
		},
	})
}


func registerTickets(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[models.Ticket](db)
	repo.SearchFields = []string{"Title", "Assignee", "Description"}
	repo.FilterFields = []string{"Status", "Priority", "ProjectID"}
	repo.Preloads = []string{"Project"}
	repo.DefaultOrder = "id desc"

	app.Register(goadmin.Resource{
		Name:           "tickets",
		Path:           "tickets",
		Title:          "Tickets",
		Description:    "Generic GORM CRUD demo with belongs-to relation and filters.",
		Permission:     "tickets.manage",
		EmptyText:      "No tickets yet.",
		CapabilityTags: []string{"generic-crud", "belongs-to"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Title", "Title").SortableColumn()
			b.Column("Project.Name", "Project")
			b.Column("Assignee", "Assignee")
			b.Column("Priority", "Priority").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-ok"
				if text == "high" {
					class = "badge badge-warn"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("Status", "Status").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-warn"
				if text == "done" {
					class = "badge badge-ok"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("DueDate", "Due Date")
			b.QuickSearch("Title", "Assignee", "Description")
			b.Filter("Status", "Status", grid.FilterSelect).WithOptions(
				grid.Option{Value: "open", Label: "Open"},
				grid.Option{Value: "in_progress", Label: "In Progress"},
				grid.Option{Value: "done", Label: "Done"},
			)
			b.Filter("Priority", "Priority", grid.FilterSelect).WithOptions(
				grid.Option{Value: "low", Label: "Low"},
				grid.Option{Value: "medium", Label: "Medium"},
				grid.Option{Value: "high", Label: "High"},
			)
			b.PageAction("Open Projects", "/admin/projects").WithStyle(grid.ActionGhost)
			b.PageAction("Open Reports", "/admin/reports").WithStyle(grid.ActionGhost)
			b.RowAction("Project", func(record any) string {
				ticket, ok := record.(models.Ticket)
				if !ok || ticket.ProjectID == 0 {
					return ""
				}
				return fmt.Sprintf("/admin/projects/%d", ticket.ProjectID)
			})
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Title", "Title").MarkRequired()
			b.Select("ProjectID", "Project", projectOptions(db)...)
			b.Text("Assignee", "Assignee").MarkRequired()
			b.Select("Priority", "Priority",
				form.Option{Value: "low", Label: "Low"},
				form.Option{Value: "medium", Label: "Medium"},
				form.Option{Value: "high", Label: "High"},
			)
			b.Select("Status", "Status",
				form.Option{Value: "open", Label: "Open"},
				form.Option{Value: "in_progress", Label: "In Progress"},
				form.Option{Value: "done", Label: "Done"},
			)
			b.Date("DueDate", "Due Date")
			b.Textarea("Description", "Description")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Title", "Title")
			b.Field("Project.Name", "Project")
			b.Field("Assignee", "Assignee")
			b.Field("Priority", "Priority")
			b.Field("Status", "Status")
			b.Field("DueDate", "Due Date")
			b.Field("Description", "Description")
			b.Field("CreatedAt", "Created At")
			b.Field("UpdatedAt", "Updated At")
		},
	})
}

func registerReports(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[models.ReportSnapshot](db)
	repo.SearchFields = []string{"Name", "Metric", "Dimension", "Description"}
	repo.FilterFields = []string{"Trend", "Metric", "Dimension"}
	repo.DefaultOrder = "snapshot_at desc, id desc"

	app.Register(goadmin.Resource{
		Name:           "reports",
		Path:           "reports",
		Title:          "Reports",
		Description:    "Read-only report snapshots demonstrating browse-only metric pages.",
		Permission:     "reports.view",
		EmptyText:      "No report snapshots yet.",
		CapabilityTags: []string{"readonly", "reporting"},
		Repository: repo,
		BuildGrid: func(b *grid.Builder) {
			b.DisableCreate = true
			b.DisableEdit = true
			b.DisableDelete = true
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name").SortableColumn()
			b.Column("Metric", "Metric")
			b.Column("Dimension", "Dimension")
			b.Column("Value", "Value").SortableColumn()
			b.Column("Trend", "Trend").Display(func(_ any, value any) template.HTML {
				text := fmt.Sprint(value)
				class := "badge badge-ok"
				if text == "down" {
					class = "badge badge-warn"
				}
				return template.HTML(fmt.Sprintf(`<span class="%s">%s</span>`, class, template.HTMLEscapeString(text)))
			})
			b.Column("SnapshotAt", "Snapshot At").SortableColumn()
			b.QuickSearch("Name", "Metric", "Dimension", "Description")
			b.Filter("Trend", "Trend", grid.FilterSelect).WithOptions(
				grid.Option{Value: "up", Label: "Up"},
				grid.Option{Value: "stable", Label: "Stable"},
				grid.Option{Value: "down", Label: "Down"},
			)
			b.Filter("Dimension", "Dimension", grid.FilterSelect).WithOptions(
				grid.Option{Value: "daily", Label: "Daily"},
				grid.Option{Value: "weekly", Label: "Weekly"},
				grid.Option{Value: "monthly", Label: "Monthly"},
			)
			b.PageAction("Open Audit Logs", "/admin/audits").WithStyle(grid.ActionGhost)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Metric", "Metric")
			b.Field("Dimension", "Dimension")
			b.Field("Value", "Value")
			b.Field("Trend", "Trend")
			b.Field("SnapshotAt", "Snapshot At")
			b.Field("Description", "Description")
			b.Field("CreatedAt", "Created At")
		},
	})
}

func categoryOptions(db *gorm.DB) []form.Option {
	var categories []models.Category
	_ = db.Order("sort asc, id asc").Find(&categories).Error
	options := make([]form.Option, 0, len(categories))
	for _, category := range categories {
		options = append(options, form.Option{
			Value: fmt.Sprintf("%d", category.ID),
			Label: category.Name,
		})
	}
	return options
}

func roleOptions(db *gorm.DB) []form.Option {
	var roles []auth.Role
	_ = db.Order("id asc").Find(&roles).Error
	options := make([]form.Option, 0, len(roles))
	for _, role := range roles {
		options = append(options, form.Option{
			Value: fmt.Sprintf("%d", role.ID),
			Label: role.Name,
		})
	}
	return options
}

func permissionIDOptions(db *gorm.DB) []form.Option {
	var permissions []auth.Permission
	_ = db.Order("id asc").Find(&permissions).Error
	options := make([]form.Option, 0, len(permissions))
	for _, permission := range permissions {
		options = append(options, form.Option{
			Value: fmt.Sprintf("%d", permission.ID),
			Label: permission.Name,
		})
	}
	return options
}

func permissionSlugOptions(db *gorm.DB) []form.Option {
	var permissions []auth.Permission
	_ = db.Order("id asc").Find(&permissions).Error
	options := make([]form.Option, 0, len(permissions))
	for _, permission := range permissions {
		options = append(options, form.Option{
			Value: permission.Slug,
			Label: permission.Name,
		})
	}
	return options
}

func menuOptions(db *gorm.DB) []form.Option {
	var menus []auth.Menu
	_ = db.Order("parent_id asc, `order` asc, id asc").Find(&menus).Error
	options := make([]form.Option, 0, len(menus))
	for _, menu := range menus {
		options = append(options, form.Option{
			Value: fmt.Sprintf("%d", menu.ID),
			Label: menu.Title,
		})
	}
	return options
}

func projectOptions(db *gorm.DB) []form.Option {
	var projects []models.Project
	_ = db.Order("id asc").Find(&projects).Error
	options := make([]form.Option, 0, len(projects))
	for _, project := range projects {
		options = append(options, form.Option{
			Value: fmt.Sprintf("%d", project.ID),
			Label: project.Name,
		})
	}
	return options
}

func formatDashboardTime(value time.Time) string {
	if value.IsZero() {
		return "Not scheduled"
	}
	return value.Format("2006-01-02 15:04")
}

func formatDashboardDate(value time.Time) string {
	if value.IsZero() {
		return "—"
	}
	return value.Format("2006-01-02")
}

func nonEmptyCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func faqCount(value any) int {
	switch typed := value.(type) {
	case []models.ArticleFAQ:
		return len(typed)
	case string:
		var items []map[string]string
		if strings.TrimSpace(typed) == "" {
			return 0
		}
		if err := json.Unmarshal([]byte(typed), &items); err != nil {
			return 0
		}
		return len(items)
	default:
		return 0
	}
}

func linkCount(value any) int {
	switch typed := value.(type) {
	case []models.ArticleLink:
		return len(typed)
	case string:
		return len(nonEmptyCSV(typed))
	default:
		return 0
	}
}

func renderAssetPreview(path, className string) template.HTML {
	path = strings.TrimSpace(path)
	if path == "" {
		return template.HTML("—")
	}
	escaped := template.HTMLEscapeString(path)
	if isDemoImagePath(path) {
		return template.HTML(fmt.Sprintf(`<a href="%s" target="_blank" rel="noreferrer"><img src="%s" alt="asset" class="%s"></a>`, escaped, escaped, className))
	}
	return template.HTML(fmt.Sprintf(`<a href="%s" target="_blank" rel="noreferrer">file</a>`, escaped))
}

func isDemoImagePath(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp":
		return true
	default:
		return false
	}
}

func renderGalleryList(value string) template.HTML {
	items := nonEmptyCSV(value)
	if len(items) == 0 {
		return template.HTML("—")
	}
	var buf strings.Builder
	buf.WriteString(`<div class="stack">`)
	for _, item := range items {
		escaped := template.HTMLEscapeString(item)
		buf.WriteString(fmt.Sprintf(`<a href="%s" target="_blank" rel="noreferrer">%s</a>`, escaped, escaped))
	}
	buf.WriteString(`</div>`)
	return template.HTML(buf.String())
}

func renderTagBadges(value string) template.HTML {
	tags := nonEmptyCSV(value)
	if len(tags) == 0 {
		return template.HTML("—")
	}
	var buf strings.Builder
	for _, tag := range tags {
		buf.WriteString(fmt.Sprintf(`<span class="badge badge-ok">%s</span> `, template.HTMLEscapeString(tag)))
	}
	return template.HTML(strings.TrimSpace(buf.String()))
}

func renderFAQList(value any) template.HTML {
	var items []map[string]string
	switch typed := value.(type) {
	case []models.ArticleFAQ:
		for _, faq := range typed {
			items = append(items, map[string]string{"Question": faq.Question, "Answer": faq.Answer})
		}
	case string:
		if strings.TrimSpace(typed) == "" {
			return template.HTML("—")
		}
		if err := json.Unmarshal([]byte(typed), &items); err != nil || len(items) == 0 {
			return template.HTML(template.HTMLEscapeString(typed))
		}
	default:
		return template.HTML("—")
	}
	if len(items) == 0 {
		return template.HTML("—")
	}
	var buf strings.Builder
	buf.WriteString(`<div class="stack">`)
	for _, item := range items {
		buf.WriteString(`<div class="detail-row">`)
		buf.WriteString(fmt.Sprintf(`<div><div class="detail-label">%s</div><div class="metric-hint">%s</div></div>`,
			template.HTMLEscapeString(item["Question"]),
			template.HTMLEscapeString(item["Answer"]),
		))
		buf.WriteString(`</div>`)
	}
	buf.WriteString(`</div>`)
	return template.HTML(buf.String())
}

func renderLinkList(value any) template.HTML {
	var items []models.ArticleLink
	switch typed := value.(type) {
	case []models.ArticleLink:
		items = typed
	default:
		return template.HTML("—")
	}
	if len(items) == 0 {
		return template.HTML("—")
	}
	var buf strings.Builder
	buf.WriteString(`<div class="stack">`)
	for _, link := range items {
		buf.WriteString(fmt.Sprintf(
			`<a href="%s" target="_blank" rel="noreferrer">%s</a>`,
			template.HTMLEscapeString(link.URL),
			template.HTMLEscapeString(link.Label),
		))
	}
	buf.WriteString(`</div>`)
	return template.HTML(buf.String())
}
