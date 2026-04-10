package demo

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/zhenyangze/go-admin-build/demo/models"
	demoseed "github.com/zhenyangze/go-admin-build/demo/seed"
	"github.com/zhenyangze/go-admin-build/goadmin"
	"github.com/zhenyangze/go-admin-build/goadmin/auth"
	"github.com/zhenyangze/go-admin-build/goadmin/form"
	"github.com/zhenyangze/go-admin-build/goadmin/grid"
	"github.com/zhenyangze/go-admin-build/goadmin/show"
	"github.com/zhenyangze/go-admin-build/goadmin/store/gormstore"
	"github.com/zhenyangze/go-admin-build/goadmin/tree"
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
	app, err := BuildWithDB(db)
	if err != nil {
		return nil, nil, err
	}
	return app, db, nil
}

// BuildWithDB wires the reusable admin package to the demo models.
func BuildWithDB(db *gorm.DB) (*goadmin.App, error) {
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
	}, auth.NewService(db))
	if err != nil {
		return nil, err
	}

	app.SetDashboard(func(ctx context.Context) (goadmin.DashboardData, error) {
		var users int64
		var roles int64
		var articles int64
		var categories int64
		if err := db.WithContext(ctx).Model(&auth.User{}).Count(&users).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&auth.Role{}).Count(&roles).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&models.Article{}).Count(&articles).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		if err := db.WithContext(ctx).Model(&models.Category{}).Count(&categories).Error; err != nil {
			return goadmin.DashboardData{}, err
		}
		return goadmin.DashboardData{
			Title:       "Dashboard",
			Description: "Dcat-inspired modules running on Go + GORM + TailwindCSS.",
			Cards: []goadmin.DashboardCard{
				{Title: "Admin users", Value: fmt.Sprintf("%d", users), Hint: "Built-in auth model"},
				{Title: "Roles", Value: fmt.Sprintf("%d", roles), Hint: "RBAC seed data"},
				{Title: "Articles", Value: fmt.Sprintf("%d", articles), Hint: "Grid/Form/Show demo"},
				{Title: "Categories", Value: fmt.Sprintf("%d", categories), Hint: "Tree demo"},
			},
		}, nil
	})

	registerUsers(app, db)
	registerRoles(app, db)
	registerArticles(app, db)
	registerCategories(app, db)

	return app, nil
}

func migrate(db *gorm.DB) error {
	if err := auth.AutoMigrate(db); err != nil {
		return err
	}
	return db.AutoMigrate(&models.Category{}, &models.Article{})
}

func seed(db *gorm.DB) error {
	return demoseed.Run(db)
}

func registerUsers(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[auth.User](db)
	repo.SearchFields = []string{"Username", "Name"}
	repo.FilterFields = []string{"Username"}
	repo.DefaultOrder = "id desc"
	repo.Mutators["Password"] = func(raw string) (any, error) {
		return auth.HashPassword(raw)
	}

	app.Register(goadmin.Resource{
		Name:        "users",
		Path:        "users",
		Title:       "Users",
		Description: "Built-in admin accounts",
		Permission:  "users.manage",
		Repository:  repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Username", "Username").SortableColumn()
			b.Column("Name", "Name")
			b.Column("CreatedAt", "Created")
			b.QuickSearch("Username", "Name")
			b.Filter("Username", "Username", grid.FilterText)
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Username", "Username").MarkRequired()
			b.Text("Name", "Display Name").MarkRequired()
			b.Password("Password", "Password").WithHelp("Leave blank on edit to keep the current password.")
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Username", "Username")
			b.Field("Name", "Display Name")
			b.Divider("Meta")
			b.Field("CreatedAt", "Created At")
			b.Field("UpdatedAt", "Updated At")
		},
	})
}

func registerRoles(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[auth.Role](db)
	repo.SearchFields = []string{"Name", "Slug"}
	repo.DefaultOrder = "id asc"

	app.Register(goadmin.Resource{
		Name:        "roles",
		Path:        "roles",
		Title:       "Roles",
		Description: "RBAC roles",
		Permission:  "roles.manage",
		Repository:  repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name")
			b.Column("Slug", "Slug").SortableColumn()
			b.QuickSearch("Name", "Slug")
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Name", "Name").MarkRequired()
			b.Text("Slug", "Slug").MarkRequired()
			b.HideDelete = true
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Name", "Name")
			b.Field("Slug", "Slug")
			b.Field("CreatedAt", "Created At")
		},
	})
}

func registerArticles(app *goadmin.App, db *gorm.DB) {
	repo := gormstore.New[models.Article](db)
	repo.SearchFields = []string{"Title", "Summary"}
	repo.FilterFields = []string{"Status", "CategoryID"}
	repo.Preloads = []string{"Category"}
	repo.DefaultOrder = "id desc"

	app.Register(goadmin.Resource{
		Name:        "articles",
		Path:        "articles",
		Title:       "Articles",
		Description: "Grid / Form / Show demo with GORM relation loading.",
		Permission:  "articles.manage",
		Repository:  repo,
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
			b.Column("Category.Name", "Category")
			b.Column("CreatedAt", "Created")
			b.QuickSearch("Title", "Summary")
			b.Filter("Status", "Status", grid.FilterSelect).WithOptions(
				grid.Option{Value: "draft", Label: "Draft"},
				grid.Option{Value: "published", Label: "Published"},
			)
		},
		BuildForm: func(b *form.Builder) {
			b.Display("ID", "ID")
			b.Text("Title", "Title").MarkRequired()
			b.Textarea("Summary", "Summary").MarkRequired()
			b.Select("Status", "Status",
				form.Option{Value: "draft", Label: "Draft"},
				form.Option{Value: "published", Label: "Published"},
			)
			b.Select("CategoryID", "Category", categoryOptions(db)...)
		},
		BuildShow: func(b *show.Builder) {
			b.Field("ID", "ID")
			b.Field("Title", "Title")
			b.Field("Summary", "Summary")
			b.Divider("Publishing")
			b.Field("Status", "Status")
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
		Name:        "categories",
		Path:        "categories",
		Title:       "Categories",
		Description: "Tree and flat management demo.",
		Permission:  "categories.manage",
		Repository:  repo,
		BuildGrid: func(b *grid.Builder) {
			b.Column("ID", "ID").SortableColumn()
			b.Column("Name", "Name")
			b.Column("ParentID", "Parent ID")
			b.Column("Sort", "Sort").SortableColumn()
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
