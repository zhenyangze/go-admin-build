package seed

import (
	"time"

	"github.com/zhenyangze/go-admin-build/examples/demo/models"
	"github.com/zhenyangze/goadmin/auth"
	"gorm.io/gorm"
)

// Run ensures the demo database contains a usable baseline.
func Run(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedAuth(tx); err != nil {
			return err
		}
		if err := seedContent(tx); err != nil {
			return err
		}
		return nil
	})
}

func seedAuth(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&auth.User{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	permissions := []auth.Permission{
		{Name: "Users", Slug: "users.manage"},
		{Name: "Roles", Slug: "roles.manage"},
		{Name: "Permissions", Slug: "permissions.manage"},
		{Name: "Menus", Slug: "menus.manage"},
		{Name: "Articles", Slug: "articles.manage"},
		{Name: "Categories", Slug: "categories.manage"},
		{Name: "Projects", Slug: "projects.manage"},
		{Name: "Audit Logs", Slug: "audits.view"},
		{Name: "Tickets", Slug: "tickets.manage"},
		{Name: "Reports", Slug: "reports.view"},
	}
	for i := range permissions {
		if err := tx.Create(&permissions[i]).Error; err != nil {
			return err
		}
	}

	// 创建角色，包括 administrator 角色
	roles := []auth.Role{
		{Name: "超级管理员", Slug: auth.AdministratorRole},
		{Name: "管理员", Slug: "admin"},
	}
	for i := range roles {
		if err := tx.Create(&roles[i]).Error; err != nil {
			return err
		}
	}

	menus := []auth.Menu{
		{Title: "Dashboard", URI: "/"},
		{Title: "Security"},
		{Title: "Content"},
	}
	for i := range menus {
		if err := tx.Create(&menus[i]).Error; err != nil {
			return err
		}
	}

	securityChildren := []auth.Menu{
		{ParentID: menus[1].ID, Order: 10, Title: "Users", URI: "users", PermissionSlug: "users.manage"},
		{ParentID: menus[1].ID, Order: 20, Title: "Roles", URI: "roles", PermissionSlug: "roles.manage"},
		{ParentID: menus[1].ID, Order: 30, Title: "Permissions", URI: "permissions", PermissionSlug: "permissions.manage"},
		{ParentID: menus[1].ID, Order: 40, Title: "Menus", URI: "menus/tree", PermissionSlug: "menus.manage"},
	}
	contentChildren := []auth.Menu{
		{ParentID: menus[2].ID, Order: 10, Title: "Articles", URI: "articles", PermissionSlug: "articles.manage"},
		{ParentID: menus[2].ID, Order: 20, Title: "Categories", URI: "categories/tree", PermissionSlug: "categories.manage"},
		{ParentID: menus[2].ID, Order: 30, Title: "Projects", URI: "projects", PermissionSlug: "projects.manage"},
		{ParentID: menus[2].ID, Order: 40, Title: "Audit Logs", URI: "audits", PermissionSlug: "audits.view"},
		{ParentID: menus[2].ID, Order: 50, Title: "Tickets", URI: "tickets", PermissionSlug: "tickets.manage"},
		{ParentID: menus[2].ID, Order: 60, Title: "Reports", URI: "reports", PermissionSlug: "reports.view"},
	}
	for i := range securityChildren {
		if err := tx.Create(&securityChildren[i]).Error; err != nil {
			return err
		}
	}
	for i := range contentChildren {
		if err := tx.Create(&contentChildren[i]).Error; err != nil {
			return err
		}
	}

	role := auth.Role{
		Name: "Administrator",
		Slug: auth.AdministratorRole,
	}
	if err := tx.Create(&role).Error; err != nil {
		return err
	}
	if err := tx.Model(&role).Association("Permissions").Append(&permissions); err != nil {
		return err
	}
	allMenus := append(append(menus, securityChildren...), contentChildren...)
	if err := tx.Model(&role).Association("Menus").Append(&allMenus); err != nil {
		return err
	}

	password, err := auth.HashPassword("admin")
	if err != nil {
		return err
	}

	user := auth.User{
		Username:  "admin",
		Password:  password,
		Name:      "Administrator",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := tx.Create(&user).Error; err != nil {
		return err
	}
	return tx.Model(&user).Association("Roles").Append(&role)
}

func seedContent(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&models.Category{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		root := models.Category{Name: "Guides", Description: "Documentation and tutorials", Sort: 10}
		ops := models.Category{Name: "Operations", Description: "Deployment and maintenance", Sort: 20}
		if err := tx.Create(&root).Error; err != nil {
			return err
		}
		if err := tx.Create(&ops).Error; err != nil {
			return err
		}
		children := []models.Category{
			{ParentID: root.ID, Name: "Getting Started", Description: "Entry-level walkthroughs", Sort: 10},
			{ParentID: root.ID, Name: "Advanced", Description: "Deep-dive techniques", Sort: 20},
			{ParentID: ops.ID, Name: "Release", Description: "Release process notes", Sort: 10},
		}
		for i := range children {
			if err := tx.Create(&children[i]).Error; err != nil {
				return err
			}
		}
	}

	if err := tx.Model(&models.Article{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var categories []models.Category
	if err := tx.Order("id asc").Find(&categories).Error; err != nil {
		return err
	}
	if len(categories) == 0 {
		return nil
	}

	articles := []models.Article{
		{Title: "Building Admin DSLs in Go", Summary: "How Grid/Form/Show map cleanly into typed builders.", Status: "published", Featured: true, Tags: "go,admin,dsl", Image: "/admin/assets/admin.css", Gallery: "/admin/assets/admin.css,/admin/assets/admin.css", PublishedAt: time.Now().Add(-4 * time.Hour), VisibleFrom: time.Now().AddDate(0, 0, -7), VisibleTo: time.Now().AddDate(0, 0, 30), CategoryID: categories[0].ID},
		{Title: "SQLite Demo Notes", Summary: "A lightweight verification target for the first framework slice.", Status: "draft", Featured: false, Tags: "sqlite,demo", VisibleFrom: time.Now().AddDate(0, 0, -2), VisibleTo: time.Now().AddDate(0, 0, 14), CategoryID: categories[len(categories)-1].ID},
	}
	for i := range articles {
		if err := tx.Create(&articles[i]).Error; err != nil {
			return err
		}
	}
	faqs := []models.ArticleFAQ{
		{ArticleID: articles[0].ID, Sort: 1, Question: "Why Go?", Answer: "Static binaries and typed builders."},
		{ArticleID: articles[0].ID, Sort: 2, Question: "Why Dcat-style DSL?", Answer: "Fast admin screens with clear structure."},
	}
	if err := tx.Create(&faqs).Error; err != nil {
		return err
	}
	links := []models.ArticleLink{
		{ArticleID: articles[0].ID, Sort: 1, Label: "Repository", URL: "https://example.com/repo"},
		{ArticleID: articles[0].ID, Sort: 2, Label: "Docs", URL: "https://example.com/docs"},
	}
	if err := tx.Create(&links).Error; err != nil {
		return err
	}

	if err := tx.Model(&models.Project{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		projects := []models.Project{
			{Name: "Framework Launch", Owner: "Alice", Status: "active", Budget: 120000, StartsAt: time.Now().AddDate(0, -1, 0), EndsAt: time.Now().AddDate(0, 2, 0), Description: "Deliver the first public demo and adoption docs."},
			{Name: "Operations Dashboard", Owner: "Bob", Status: "planning", Budget: 80000, StartsAt: time.Now().AddDate(0, 0, 7), EndsAt: time.Now().AddDate(0, 3, 0), Description: "Build the second wave of admin operational tooling."},
		}
		for i := range projects {
			if err := tx.Create(&projects[i]).Error; err != nil {
				return err
			}
		}
		milestones := []models.ProjectMilestone{
			{ProjectID: projects[0].ID, Sort: 1, Title: "Core modules done", Deadline: time.Now().AddDate(0, 0, 14)},
			{ProjectID: projects[0].ID, Sort: 2, Title: "Demo smoke stable", Deadline: time.Now().AddDate(0, 0, 30)},
			{ProjectID: projects[1].ID, Sort: 1, Title: "PRD approved", Deadline: time.Now().AddDate(0, 0, 10)},
		}
		if err := tx.Create(&milestones).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&models.Ticket{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		var projects []models.Project
		if err := tx.Order("id asc").Find(&projects).Error; err != nil {
			return err
		}
		if len(projects) > 0 {
			tickets := []models.Ticket{
				{Title: "Polish dashboard spacing", ProjectID: projects[0].ID, Assignee: "Alice", Priority: "high", Status: "open", Description: "Tighten the top-level dashboard layout and labels.", DueDate: time.Now().AddDate(0, 0, 7)},
				{Title: "Add smoke coverage for tickets", ProjectID: projects[0].ID, Assignee: "Bob", Priority: "medium", Status: "in_progress", Description: "Extend smoke script and integration coverage for the ticket module.", DueDate: time.Now().AddDate(0, 0, 14)},
			}
			if err := tx.Create(&tickets).Error; err != nil {
				return err
			}
		}
	}
	if err := tx.Model(&models.ReportSnapshot{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		reports := []models.ReportSnapshot{
			{Name: "Weekly Signups", Metric: "signups", Dimension: "weekly", Value: 428, Trend: "up", SnapshotAt: time.Now().Add(-24 * time.Hour), Description: "Weekly user signup total compared with the previous period."},
			{Name: "Project Throughput", Metric: "deliveries", Dimension: "monthly", Value: 18, Trend: "stable", SnapshotAt: time.Now().Add(-48 * time.Hour), Description: "Completed project milestones within the current month."},
			{Name: "Support Backlog", Metric: "tickets", Dimension: "daily", Value: 11, Trend: "down", SnapshotAt: time.Now().Add(-2 * time.Hour), Description: "Open ticket backlog at the last reporting checkpoint."},
		}
		if err := tx.Create(&reports).Error; err != nil {
			return err
		}
	}
	if err := tx.Model(&models.AuditLog{}).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		logs := []models.AuditLog{
			{Actor: "admin", Action: "create", Resource: "article", ResourceID: "1", Level: "info", IP: "127.0.0.1", Detail: "Created the first seeded article.", CreatedAt: time.Now().Add(-6 * time.Hour)},
			{Actor: "admin", Action: "update", Resource: "project", ResourceID: "1", Level: "warning", IP: "127.0.0.1", Detail: "Adjusted project delivery dates after review.", CreatedAt: time.Now().Add(-3 * time.Hour)},
			{Actor: "system", Action: "smoke", Resource: "demo", ResourceID: "n/a", Level: "info", IP: "127.0.0.1", Detail: "Automated smoke validation completed successfully.", CreatedAt: time.Now().Add(-30 * time.Minute)},
		}
		if err := tx.Create(&logs).Error; err != nil {
			return err
		}
	}
	return nil
}
