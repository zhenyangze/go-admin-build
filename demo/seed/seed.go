package seed

import (
	"time"

	"github.com/zhenyangze/go-admin-build/demo/models"
	"github.com/zhenyangze/go-admin-build/goadmin/auth"
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
		{Name: "Articles", Slug: "articles.manage"},
		{Name: "Categories", Slug: "categories.manage"},
	}
	for i := range permissions {
		if err := tx.Create(&permissions[i]).Error; err != nil {
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
	}
	contentChildren := []auth.Menu{
		{ParentID: menus[2].ID, Order: 10, Title: "Articles", URI: "articles", PermissionSlug: "articles.manage"},
		{ParentID: menus[2].ID, Order: 20, Title: "Categories", URI: "categories/tree", PermissionSlug: "categories.manage"},
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
		{Title: "Building Admin DSLs in Go", Summary: "How Grid/Form/Show map cleanly into typed builders.", Status: "published", CategoryID: categories[0].ID},
		{Title: "SQLite Demo Notes", Summary: "A lightweight verification target for the first framework slice.", Status: "draft", CategoryID: categories[len(categories)-1].ID},
	}
	for i := range articles {
		if err := tx.Create(&articles[i]).Error; err != nil {
			return err
		}
	}
	return nil
}
